package api

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewClient(port string) *Client {
	return &Client{
		BaseURL: fmt.Sprintf("http://127.0.0.1:%s", port),
		HTTPClient: &http.Client{
			Timeout: 0, // Streaming responses require no overall timeout
		},
	}
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Messages    []ChatMessage `json:"messages"`
	Stream      bool          `json:"stream"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
}

type ChatChunkDelta struct {
	Content          string `json:"content"`
	ReasoningContent string `json:"reasoning_content"`
}

type ChatChunkChoice struct {
	Delta        ChatChunkDelta `json:"delta"`
	FinishReason *string        `json:"finish_reason"`
	Index        int            `json:"index"`
}

type ChatChunk struct {
	Choices []ChatChunkChoice `json:"choices"`
}

type StreamEvent struct {
	Token     string
	IsThink   bool
	Done      bool
	Error     error
	RawChunk  string
}

func (c *Client) StreamChat(ctx context.Context, messages []ChatMessage, temp float64, maxTokens int, out chan<- StreamEvent) {
	defer close(out)

	reqBody := ChatRequest{
		Messages:    messages,
		Stream:      true,
		Temperature: temp,
		MaxTokens:   maxTokens,
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		out <- StreamEvent{Error: fmt.Errorf("marshal error: %w", err)}
		return
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/v1/chat/completions", bytes.NewReader(payload))
	if err != nil {
		out <- StreamEvent{Error: fmt.Errorf("request creation error: %w", err)}
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		out <- StreamEvent{Error: fmt.Errorf("connection error: %w", err)}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		out <- StreamEvent{Error: fmt.Errorf("server returned HTTP %d", resp.StatusCode)}
		return
	}

	scanner := bufio.NewScanner(resp.Body)
	// Larger buffer for big chunks
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	inThinkTag := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				out <- StreamEvent{Done: true}
				return
			}

			var chunk ChatChunk
			if err := json.Unmarshal([]byte(data), &chunk); err == nil {
				if len(chunk.Choices) > 0 {
					delta := chunk.Choices[0].Delta
					token := delta.Content
					if token == "" {
						token = delta.ReasoningContent
					}

					if strings.Contains(token, "<think>") {
						inThinkTag = true
					}
					if strings.Contains(token, "</think>") {
						inThinkTag = false
					}

					out <- StreamEvent{
						Token:    token,
						IsThink:  inThinkTag,
						RawChunk: data,
					}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		out <- StreamEvent{Error: fmt.Errorf("stream read error: %w", err)}
	}
}

type ServerStatus struct {
	Status              string             `json:"status"`
	TunnelURL           string             `json:"tunnel_url"`
	KnowledgeBase       KnowledgeBaseStats `json:"knowledge_base"`
	Temperature         float64            `json:"temperature"`
	TopK                int                `json:"top_k"`
	SimilarityThreshold float64            `json:"similarity_threshold"`
}

type KnowledgeBaseStats struct {
	Documents int `json:"documents"`
	Chunks    int `json:"chunks"`
}

func (c *Client) GetStatus() (*ServerStatus, error) {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(c.BaseURL + "/status")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var st ServerStatus
	if err := json.NewDecoder(resp.Body).Decode(&st); err != nil {
		return nil, err
	}
	return &st, nil
}

func (c *Client) CheckHealth() bool {
	client := &http.Client{Timeout: 1 * time.Second}
	resp, err := client.Get(c.BaseURL + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

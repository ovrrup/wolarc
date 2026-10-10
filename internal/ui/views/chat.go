package views

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ovrrup/wolarc/internal/api"
	"github.com/ovrrup/wolarc/internal/config"
	"github.com/ovrrup/wolarc/internal/ui/styles"
)

type ChatMessageItem struct {
	Role       string
	Content    string
	Thinking   string
	Tokens     int
	TokPerSec  float64
	ElapsedSec float64
}

type TokenMsg api.StreamEvent

type ChatModel struct {
	Config        *config.Config
	Client        *api.Client
	Viewport      viewport.Model
	TextArea      textarea.Model
	Messages      []ChatMessageItem
	Streaming     bool
	StreamStart   time.Time
	TokensEmitted int
	Spinner       spinner.Model
	FoldThinking  bool
	CurrentResp   strings.Builder
	CurrentThnk   strings.Builder
	CancelFunc    context.CancelFunc
	Width         int
	Height        int
}

func NewChatModel(cfg *config.Config) ChatModel {
	ta := textarea.New()
	ta.Placeholder = "Ask anything (AI intelligence enhanced with your personal knowledge & files)..."
	ta.Focus()
	ta.Prompt = "┃ "
	ta.CharLimit = 2048
	ta.SetWidth(78)
	ta.SetHeight(3)
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.ShowLineNumbers = false

	vp := viewport.New(78, 16)
	vp.SetContent(styles.DescStyle.Render("Welcome to Wolarc Chat! Ask any question or explore your personal knowledge base.\n"))

	sp := spinner.New()
	sp.Spinner = spinner.Points
	sp.Style = lipgloss.NewStyle().Foreground(styles.ColorCyan)

	return ChatModel{
		Config:       cfg,
		Client:       api.NewClient(cfg.AgentPort),
		Viewport:     vp,
		TextArea:     ta,
		Messages:     []ChatMessageItem{},
		Spinner:      sp,
		FoldThinking: false,
		Width:        80,
		Height:       24,
	}
}

func (m *ChatModel) SetSize(width, height int) {
	m.Width = width
	m.Height = height
	contentWidth := max(20, width-4)
	m.TextArea.SetWidth(contentWidth)

	vpHeight := max(6, height-m.TextArea.Height()-8)
	m.Viewport.Width = contentWidth
	m.Viewport.Height = vpHeight
	m.renderConversation()
}

func (m *ChatModel) SendCurrentMessage() tea.Cmd {
	text := strings.TrimSpace(m.TextArea.Value())
	if text == "" || m.Streaming {
		return nil
	}

	m.TextArea.Reset()
	m.Messages = append(m.Messages, ChatMessageItem{
		Role:    "user",
		Content: text,
	})
	m.Streaming = true
	m.StreamStart = time.Now()
	m.TokensEmitted = 0
	m.CurrentResp.Reset()
	m.CurrentThnk.Reset()
	m.renderConversation()

	// Convert to API messages
	var apiMsgs []api.ChatMessage
	for _, msg := range m.Messages {
		apiMsgs = append(apiMsgs, api.ChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.CancelFunc = cancel

	outChan := make(chan api.StreamEvent, 64)
	go m.Client.StreamChat(ctx, apiMsgs, m.Config.Temperature, 2048, outChan)

	return tea.Batch(waitForToken(outChan), m.Spinner.Tick)
}

func (m ChatModel) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, m.Spinner.Tick)
}

func waitForToken(c <-chan api.StreamEvent) tea.Cmd {
	return func() tea.Msg {
		event, ok := <-c
		if !ok {
			return TokenMsg{Done: true}
		}
		return TokenMsg(event)
	}
}

func (m *ChatModel) Update(msg tea.Msg) (ChatModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case spinner.TickMsg:
		var spCmd tea.Cmd
		m.Spinner, spCmd = m.Spinner.Update(msg)
		if m.Streaming {
			m.renderConversation()
		}
		return *m, spCmd

	case TokenMsg:
		if msg.Done || msg.Error != nil {
			m.Streaming = false
			elapsed := time.Since(m.StreamStart).Seconds()
			tokPerSec := 0.0
			if elapsed > 0 && m.TokensEmitted > 0 {
				tokPerSec = float64(m.TokensEmitted) / elapsed
			}

			assistantMsg := ChatMessageItem{
				Role:       "assistant",
				Content:    m.CurrentResp.String(),
				Thinking:   m.CurrentThnk.String(),
				Tokens:     m.TokensEmitted,
				ElapsedSec: elapsed,
				TokPerSec:  tokPerSec,
			}
			if msg.Error != nil {
				assistantMsg.Content += fmt.Sprintf("\n[Error: %v]", msg.Error)
			}
			m.Messages = append(m.Messages, assistantMsg)
			m.CurrentResp.Reset()
			m.CurrentThnk.Reset()
			m.renderConversation()
			return *m, nil
		}

		m.TokensEmitted++
		if msg.IsThink {
			m.CurrentThnk.WriteString(msg.Token)
		} else {
			m.CurrentResp.WriteString(msg.Token)
		}
		m.renderConversation()
		return *m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if !msg.Alt {
				cmd := m.SendCurrentMessage()
				return *m, cmd
			}
		case tea.KeyCtrlC:
			if m.Streaming && m.CancelFunc != nil {
				m.CancelFunc()
				m.Streaming = false
				m.renderConversation()
				return *m, nil
			}
		case tea.KeyCtrlL:
			m.Messages = nil
			m.renderConversation()
			return *m, nil
		}

		if msg.String() == "ctrl+t" {
			m.FoldThinking = !m.FoldThinking
			m.renderConversation()
			return *m, nil
		}

		if msg.String() == "ctrl+y" {
			// Copy last assistant message
			for i := len(m.Messages) - 1; i >= 0; i-- {
				if m.Messages[i].Role == "assistant" {
					_ = clipboard.WriteAll(m.Messages[i].Content)
					break
				}
			}
			return *m, nil
		}
	}

	var taCmd tea.Cmd
	m.TextArea, taCmd = m.TextArea.Update(msg)
	cmds = append(cmds, taCmd)

	var vpCmd tea.Cmd
	m.Viewport, vpCmd = m.Viewport.Update(msg)
	cmds = append(cmds, vpCmd)

	return *m, tea.Batch(cmds...)
}

func (m *ChatModel) renderConversation() {
	var sb strings.Builder

	for _, msg := range m.Messages {
		if msg.Role == "user" {
			sb.WriteString(styles.UserMessageStyle.Render("You: ") + msg.Content + "\n\n")
		} else {
			if msg.Thinking != "" {
				if m.FoldThinking {
					foldBadge := styles.DescStyle.Render("▸ [Reasoning Process folded - press Ctrl+T to view]")
					sb.WriteString(foldBadge + "\n\n")
				} else {
					thnkBox := styles.ThinkingBoxStyle.Width(m.Viewport.Width - 4).Render("Thinking Process:\n" + msg.Thinking)
					sb.WriteString(thnkBox + "\n")
				}
			}

			// Assistant header with metrics badge if available
			var metaBadge string
			if msg.TokPerSec > 0 {
				metaBadge = " " + lipgloss.NewStyle().Foreground(styles.ColorCyan).Render(fmt.Sprintf("[⚡ %.1f tok/s • %d tok]", msg.TokPerSec, msg.Tokens))
			}
			sb.WriteString(styles.AssistantMessageStyle.Render("Wola:") + metaBadge + "\n" + msg.Content + "\n\n")
		}
	}

	// Live streaming chunk
	if m.Streaming {
		if m.CurrentThnk.Len() > 0 {
			thnkBox := styles.ThinkingBoxStyle.Width(m.Viewport.Width - 4).Render("Thinking Process:\n" + m.CurrentThnk.String())
			sb.WriteString(thnkBox + "\n")
		}
		if m.CurrentResp.Len() > 0 {
			sb.WriteString(styles.AssistantMessageStyle.Render("Wola: ") + m.CurrentResp.String() + " ▌\n\n")
		} else if m.CurrentThnk.Len() == 0 {
			sb.WriteString(styles.AssistantMessageStyle.Render("Wola: ") + m.Spinner.View() + " " + styles.DescStyle.Render("Thinking...") + "\n\n")
		}
	}

	m.Viewport.SetContent(sb.String())
	m.Viewport.GotoBottom()
}

func (m ChatModel) View() string {
	vpBox := styles.CardStyle.Width(m.Width - 2).Height(m.Viewport.Height + 2).Render(m.Viewport.View())

	charCount := len(m.TextArea.Value())
	inputLabel := styles.DescStyle.Render(fmt.Sprintf(" Prompt (%d chars • ~%d tokens)", charCount, charCount/4))
	inputBox := styles.CardStyle.Width(m.Width - 2).Render(inputLabel + "\n" + m.TextArea.View())

	hints := fmt.Sprintf(
		" %s Send  •  %s Clear  •  %s Copy Answer  •  %s Fold Reasoning  •  %s Abort",
		styles.KeyStyle.Render("Enter"),
		styles.KeyStyle.Render("Ctrl+L"),
		styles.KeyStyle.Render("Ctrl+Y"),
		styles.KeyStyle.Render("Ctrl+T"),
		styles.KeyStyle.Render("Ctrl+C"),
	)
	hintsBar := styles.DescStyle.Render(hints)

	return lipgloss.JoinVertical(lipgloss.Left, vpBox, inputBox, hintsBar)
}

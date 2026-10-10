package views

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ovrrup/wolarc/internal/api"
	"github.com/ovrrup/wolarc/internal/config"
	"github.com/ovrrup/wolarc/internal/ui/styles"
)

type ChatMessageItem struct {
	Role     string
	Content  string
	Thinking string
}

type TokenMsg api.StreamEvent

type ChatModel struct {
	Config      *config.Config
	Client      *api.Client
	Viewport    viewport.Model
	TextArea    textarea.Model
	Messages    []ChatMessageItem
	Streaming   bool
	CurrentResp strings.Builder
	CurrentThnk strings.Builder
	CancelFunc  context.CancelFunc
	Width       int
	Height      int
}

func NewChatModel(cfg *config.Config) ChatModel {
	ta := textarea.New()
	ta.Placeholder = "Ask a question (RAG context will be injected automatically)..."
	ta.Focus()
	ta.Prompt = "┃ "
	ta.CharLimit = 2048
	ta.SetWidth(78)
	ta.SetHeight(3)
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.ShowLineNumbers = false

	vp := viewport.New(78, 16)
	vp.SetContent(styles.DescStyle.Render("Welcome to Wolarc Chat! Ask a question or discuss your indexed notes.\n"))

	return ChatModel{
		Config:   cfg,
		Client:   api.NewClient(cfg.AgentPort),
		Viewport: vp,
		TextArea: ta,
		Messages: []ChatMessageItem{},
		Width:    80,
		Height:   24,
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

	outChan := make(chan api.StreamEvent, 32)
	go m.Client.StreamChat(ctx, apiMsgs, m.Config.Temperature, 2048, outChan)

	return waitForToken(outChan)
}

func (m ChatModel) Init() tea.Cmd {
	return textarea.Blink
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
	case TokenMsg:
		if msg.Done || msg.Error != nil {
			m.Streaming = false
			assistantMsg := ChatMessageItem{
				Role:     "assistant",
				Content:  m.CurrentResp.String(),
				Thinking: m.CurrentThnk.String(),
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
				thnkBox := styles.ThinkingBoxStyle.Width(m.Viewport.Width - 4).Render("Thinking Process:\n" + msg.Thinking)
				sb.WriteString(thnkBox + "\n")
			}
			sb.WriteString(styles.AssistantMessageStyle.Render("Wola: ") + msg.Content + "\n\n")
		}
	}

	// Live streaming chunk
	if m.Streaming {
		if m.CurrentThnk.Len() > 0 {
			thnkBox := styles.ThinkingBoxStyle.Width(m.Viewport.Width - 4).Render("Thinking Process:\n" + m.CurrentThnk.String())
			sb.WriteString(thnkBox + "\n")
		}
		if m.CurrentResp.Len() > 0 {
			sb.WriteString(styles.AssistantMessageStyle.Render("Wola: ") + m.CurrentResp.String() + " █\n\n")
		} else if m.CurrentThnk.Len() == 0 {
			sb.WriteString(styles.AssistantMessageStyle.Render("Wola: ") + styles.DescStyle.Render("Thinking...") + "\n\n")
		}
	}

	m.Viewport.SetContent(sb.String())
	m.Viewport.GotoBottom()
}

func (m ChatModel) View() string {
	vpBox := styles.CardStyle.Width(m.Width - 2).Height(m.Viewport.Height + 2).Render(m.Viewport.View())
	inputBox := styles.CardStyle.Width(m.Width - 2).Render(m.TextArea.View())

	hints := fmt.Sprintf(
		" %s Send  •  %s Clear History  •  %s Abort Stream  •  %s Switch Tab",
		styles.KeyStyle.Render("Enter"),
		styles.KeyStyle.Render("Ctrl+L"),
		styles.KeyStyle.Render("Ctrl+C"),
		styles.KeyStyle.Render("Tab"),
	)
	hintsBar := styles.DescStyle.Render(hints)

	return lipgloss.JoinVertical(lipgloss.Left, vpBox, inputBox, hintsBar)
}

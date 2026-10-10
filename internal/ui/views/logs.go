package views

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ovrrup/wolarc/internal/config"
	"github.com/ovrrup/wolarc/internal/ui/styles"
)

type LogStream struct {
	Name     string
	Filename string
	Shortcut string
	Color    lipgloss.TerminalColor
}

type LogsModel struct {
	Config       *config.Config
	Streams      []LogStream
	ActiveStream int
	Viewport     viewport.Model
	AutoFollow   bool
	StatusMsg    string
	Width        int
	Height       int
}

func NewLogsModel(cfg *config.Config) LogsModel {
	homeDir, _ := os.UserHomeDir()

	streams := []LogStream{
		{Name: "Chat LLM", Filename: filepath.Join(homeDir, "chat.log"), Shortcut: "1", Color: styles.ColorPrimary},
		{Name: "Embed LLM", Filename: filepath.Join(homeDir, "embed.log"), Shortcut: "2", Color: styles.ColorSuccess},
		{Name: "Agent API", Filename: filepath.Join(homeDir, "agent.log"), Shortcut: "3", Color: styles.ColorAccent},
		{Name: "SSH Tunnel", Filename: filepath.Join(homeDir, "tunnel.log"), Shortcut: "4", Color: styles.ColorWarning},
		{Name: "Indexer", Filename: filepath.Join(homeDir, "indexer.log"), Shortcut: "5", Color: styles.ColorPrimary},
		{Name: "Thinking", Filename: filepath.Join(homeDir, ".wola_thinking.log"), Shortcut: "6", Color: styles.ColorDim},
	}

	vp := viewport.New(78, 16)
	vp.SetContent(styles.DescStyle.Render("Initializing log tailer...\n"))

	m := LogsModel{
		Config:       cfg,
		Streams:      streams,
		ActiveStream: 0,
		Viewport:     vp,
		AutoFollow:   true,
		Width:        80,
		Height:       24,
	}
	m.Refresh()
	return m
}

func (m *LogsModel) Refresh() {
	if m.ActiveStream < 0 || m.ActiveStream >= len(m.Streams) {
		return
	}
	st := m.Streams[m.ActiveStream]

	content, size, lineCount := readLastLines(st.Filename, 500)
	if content == "" {
		m.Viewport.SetContent(styles.DescStyle.Render(fmt.Sprintf("(Log file is empty or does not exist yet: %s)", st.Filename)))
	} else {
		m.Viewport.SetContent(content)
	}

	if m.AutoFollow {
		m.Viewport.GotoBottom()
	}

	sizeStr := fmt.Sprintf("%.1f KB", float64(size)/1024.0)
	if size > 1024*1024 {
		sizeStr = fmt.Sprintf("%.2f MB", float64(size)/(1024.0*1024.0))
	}

	followBadge := "○ PAUSED"
	if m.AutoFollow {
		followBadge = "● TAILING"
	}

	m.StatusMsg = fmt.Sprintf("[%s] %s  •  %d lines  •  %s  •  %s",
		st.Name,
		filepath.Base(st.Filename),
		lineCount,
		sizeStr,
		followBadge,
	)
}

func readLastLines(path string, maxLines int) (string, int64, int) {
	fi, err := os.Stat(path)
	if err != nil {
		return "", 0, 0
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return "", fi.Size(), 0
	}

	lines := strings.Split(string(raw), "\n")
	total := len(lines)
	if total > 0 && lines[total-1] == "" {
		lines = lines[:total-1]
		total--
	}

	start := 0
	if total > maxLines {
		start = total - maxLines
	}

	return strings.Join(lines[start:], "\n"), fi.Size(), total
}

func (m *LogsModel) ClearCurrentLog() {
	if m.ActiveStream < 0 || m.ActiveStream >= len(m.Streams) {
		return
	}
	path := m.Streams[m.ActiveStream].Filename
	_ = os.WriteFile(path, []byte{}, 0644)
	m.Refresh()
	m.StatusMsg = fmt.Sprintf("Cleared log file: %s", filepath.Base(path))
}

func (m LogsModel) Init() tea.Cmd {
	return nil
}

func (m LogsModel) Update(msg tea.Msg) (LogsModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "1", "2", "3", "4", "5", "6":
			idx := int(msg.String()[0] - '1')
			if idx >= 0 && idx < len(m.Streams) {
				m.ActiveStream = idx
				m.Refresh()
			}
			return m, nil
		case "h", "left":
			if m.ActiveStream > 0 {
				m.ActiveStream--
				m.Refresh()
			}
			return m, nil
		case "l", "right":
			if m.ActiveStream < len(m.Streams)-1 {
				m.ActiveStream++
				m.Refresh()
			}
			return m, nil
		case "f":
			m.AutoFollow = !m.AutoFollow
			if m.AutoFollow {
				m.Viewport.GotoBottom()
			}
			m.Refresh()
			return m, nil
		case "c":
			m.ClearCurrentLog()
			return m, nil
		case "r":
			m.Refresh()
			return m, nil
		case "g":
			m.AutoFollow = false
			m.Viewport.GotoTop()
			return m, nil
		case "G":
			m.Viewport.GotoBottom()
			return m, nil
		}
	}

	m.Viewport, cmd = m.Viewport.Update(msg)
	return m, cmd
}

func (m LogsModel) View() string {
	// 1. STREAM SELECTOR BAR
	var tabItems []string
	for idx, st := range m.Streams {
		label := fmt.Sprintf("[%s] %s", st.Shortcut, st.Name)
		if idx == m.ActiveStream {
			tabItems = append(tabItems, lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#0f172a")).
				Background(st.Color).
				Padding(0, 1).
				Render(label))
		} else {
			tabItems = append(tabItems, lipgloss.NewStyle().
				Foreground(styles.ColorDim).
				Padding(0, 1).
				Render(label))
		}
	}
	streamBar := lipgloss.JoinHorizontal(lipgloss.Top, tabItems...)

	// 2. VIEWPORT CARD
	vpBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorBorder).
		Padding(0, 1).
		Width(m.Width - 4).
		Render(m.Viewport.View())

	// 3. STATUS & ACTION FOOTER
	statusBar := styles.DescStyle.Render(m.StatusMsg)

	actions := []string{
		styles.KeyStyle.Render("1-6") + ": Switch Stream",
		styles.KeyStyle.Render("f") + ": Toggle Follow",
		styles.KeyStyle.Render("c") + ": Clear Log",
		styles.KeyStyle.Render("r") + ": Reload",
		styles.KeyStyle.Render("↑/↓/PgUp/PgDn") + ": Scroll",
		styles.KeyStyle.Render("g/G") + ": Top/Bottom",
	}
	actionBar := styles.ActionStyle.Render(strings.Join(actions, "  │  "))

	return lipgloss.JoinVertical(lipgloss.Left, streamBar, vpBox, statusBar, actionBar)
}

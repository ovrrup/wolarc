package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ovrrup/wolarc/internal/config"
	"github.com/ovrrup/wolarc/internal/system"
	"github.com/ovrrup/wolarc/internal/ui/components"
	"github.com/ovrrup/wolarc/internal/ui/styles"
	"github.com/ovrrup/wolarc/internal/ui/views"
)

const (
	TabDashboard = 0
	TabChat      = 1
	TabModels    = 2
	TabKnowledge = 3
	TabLogs      = 4
)

type TickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

type AppModel struct {
	Config    *config.Config
	ActiveTab int
	Dashboard views.DashboardModel
	Chat      views.ChatModel
	Models    views.ModelsModel
	Knowledge views.KnowledgeModel
	Logs      views.LogsModel
	Width     int
	Height    int
	StartTime time.Time
	IsOnline  bool
	TunnelURL string
	Quitting  bool
}

func NewAppModel(cfg *config.Config, initialTab int) AppModel {
	if initialTab < TabDashboard || initialTab > TabLogs {
		initialTab = TabDashboard
	}

	services := system.CheckAllServices(cfg.ChatPort, cfg.EmbedPort)
	online := services.ChatServer.Running || services.AgentGateway.Running

	return AppModel{
		Config:    cfg,
		ActiveTab: initialTab,
		Dashboard: views.NewDashboardModel(cfg),
		Chat:      views.NewChatModel(cfg),
		Models:    views.NewModelsModel(cfg),
		Knowledge: views.NewKnowledgeModel(cfg),
		Logs:      views.NewLogsModel(cfg),
		Width:     80,
		Height:    24,
		StartTime: time.Now(),
		IsOnline:  online,
		TunnelURL: config.GetTunnelURL(cfg),
	}
}

func (m AppModel) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		tickCmd(),
		m.Chat.Init(),
		m.Models.Init(),
		m.Knowledge.Init(),
		m.Logs.Init(),
	)
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

		m.Dashboard.Width = msg.Width
		m.Dashboard.Height = msg.Height

		m.Chat.SetSize(msg.Width, msg.Height)

		m.Models.Width = msg.Width
		m.Models.Height = msg.Height

		m.Knowledge.Width = msg.Width
		m.Knowledge.Height = msg.Height

		m.Logs.Width = msg.Width
		m.Logs.Height = msg.Height
		m.Logs.Viewport.Width = max(20, msg.Width-6)
		m.Logs.Viewport.Height = max(6, msg.Height-12)

		return m, nil

	case TickMsg:
		services := system.CheckAllServices(m.Config.ChatPort, m.Config.EmbedPort)
		m.IsOnline = services.ChatServer.Running || services.AgentGateway.Running
		m.TunnelURL = config.GetTunnelURL(m.Config)

		if m.ActiveTab == TabDashboard {
			m.Dashboard.Refresh()
		} else if m.ActiveTab == TabLogs && m.Logs.AutoFollow {
			m.Logs.Refresh()
		}

		return m, tickCmd()

	case tea.KeyMsg:
		// Global Quit
		if msg.String() == "ctrl+c" {
			if m.ActiveTab == TabChat && m.Chat.Streaming {
				// Abort stream instead of quitting immediately
				var chatCmd tea.Cmd
				m.Chat, chatCmd = m.Chat.Update(msg)
				return m, chatCmd
			}
			m.Quitting = true
			return m, tea.Quit
		}

		// Direct Tab Hotkeys (F1-F5, Ctrl+1..5)
		switch msg.String() {
		case "f1", "ctrl+1":
			m.ActiveTab = TabDashboard
			m.Dashboard.Refresh()
			return m, nil
		case "f2", "ctrl+2":
			m.ActiveTab = TabChat
			return m, nil
		case "f3", "ctrl+3":
			m.ActiveTab = TabModels
			m.Models.Refresh()
			return m, nil
		case "f4", "ctrl+4":
			m.ActiveTab = TabKnowledge
			m.Knowledge.Refresh()
			return m, nil
		case "f5", "ctrl+5":
			m.ActiveTab = TabLogs
			m.Logs.Refresh()
			return m, nil
		case "ctrl+t", "ctrl+n":
			m.ActiveTab = (m.ActiveTab + 1) % 5
			return m, nil
		case "ctrl+p":
			m.ActiveTab = (m.ActiveTab + 4) % 5
			return m, nil
		}

		// Tab cycling when not typing in text areas/inputs
		isTextInputActive := (m.ActiveTab == TabChat && m.Chat.TextArea.Focused()) ||
			(m.ActiveTab == TabModels && m.Models.Importing) ||
			(m.ActiveTab == TabKnowledge && m.Knowledge.IsSearching)

		if !isTextInputActive {
			switch msg.String() {
			case "tab":
				m.ActiveTab = (m.ActiveTab + 1) % 5
				return m, nil
			case "shift+tab":
				m.ActiveTab = (m.ActiveTab + 4) % 5
				return m, nil
			case "1":
				m.ActiveTab = TabDashboard
				m.Dashboard.Refresh()
				return m, nil
			case "2":
				m.ActiveTab = TabChat
				return m, nil
			case "3":
				m.ActiveTab = TabModels
				m.Models.Refresh()
				return m, nil
			case "4":
				m.ActiveTab = TabKnowledge
				m.Knowledge.Refresh()
				return m, nil
			case "5":
				m.ActiveTab = TabLogs
				m.Logs.Refresh()
				return m, nil
			case "q":
				m.Quitting = true
				return m, tea.Quit
			}
		}

		// Quick shortcuts on Dashboard
		if m.ActiveTab == TabDashboard {
			switch msg.String() {
			case "c":
				m.ActiveTab = TabChat
				return m, nil
			case "m":
				m.ActiveTab = TabModels
				m.Models.Refresh()
				return m, nil
			case "k":
				m.ActiveTab = TabKnowledge
				m.Knowledge.Refresh()
				return m, nil
			case "l":
				m.ActiveTab = TabLogs
				m.Logs.Refresh()
				return m, nil
			}
		}
	}

	// Dispatch to active sub-view
	switch m.ActiveTab {
	case TabDashboard:
		var cmd tea.Cmd
		m.Dashboard, cmd = m.Dashboard.Update(msg)
		cmds = append(cmds, cmd)
	case TabChat:
		var cmd tea.Cmd
		m.Chat, cmd = m.Chat.Update(msg)
		cmds = append(cmds, cmd)
	case TabModels:
		var cmd tea.Cmd
		m.Models, cmd = m.Models.Update(msg)
		cmds = append(cmds, cmd)
	case TabKnowledge:
		var cmd tea.Cmd
		m.Knowledge, cmd = m.Knowledge.Update(msg)
		cmds = append(cmds, cmd)
	case TabLogs:
		var cmd tea.Cmd
		m.Logs, cmd = m.Logs.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m AppModel) getUptime() string {
	d := time.Since(m.StartTime)
	h := int(d.Hours())
	min := int(d.Minutes()) % 60
	sec := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%02dh %02dm %02ds", h, min, sec)
	}
	return fmt.Sprintf("%02dm %02ds", min, sec)
}

func (m AppModel) View() string {
	if m.Quitting {
		return styles.DescStyle.Render("\n  🐺 Exiting Wolarc TUI. Have a productive day!\n\n")
	}

	// 1. Header Bar
	header := components.RenderHeader(m.Width, m.IsOnline, m.getUptime(), m.TunnelURL)

	// 2. Tab Navigation Bar
	tabs := []string{
		"1 ⚡ Dashboard",
		"2 💬 Chat",
		"3 📦 Models",
		"4 📚 Knowledge",
		"5 📋 Logs",
	}

	var renderedTabs []string
	for i, t := range tabs {
		if i == m.ActiveTab {
			renderedTabs = append(renderedTabs, styles.ActiveTabStyle.Render(t))
		} else {
			renderedTabs = append(renderedTabs, styles.TabStyle.Render(t))
		}
	}
	tabRow := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	tabBorder := lipgloss.NewStyle().
		Foreground(styles.ColorBorder).
		Render(strings.Repeat("─", max(20, m.Width-lipgloss.Width(tabRow)-4)))
	tabBar := lipgloss.JoinHorizontal(lipgloss.Center, tabRow, tabBorder)

	// 3. Active Tab View
	var currentView string
	switch m.ActiveTab {
	case TabDashboard:
		currentView = m.Dashboard.View()
	case TabChat:
		currentView = m.Chat.View()
	case TabModels:
		currentView = m.Models.View()
	case TabKnowledge:
		currentView = m.Knowledge.View()
	case TabLogs:
		currentView = m.Logs.View()
	}

	// 4. Global Footer Keymaps
	globalHints := fmt.Sprintf(
		" %s Switch Tab  •  %s Direct Tab  •  %s Quit",
		styles.KeyStyle.Render("Tab / Ctrl+T"),
		styles.KeyStyle.Render("F1-F5"),
		styles.KeyStyle.Render("q / Ctrl+C"),
	)
	footer := styles.DescStyle.Render(globalHints)

	return lipgloss.JoinVertical(lipgloss.Left, header, tabBar, currentView, footer)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

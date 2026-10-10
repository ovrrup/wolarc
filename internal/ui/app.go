package ui

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/spinner"
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

func tickCmd(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
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
	Palette   components.PaletteModel
	HelpModal components.HelpModalModel
	Toast     *components.Toast
	Width     int
	Height    int
	StartTime time.Time
	IsOnline  bool
	TunnelURL string
	Quitting  bool
	TickCount int
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
		Palette:   components.NewPaletteModel(),
		HelpModal: components.NewHelpModalModel(),
		Width:     80,
		Height:    24,
		StartTime: time.Now(),
		IsOnline:  online,
		TunnelURL: config.GetTunnelURL(),
	}
}

func (m *AppModel) SetToast(msg string, t components.ToastType) {
	toast := components.NewToast(msg, t)
	m.Toast = &toast
}

func (m AppModel) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		tickCmd(time.Second),
		m.Dashboard.Init(),
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

	case spinner.TickMsg:
		// Broadcast spinner tick to active view with animation
		var cmd tea.Cmd
		if m.ActiveTab == TabDashboard {
			m.Dashboard, cmd = m.Dashboard.Update(msg)
			cmds = append(cmds, cmd)
		} else if m.ActiveTab == TabChat {
			m.Chat, cmd = m.Chat.Update(msg)
			cmds = append(cmds, cmd)
		} else if m.ActiveTab == TabKnowledge {
			m.Knowledge, cmd = m.Knowledge.Update(msg)
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	case TickMsg:
		m.TickCount++
		services := system.CheckAllServices(m.Config.ChatPort, m.Config.EmbedPort)
		m.IsOnline = services.ChatServer.Running || services.AgentGateway.Running
		m.TunnelURL = config.GetTunnelURL()

		if m.ActiveTab == TabDashboard {
			m.Dashboard.Refresh()
		} else if m.ActiveTab == TabLogs && m.Logs.AutoFollow {
			m.Logs.Refresh()
		}

		// Adaptive tick rate: fast 200ms when chat is streaming or daemons warming up, 1s when idle
		interval := time.Second
		if m.Chat.Streaming || m.Dashboard.Services.ChatServer.Running {
			interval = 250 * time.Millisecond
		}
		return m, tickCmd(interval)

	case tea.KeyMsg:
		// 1. If Command Palette is open, route directly to Palette
		if m.Palette.Active {
			var cmd tea.Cmd
			var action *components.PaletteActionID
			m.Palette, cmd, action = m.Palette.Update(msg)
			if action != nil {
				m.handlePaletteAction(*action)
			}
			return m, cmd
		}

		// 2. If Help Modal is open, handle dismissal
		if m.HelpModal.Active {
			switch msg.String() {
			case "esc", "?", "q":
				m.HelpModal.Active = false
				return m, nil
			}
			return m, nil
		}

		// 3. Global hotkeys
		switch msg.String() {
		case "ctrl+k", ":":
			// Open Command Palette
			m.Palette.Open()
			return m, nil
		case "?":
			// Toggle Help Modal
			isTyping := (m.ActiveTab == TabChat && m.Chat.TextArea.Focused()) ||
				(m.ActiveTab == TabModels && m.Models.Importing) ||
				(m.ActiveTab == TabKnowledge && m.Knowledge.IsSearching)
			if !isTyping {
				m.HelpModal.Toggle()
				return m, nil
			}
		case "ctrl+c":
			if m.ActiveTab == TabChat && m.Chat.Streaming {
				// Abort active LLM token stream
				var chatCmd tea.Cmd
				m.Chat, chatCmd = m.Chat.Update(msg)
				m.SetToast("Aborted generation stream", components.ToastWarning)
				return m, chatCmd
			}
			m.Quitting = true
			return m, tea.Quit
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
		case "ctrl+n":
			m.ActiveTab = (m.ActiveTab + 1) % 5
			return m, nil
		case "ctrl+p":
			m.ActiveTab = (m.ActiveTab + 4) % 5
			return m, nil
		}

		// 4. Tab cycling when text inputs / modal sub-modes are not focused
		isTextInputActive := (m.ActiveTab == TabChat && m.Chat.TextArea.Focused()) ||
			(m.ActiveTab == TabModels && (m.Models.Importing || m.Models.Deleting)) ||
			(m.ActiveTab == TabKnowledge && m.Knowledge.IsSearching) ||
			(m.ActiveTab == TabLogs && m.Logs.IsFiltering) ||
			(m.ActiveTab == TabDashboard && m.Dashboard.Tuning)

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

		// 5. Dashboard-specific single-key shortcuts (when not in tuning mode)
		if m.ActiveTab == TabDashboard && !m.Dashboard.Tuning {
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

func (m *AppModel) handlePaletteAction(action components.PaletteActionID) {
	switch action {
	case components.ActionStartStack:
		_ = system.StartServices()
		m.SetToast("Booting Wolarc stack (wola.r)...", components.ToastSuccess)
		m.Dashboard.Refresh()
	case components.ActionStopStack:
		_ = system.StopServices()
		m.SetToast("Services cleanly stopped (wola.k)", components.ToastWarning)
		m.Dashboard.Refresh()
	case components.ActionRestartStack:
		_ = system.RestartServices()
		m.SetToast("Restarting Wolarc stack (wola.rs)...", components.ToastInfo)
		m.Dashboard.Refresh()
	case components.ActionCopyURL:
		url := config.GetTunnelURL()
		if url != "" {
			_ = clipboard.WriteAll(url)
			_ = exec.Command("termux-clipboard-set", url).Run()
			m.SetToast("Copied tunnel URL to clipboard: "+url, components.ToastSuccess)
		} else {
			m.SetToast("No active tunnel URL found", components.ToastWarning)
		}
	case components.ActionLocalGateway:
		out, err := exec.Command("wola.url", "local-raw").Output()
		localURL := strings.TrimSpace(string(out))
		if err == nil && localURL != "" {
			_ = clipboard.WriteAll(localURL)
			_ = exec.Command("termux-clipboard-set", localURL).Run()
			m.SetToast("Copied offline link to clipboard: "+localURL, components.ToastSuccess)
		} else {
			fallbackURL := fmt.Sprintf("http://192.168.43.1:%s", m.Config.AgentPort)
			_ = clipboard.WriteAll(fallbackURL)
			_ = exec.Command("termux-clipboard-set", fallbackURL).Run()
			m.SetToast("Copied hotspot link: "+fallbackURL, components.ToastSuccess)
		}
	case components.ActionRunIndexer:
		go func() {
			cmd := exec.Command("wola.i", "--once")
			_ = cmd.Run()
		}()
		m.SetToast("Triggered background document vectorizer pass", components.ToastInfo)
		m.Knowledge.Refresh()
	case components.ActionRunBench:
		m.SetToast("Run 'wola.bench' in terminal for optimal thread benchmarking", components.ToastInfo)
	case components.ActionGotoDash:
		m.ActiveTab = TabDashboard
		m.Dashboard.Refresh()
	case components.ActionGotoChat:
		m.ActiveTab = TabChat
	case components.ActionGotoModels:
		m.ActiveTab = TabModels
		m.Models.Refresh()
	case components.ActionGotoKnow:
		m.ActiveTab = TabKnowledge
		m.Knowledge.Refresh()
	case components.ActionGotoLogs:
		m.ActiveTab = TabLogs
		m.Logs.Refresh()
	case components.ActionReloadConfig:
		m.Config = config.LoadConfig()
		m.Dashboard.Config = m.Config
		m.Chat.Config = m.Config
		m.Models.Config = m.Config
		m.Knowledge.Config = m.Config
		m.Logs.Config = m.Config
		m.SetToast("Configuration reloaded from ~/.wolarc", components.ToastSuccess)
	case components.ActionQuitApp:
		m.Quitting = true
	}
}

func (m AppModel) getTabName() string {
	switch m.ActiveTab {
	case TabDashboard:
		return "Dashboard"
	case TabChat:
		return "Interactive Chat"
	case TabModels:
		return "Models & Storage"
	case TabKnowledge:
		return "Vector Knowledge"
	case TabLogs:
		return "Multi-Daemon Logs"
	default:
		return "Overview"
	}
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

	// 1. Animated Header Bar with Breadcrumbs & Radar Pulse
	header := components.RenderHeader(m.Width, m.IsOnline, m.getUptime(), m.TunnelURL, m.TickCount, m.getTabName())

	// 2. Tab Navigation Bar with Micro-transitions
	tabLabels := []struct {
		Icon string
		Name string
	}{
		{"⚡", "Dashboard"},
		{"💬", "Chat"},
		{"📦", "Models"},
		{"📚", "Knowledge"},
		{"📋", "Logs"},
	}

	var renderedTabs []string
	for i, t := range tabLabels {
		if i == m.ActiveTab {
			label := fmt.Sprintf("▸ %d %s %s ◂", i+1, t.Icon, t.Name)
			renderedTabs = append(renderedTabs, styles.ActiveTabStyle.Render(label))
		} else {
			label := fmt.Sprintf("%d %s %s", i+1, t.Icon, t.Name)
			renderedTabs = append(renderedTabs, styles.TabStyle.Render(label))
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

	// 4. Global Footer Keymaps & Micro-Toast Pill
	var footerRow string
	toastView := components.RenderToast(m.Toast, m.Width)
	if toastView != "" {
		footerRow = toastView
	} else {
		globalHints := fmt.Sprintf(
			" %s Switch Tab  •  %s Direct Tab  •  %s Quick Palette  •  %s Help  •  %s Quit",
			styles.KeyStyle.Render("Tab / Shift+Tab"),
			styles.KeyStyle.Render("F1-F5"),
			styles.KeyStyle.Render("Ctrl+K / :"),
			styles.KeyStyle.Render("?"),
			styles.KeyStyle.Render("q / Ctrl+C"),
		)
		footerRow = styles.DescStyle.Render(globalHints)
	}

	baseLayout := lipgloss.JoinVertical(lipgloss.Left, header, tabBar, currentView, footerRow)

	// 5. Modal Overlays (Command Palette & Help Cheat Sheet)
	if m.Palette.Active {
		return m.Palette.View(m.Width, m.Height)
	}
	if m.HelpModal.Active {
		return m.HelpModal.View(m.Width, m.Height)
	}

	return baseLayout
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

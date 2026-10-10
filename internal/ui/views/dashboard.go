package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ovrrup/wolarc/internal/config"
	"github.com/ovrrup/wolarc/internal/system"
	"github.com/ovrrup/wolarc/internal/ui/styles"
)

type DashboardModel struct {
	Config          *config.Config
	Services        system.ServiceStatus
	RAM             system.RAMInfo
	Metrics         config.MetricsData
	VelocityHistory []float64
	Spinner         spinner.Model
	Tuning          bool
	TunerIndex      int
	StatusMsg       string
	Width           int
	Height          int
}

func NewDashboardModel(cfg *config.Config) DashboardModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(styles.ColorCyan)

	m := DashboardModel{
		Config:          cfg,
		Services:        system.CheckAllServices(cfg.ChatPort, cfg.EmbedPort),
		RAM:             system.GetRAMInfo(),
		Metrics:         config.GetMetrics(cfg),
		VelocityHistory: make([]float64, 0, 16),
		Spinner:         sp,
		Tuning:          false,
		TunerIndex:      0,
		Width:           80,
		Height:          24,
	}
	m.Refresh()
	return m
}

func (m *DashboardModel) Refresh() {
	m.Services = system.CheckAllServices(m.Config.ChatPort, m.Config.EmbedPort)
	m.RAM = system.GetRAMInfo()
	m.Metrics = config.GetMetrics(m.Config)

	// Update velocity history
	vel := m.Metrics.PredictedPerSecond
	if len(m.VelocityHistory) >= 16 {
		m.VelocityHistory = append(m.VelocityHistory[1:], vel)
	} else {
		m.VelocityHistory = append(m.VelocityHistory, vel)
	}
}

func (m DashboardModel) Init() tea.Cmd {
	return m.Spinner.Tick
}

func (m DashboardModel) Update(msg tea.Msg) (DashboardModel, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case spinner.TickMsg:
		m.Spinner, cmd = m.Spinner.Update(msg)
		return m, cmd
	case tea.KeyMsg:
		if m.Tuning {
			switch msg.String() {
			case "esc", "t":
				m.Tuning = false
				m.StatusMsg = "Exited knob tuner"
				return m, nil
			case "up", "k":
				if m.TunerIndex > 0 {
					m.TunerIndex--
				}
				return m, nil
			case "down", "j":
				if m.TunerIndex < 3 {
					m.TunerIndex++
				}
				return m, nil
			case "left", "-", "h":
				m.adjustTuner(-1)
				return m, nil
			case "right", "+", "l":
				m.adjustTuner(1)
				return m, nil
			case "enter":
				m.saveTunerSettings()
				m.Tuning = false
				m.StatusMsg = "Saved runtime knobs to ~/.wolarc!"
				return m, nil
			}
			return m, nil
		}

		switch msg.String() {
		case "t":
			m.Tuning = true
			m.StatusMsg = "Interactive Tuning Mode: use arrow keys to adjust knobs"
			return m, nil
		case "s":
			if m.Services.ChatServer.Running || m.Services.AgentGateway.Running {
				_ = system.StopServices()
			} else {
				_ = system.StartServices()
			}
			m.Refresh()
			return m, nil
		case "r":
			_ = system.RestartServices()
			m.Refresh()
			return m, nil
		}
	}
	return m, nil
}

func (m *DashboardModel) adjustTuner(dir int) {
	switch m.TunerIndex {
	case 0: // Temperature
		m.Config.Temperature += float64(dir) * 0.05
		if m.Config.Temperature < 0.05 {
			m.Config.Temperature = 0.05
		}
		if m.Config.Temperature > 1.20 {
			m.Config.Temperature = 1.20
		}
	case 1: // Similarity Threshold
		m.Config.SimilarityThreshold += float64(dir) * 0.05
		if m.Config.SimilarityThreshold < 0.40 {
			m.Config.SimilarityThreshold = 0.40
		}
		if m.Config.SimilarityThreshold > 0.95 {
			m.Config.SimilarityThreshold = 0.95
		}
	case 2: // Top-K
		m.Config.TopK += dir
		if m.Config.TopK < 1 {
			m.Config.TopK = 1
		}
		if m.Config.TopK > 8 {
			m.Config.TopK = 8
		}
	case 3: // Chat Threads
		m.Config.ChatThreads += dir
		if m.Config.ChatThreads < 1 {
			m.Config.ChatThreads = 1
		}
		if m.Config.ChatThreads > 16 {
			m.Config.ChatThreads = 16
		}
	}
}

func (m *DashboardModel) saveTunerSettings() {
	_ = config.SetConfigKey("TEMPERATURE", fmt.Sprintf("%.2f", m.Config.Temperature))
	_ = config.SetConfigKey("SIMILARITY_THRESHOLD", fmt.Sprintf("%.2f", m.Config.SimilarityThreshold))
	_ = config.SetConfigKey("TOP_K", fmt.Sprintf("%d", m.Config.TopK))
	_ = config.SetConfigKey("CHAT_THREADS", fmt.Sprintf("%d", m.Config.ChatThreads))
}

func (m DashboardModel) View() string {
	cardWidth := (m.Width - 6) / 2
	if cardWidth < 34 {
		cardWidth = 34
	}

	// 1. INFERENCE TELEMETRY CARD
	speedVal := fmt.Sprintf("%.1f tok/s", m.Metrics.PredictedPerSecond)
	promptVal := fmt.Sprintf("%.1f tok/s", m.Metrics.PromptPerSecond)
	totalTokens := fmt.Sprintf("%d tokens", m.Metrics.TotalPredictedTokens)
	reqCount := fmt.Sprintf("%d completed", m.Metrics.RequestsCompleted)

	speedPercent := int(m.Metrics.PredictedPerSecond * 3.5)
	if speedPercent > 100 {
		speedPercent = 100
	}
	speedGauge := styles.GradientBar(float64(speedPercent), 14)
	sparkline := styles.Sparkline(m.VelocityHistory, 35.0)
	if sparkline == "" {
		sparkline = styles.DescStyle.Render("── (warming up)")
	}

	telemetryContent := fmt.Sprintf(
		"%s\n\n"+
			"• Gen Velocity : %s  %s\n"+
			"• Speed Spark  : %s\n"+
			"• Prompt Eval  : %s\n"+
			"• Output Total : %s\n"+
			"• Total Queries: %s",
		styles.CardHeader.Render("⚡ INFERENCE TELEMETRY"),
		styles.KeyStyle.Render(speedVal),
		speedGauge,
		sparkline,
		styles.DescStyle.Render(promptVal),
		styles.DescStyle.Render(totalTokens),
		styles.DescStyle.Render(reqCount),
	)
	telemetryCard := styles.CardStyle.Width(cardWidth).Render(telemetryContent)

	// 2. HARDWARE & MEMORY CARD
	ramUsedStr := fmt.Sprintf("%d MB / %d MB (%d%%)", m.RAM.UsedMB, m.RAM.TotalMB, m.RAM.UsedPercent)
	ramGauge := styles.GradientBar(float64(m.RAM.UsedPercent), 14)

	threadsStr := fmt.Sprintf("%d chat / %d embed", m.Config.ChatThreads, m.Config.EmbedThreads)
	flashStr := m.Config.FlashAttention
	mlockStr := "Disabled"
	if m.Config.MLock {
		mlockStr = "Active"
	}

	ramContent := fmt.Sprintf(
		"%s\n\n"+
			"• Memory Usage : %s\n"+
			"• Memory Bar   : %s\n"+
			"• CPU Threads  : %s\n"+
			"• Flash-Attn   : %s  |  MLock: %s",
		styles.CardHeader.Render("🧠 HARDWARE & MEMORY"),
		styles.KeyStyle.Render(ramUsedStr),
		ramGauge,
		styles.DescStyle.Render(threadsStr),
		styles.DescStyle.Render(flashStr),
		styles.DescStyle.Render(mlockStr),
	)
	ramCard := styles.CardStyle.Width(cardWidth).Render(ramContent)

	row1 := lipgloss.JoinHorizontal(lipgloss.Top, telemetryCard, " ", ramCard)

	// 3. SERVICE DAEMONS HEALTH CARD (with animated spinners)
	renderServiceLine := func(name string, running bool, port string, pid int) string {
		var badge string
		if running {
			badge = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#0f172a")).
				Background(styles.ColorGreen).
				Padding(0, 1).
				Render(m.Spinner.View() + " ACTIVE")
		} else {
			badge = styles.BadgeOffline.Render("○ STOPPED")
		}
		var meta string
		if running {
			meta = fmt.Sprintf("(Port %s • PID %d)", port, pid)
		} else {
			meta = fmt.Sprintf("(Port %s)", port)
		}
		return fmt.Sprintf(" %-17s %s  %s", name, badge, styles.DescStyle.Render(meta))
	}

	servicesContent := fmt.Sprintf(
		"%s\n\n"+
			"%s\n"+
			"%s\n"+
			"%s\n"+
			"%s\n"+
			"%s",
		styles.CardHeader.Render("🛡️  DAEMON ORCHESTRATION"),
		renderServiceLine("Chat Engine", m.Services.ChatServer.Running, m.Config.ChatPort, m.Services.ChatServer.PID),
		renderServiceLine("Embed Engine", m.Services.EmbedServer.Running, m.Config.EmbedPort, m.Services.EmbedServer.PID),
		renderServiceLine("Agent Gateway", m.Services.AgentGateway.Running, m.Config.AgentPort, m.Services.AgentGateway.PID),
		renderServiceLine("Public Tunnel", m.Services.Tunnel.Running, "80", m.Services.Tunnel.PID),
		renderServiceLine("RAG Indexer", m.Services.Indexer.Running, "auto", m.Services.Indexer.PID),
	)
	servicesCard := styles.CardStyle.Width(cardWidth).Render(servicesContent)

	// 4. ACTIVE MODELS & TUNING CARD OR INTERACTIVE TUNER DRAWER
	var modelsCard string
	if m.Tuning {
		var tunerSb strings.Builder
		tunerSb.WriteString(styles.CardHeader.Render("🎛️  LIVE KNOB TUNER (Press Enter to Save)") + "\n\n")

		renderTunerRow := func(idx int, label, val string, bar string) {
			cursor := "  "
			lblStyle := styles.DescStyle
			if idx == m.TunerIndex {
				cursor = styles.KeyStyle.Render("▶ ")
				lblStyle = lipgloss.NewStyle().Bold(true).Foreground(styles.ColorCyan)
			}
			tunerSb.WriteString(fmt.Sprintf("%s%-18s : %-6s %s\n", cursor, lblStyle.Render(label), styles.KeyStyle.Render(val), bar))
		}

		tempBar := styles.GradientBar((m.Config.Temperature/1.2)*100, 10)
		simBar := styles.GradientBar((m.Config.SimilarityThreshold/1.0)*100, 10)
		topKBar := styles.GradientBar((float64(m.Config.TopK)/8.0)*100, 10)
		threadBar := styles.GradientBar((float64(m.Config.ChatThreads)/16.0)*100, 10)

		renderTunerRow(0, "Temperature", fmt.Sprintf("%.2f", m.Config.Temperature), tempBar)
		renderTunerRow(1, "Sim Threshold", fmt.Sprintf("%.2f", m.Config.SimilarityThreshold), simBar)
		renderTunerRow(2, "Top-K Chunks", fmt.Sprintf("%d", m.Config.TopK), topKBar)
		renderTunerRow(3, "Chat Threads", fmt.Sprintf("%d", m.Config.ChatThreads), threadBar)

		tunerSb.WriteString("\n" + styles.DescStyle.Render("↑/↓: Select knob • ←/→: Adjust • Enter: Save • Esc: Exit"))
		modelsCard = styles.CardActiveStyle.Width(cardWidth).Render(tunerSb.String())
	} else {
		chatModel := m.Config.ChatModel
		if chatModel == "" {
			chatModel = "None (Set via models tab)"
		}
		embedModel := m.Config.EmbedModel
		if embedModel == "" {
			embedModel = "None (Set via models tab)"
		}

		modelsContent := fmt.Sprintf(
			"%s\n\n"+
				"• Active Chat  : %s\n"+
				"• Active Embed : %s\n\n"+
				"• Temperature  : %.2f\n"+
				"• Top-K Chunks : %d\n"+
				"• Sim-Threshold: %.2f\n"+
				"• Parallel Slot: %d requests",
			styles.CardHeader.Render("📦 ACTIVE MODELS & KNOBS"),
			styles.KeyStyle.Render(chatModel),
			styles.DescStyle.Render(embedModel),
			m.Config.Temperature,
			m.Config.TopK,
			m.Config.SimilarityThreshold,
			m.Config.ChatParallel,
		)
		modelsCard = styles.CardStyle.Width(cardWidth).Render(modelsContent)
	}

	row2 := lipgloss.JoinHorizontal(lipgloss.Top, servicesCard, " ", modelsCard)

	// 5. QUICK ACTIONS FOOTER
	actions := []string{
		fmt.Sprintf("%s Start/Stop", styles.KeyStyle.Render("[s]")),
		fmt.Sprintf("%s Restart", styles.KeyStyle.Render("[r]")),
		fmt.Sprintf("%s Tune Knobs", styles.KeyStyle.Render("[t]")),
		fmt.Sprintf("%s Open Chat", styles.KeyStyle.Render("[c]")),
		fmt.Sprintf("%s Models", styles.KeyStyle.Render("[m]")),
		fmt.Sprintf("%s Knowledge", styles.KeyStyle.Render("[k]")),
		fmt.Sprintf("%s Logs", styles.KeyStyle.Render("[l]")),
		fmt.Sprintf("%s Palette", styles.KeyStyle.Render("[Ctrl+K]")),
	}
	actionsBar := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorBorder).
		Width(cardWidth*2 + 1).
		Padding(0, 1).
		Render("Actions: " + strings.Join(actions, "  │  "))

	return lipgloss.JoinVertical(lipgloss.Left, row1, " ", row2, " ", actionsBar)
}

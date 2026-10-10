package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ovrrup/wolarc/internal/config"
	"github.com/ovrrup/wolarc/internal/system"
	"github.com/ovrrup/wolarc/internal/ui/styles"
)

type DashboardModel struct {
	Config   *config.Config
	Services system.ServiceStatus
	RAM      system.RAMInfo
	Metrics  config.MetricsData
	Width    int
	Height   int
}

func NewDashboardModel(cfg *config.Config) DashboardModel {
	return DashboardModel{
		Config:   cfg,
		Services: system.CheckAllServices(cfg.ChatPort, cfg.EmbedPort),
		RAM:      system.GetRAMInfo(),
		Metrics:  config.GetMetrics(cfg),
		Width:    80,
		Height:   24,
	}
}

func (m *DashboardModel) Refresh() {
	m.Services = system.CheckAllServices(m.Config.ChatPort, m.Config.EmbedPort)
	m.RAM = system.GetRAMInfo()
	m.Metrics = config.GetMetrics(m.Config)
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
	speedGauge := styles.Gauge(speedPercent, 12)

	telemetryContent := fmt.Sprintf(
		"%s  %s\n\n"+
			"• Gen Velocity : %s  %s\n"+
			"• Prompt Eval  : %s\n"+
			"• Output Total : %s\n"+
			"• Total Queries: %s",
		styles.CardHeader.Render("⚡ INFERENCE TELEMETRY"),
		"",
		styles.KeyStyle.Render(speedVal),
		speedGauge,
		styles.DescStyle.Render(promptVal),
		styles.DescStyle.Render(totalTokens),
		styles.DescStyle.Render(reqCount),
	)
	telemetryCard := styles.CardStyle.Width(cardWidth).Render(telemetryContent)

	// 2. HARDWARE & MEMORY CARD
	ramUsedStr := fmt.Sprintf("%d MB / %d MB (%d%%)", m.RAM.UsedMB, m.RAM.TotalMB, m.RAM.UsedPercent)
	ramGauge := styles.Gauge(m.RAM.UsedPercent, 12)

	threadsStr := fmt.Sprintf("%d chat / %d embed", m.Config.ChatThreads, m.Config.EmbedThreads)
	flashStr := m.Config.FlashAttention
	mlockStr := "Disabled"
	if m.Config.MLock {
		mlockStr = "Active"
	}

	ramContent := fmt.Sprintf(
		"%s  %s\n\n"+
			"• Memory Usage : %s\n"+
			"• Gauge Bar    : %s\n"+
			"• CPU Threads  : %s\n"+
			"• Flash-Attn   : %s  |  MLock: %s",
		styles.CardHeader.Render("🧠 HARDWARE & MEMORY"),
		"",
		styles.KeyStyle.Render(ramUsedStr),
		ramGauge,
		styles.DescStyle.Render(threadsStr),
		styles.DescStyle.Render(flashStr),
		styles.DescStyle.Render(mlockStr),
	)
	ramCard := styles.CardStyle.Width(cardWidth).Render(ramContent)

	row1 := lipgloss.JoinHorizontal(lipgloss.Top, telemetryCard, " ", ramCard)

	// 3. SERVICE DAEMONS HEALTH CARD
	renderServiceLine := func(name string, running bool, port string, pid int) string {
		var badge string
		if running {
			badge = styles.BadgeOnline.Render("RUNNING")
		} else {
			badge = styles.BadgeOffline.Render("STOPPED")
		}
		var meta string
		if running {
			meta = fmt.Sprintf("(Port %s • PID %d)", port, pid)
		} else {
			meta = fmt.Sprintf("(Port %s)", port)
		}
		return fmt.Sprintf(" %-18s %s  %s", name, badge, styles.DescStyle.Render(meta))
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

	// 4. ACTIVE MODELS & TUNING CARD
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
	modelsCard := styles.CardStyle.Width(cardWidth).Render(modelsContent)

	row2 := lipgloss.JoinHorizontal(lipgloss.Top, servicesCard, " ", modelsCard)

	// 5. QUICK ACTIONS FOOTER
	actions := []string{
		fmt.Sprintf("%s Start/Stop", styles.KeyStyle.Render("[s]")),
		fmt.Sprintf("%s Restart", styles.KeyStyle.Render("[r]")),
		fmt.Sprintf("%s Open Chat", styles.KeyStyle.Render("[c]")),
		fmt.Sprintf("%s Models", styles.KeyStyle.Render("[m]")),
		fmt.Sprintf("%s Knowledge", styles.KeyStyle.Render("[k]")),
		fmt.Sprintf("%s Logs", styles.KeyStyle.Render("[l]")),
		fmt.Sprintf("%s Quit", styles.KeyStyle.Render("[q]")),
	}
	actionsBar := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorBorder).
		Width(cardWidth*2 + 1).
		Padding(0, 1).
		Render("Actions: " + strings.Join(actions, "  │  "))

	return lipgloss.JoinVertical(lipgloss.Left, row1, " ", row2, " ", actionsBar)
}

func (m DashboardModel) Init() tea.Cmd {
	return nil
}

func (m DashboardModel) Update(msg tea.Msg) (DashboardModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
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

package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ovrrup/wolarc/internal/config"
	"github.com/ovrrup/wolarc/internal/system"
	"github.com/ovrrup/wolarc/internal/ui"
)

const Version = "2.0.0 (Go + Bubble Tea)"

func printHelp() {
	fmt.Printf(`🐺 WOLARC TUI - High-Performance Autonomous Local RAG Suite
Version: %s

USAGE:
    wola-tui [SUBCOMMAND] [FLAGS]

SUBCOMMANDS:
    dashboard, d       Launch TUI with System Dashboard (default)
    chat, c            Launch TUI directly in Interactive RAG Chat
    models, m          Launch TUI Model Management & Local Importer
    knowledge, k, i    Launch TUI Knowledge Base & Semantic Tester
    logs, l            Launch TUI Multi-Stream Realtime Log Tailer
    status             Print quick non-interactive node telemetry

HOTKEYS INSIDE TUI:
    Tab / Shift+Tab    Cycle between views
    F1 - F5 / Ctrl+1-5 Switch directly to tabs [Dashboard, Chat, Models, Knowledge, Logs]
    Ctrl+T / Ctrl+N    Next tab
    Ctrl+P             Previous tab
    q / Ctrl+C         Quit application
`, Version)
}

func printStatus(cfg *config.Config) {
	services := system.CheckAllServices(cfg.ChatPort, cfg.EmbedPort)
	ram := system.GetRAMInfo()
	metrics := config.GetMetrics(cfg)
	tunnelURL := config.GetTunnelURL(cfg)

	fmt.Println("=================================================================")
	fmt.Printf("                   WOLARC NODE STATUS (%s)\n", Version)
	fmt.Println("=================================================================")

	renderStatus := func(name string, running bool, port string, pid int) {
		status := "STOPPED"
		if running {
			status = fmt.Sprintf("RUNNING (PID %d)", pid)
		}
		fmt.Printf("  • %-16s (Port %s): %s\n", name, port, status)
	}

	renderStatus("Chat Server", services.ChatServer.Running, cfg.ChatPort, services.ChatServer.PID)
	renderStatus("Embed Server", services.EmbedServer.Running, cfg.EmbedPort, services.EmbedServer.PID)
	renderStatus("Agent Gateway", services.AgentGateway.Running, cfg.AgentPort, services.AgentGateway.PID)
	renderStatus("Public Tunnel", services.Tunnel.Running, "80", services.Tunnel.PID)
	renderStatus("Indexer Daemon", services.Indexer.Running, "auto", services.Indexer.PID)

	fmt.Println("-----------------------------------------------------------------")
	fmt.Printf("  Memory Usage     : %d MB / %d MB (%d%%)\n", ram.UsedMB, ram.TotalMB, ram.UsedPercent)
	fmt.Printf("  Active Chat Model: %s\n", cfg.ChatModel)
	fmt.Printf("  Active Embed Mod : %s\n", cfg.EmbedModel)
	if tunnelURL != "" {
		fmt.Printf("  Gateway Tunnel   : %s\n", tunnelURL)
	} else {
		fmt.Println("  Gateway Tunnel   : Inactive")
	}
	if metrics.TotalPredictedTokens > 0 {
		fmt.Printf("  Inference Speed  : %.1f tok/s (Prompt: %.1f tok/s)\n", metrics.PredictedPerSecond, metrics.PromptPerSecond)
		fmt.Printf("  Total Generation : %d tokens (%d requests)\n", metrics.TotalPredictedTokens, metrics.RequestsCompleted)
	}
	fmt.Println("=================================================================")
}

func main() {
	cfg := config.LoadConfig()
	initialTab := ui.TabDashboard

	if len(os.Args) > 1 {
		arg := strings.ToLower(os.Args[1])
		switch arg {
		case "-h", "--help", "help":
			printHelp()
			return
		case "-v", "--version", "version":
			fmt.Printf("wola-tui %s\n", Version)
			return
		case "status":
			printStatus(cfg)
			return
		case "d", "dash", "dashboard":
			initialTab = ui.TabDashboard
		case "c", "chat":
			initialTab = ui.TabChat
		case "m", "model", "models":
			initialTab = ui.TabModels
		case "k", "know", "knowledge", "i", "index":
			initialTab = ui.TabKnowledge
		case "l", "log", "logs":
			initialTab = ui.TabLogs
		default:
			// If unknown argument, treat as search or start default dashboard
		}
	}

	app := ui.NewAppModel(cfg, initialTab)
	p := tea.NewProgram(
		app,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running Wolarc TUI: %v\n", err)
		os.Exit(1)
	}
}

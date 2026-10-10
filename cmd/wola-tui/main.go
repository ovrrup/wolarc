package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ovrrup/wolarc/internal/config"
	"github.com/ovrrup/wolarc/internal/system"
	"github.com/ovrrup/wolarc/internal/ui"
)

const Version = "2.0.0 (Go + Bubble Tea)"

func printHelp() {
	fmt.Printf(`🐺 WOLARC - Autonomous Local RAG Suite & TUI
Version: %s

USAGE:
    wola [SUBCOMMAND] [FLAGS]
    wola-tui [SUBCOMMAND] [FLAGS]

COMMAND PAIRS (Casual / Technical):
    start, r           Start all services (chat, embed, agent, tunnel)
    stop, k            Stop all services cleanly
    restart, rs        Restart all services
    check, status      Display quick non-interactive node telemetry
    dashboard, d       System Dashboard & Live Telemetry
    chat, c            Interactive RAG Chat client (or pass query: wola chat "text")
    models, m          Model Manager & GGUF switcher
    store, gt          In-terminal Model Store & Downloader
    knowledge, i       Vector Knowledge Base & Semantic Search tester
    bench, b           Hardware benchmark & inference tok/s tester
    log, l             Multi-Stream Realtime Log Tailer
    think, th          Toggle or view reasoning thinking stream
    link, url          Display active Gateway Tunnel URL & QR code
    hotspot, local     Display Local Offline Gateway & Hotspot Pairing
    tailscale, ts      Inspect Tailscale mesh IP & connection links
    config, t          Inspect or update configuration settings
    update, u          Check and install updates from GitHub
    uninstall, purge   Clean removal & system reset
    help, h            Display help and command cheat sheet
    ui, tui            Toggle default launch mode (TUI vs CLI suite)

HOTKEYS INSIDE TUI:
    Tab / Shift+Tab    Cycle between views
    F1 - F5 / Ctrl+1-5 Switch directly to tabs [Dashboard, Chat, Models, Knowledge, Logs]
    Ctrl+K / :         Open fuzzy Command Palette
    Ctrl+T / Ctrl+N    Next tab
    Ctrl+P             Previous tab
    ? / F1             Show Help modal
    q / Ctrl+C         Quit application
`, Version)
}

func printStatus(cfg *config.Config) {
	services := system.CheckAllServices(cfg.ChatPort, cfg.EmbedPort)
	ram := system.GetRAMInfo()
	metrics := config.GetMetrics(cfg)
	tunnelURL := config.GetTunnelURL()
	tsIP := system.GetTailscaleIP()

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
	if tsIP != "" {
		fmt.Printf("  Tailscale Mesh   : http://%s:%s (Active)\n", tsIP, cfg.AgentPort)
	} else {
		fmt.Println("  Tailscale Mesh   : Inactive")
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

	if len(os.Args) <= 1 {
		if !cfg.EnableTUI && !strings.HasSuffix(os.Args[0], "wola-tui") {
			cmd := exec.Command("wola-h")
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fallback := exec.Command("wola.d")
				fallback.Stdin = os.Stdin
				fallback.Stdout = os.Stdout
				fallback.Stderr = os.Stderr
				_ = fallback.Run()
			}
			return
		}
	} else {
		arg := strings.ToLower(os.Args[1])
		switch arg {
		case "-h", "--help", "help", "h":
			printHelp()
			return
		case "-v", "--version", "version":
			fmt.Printf("wola-tui %s\n", Version)
			return
		case "status", "check":
			printStatus(cfg)
			return
		case "start", "r":
			fmt.Println("[*] Booting Wolarc AI node stack (wola.r)...")
			if err := system.StartServices(); err != nil {
				fmt.Printf("Error starting services: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("✓ Wolarc stack boot initiated. Run 'wola check' or 'wola' to monitor.")
			return
		case "stop", "k":
			fmt.Println("[*] Stopping Wolarc services (wola.k)...")
			if err := system.StopServices(); err != nil {
				fmt.Printf("Error stopping services: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("✓ All Wolarc services stopped cleanly.")
			return
		case "restart", "rs":
			fmt.Println("[*] Restarting Wolarc stack (wola.rs)...")
			if err := system.RestartServices(); err != nil {
				fmt.Printf("Error restarting services: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("✓ Wolarc restart initiated.")
			return
		case "link", "url":
			if len(os.Args) > 2 {
				cmd := exec.Command("wola.url", os.Args[2:]...)
				cmd.Stdin = os.Stdin
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				_ = cmd.Run()
				return
			}
			tURL := config.GetTunnelURL()
			if tURL != "" {
				fmt.Printf("Gateway Tunnel URL : %s\n", tURL)
				fmt.Printf("OpenAI API Endpoint: %s/v1\n", tURL)
			} else {
				fmt.Println("Gateway Tunnel is currently inactive. Run 'wola start' to start services.")
			}
			return
		case "hotspot", "local":
			cmd := exec.Command("wola.url", append([]string{"local"}, os.Args[2:]...)...)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			_ = cmd.Run()
			return
		case "ts", "tailscale":
			cmd := exec.Command("wola.url", append([]string{"tailscale"}, os.Args[2:]...)...)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			_ = cmd.Run()
			return
		case "ui", "tui":
			if len(os.Args) > 2 {
				sub := strings.ToLower(os.Args[2])
				switch sub {
				case "on", "enable", "1", "true":
					if err := config.SetConfigKey("ENABLE_TUI", "true"); err != nil {
						fmt.Printf("Error updating configuration: %v\n", err)
						os.Exit(1)
					}
					fmt.Println("✓ Go Bubble Tea TUI ENABLED. 'wola' will launch the interactive TUI.")
					return
				case "off", "disable", "0", "false":
					if err := config.SetConfigKey("ENABLE_TUI", "false"); err != nil {
						fmt.Printf("Error updating configuration: %v\n", err)
						os.Exit(1)
					}
					fmt.Println("✓ Go Bubble Tea TUI DISABLED. 'wola' will use the classic CLI command suite.")
					return
				case "status":
					statusStr := "enabled (true)"
					if !cfg.EnableTUI {
						statusStr = "disabled (false)"
					}
					fmt.Printf("Go Bubble Tea TUI Status: %s\n", statusStr)
					fmt.Println("Usage: wola ui [on|off|status] (or wola tui [on|off|status])")
					return
				}
			}
			statusStr := "enabled (true)"
			if !cfg.EnableTUI {
				statusStr = "disabled (false)"
			}
			fmt.Printf("Go Bubble Tea TUI Status: %s\n", statusStr)
			fmt.Println("Usage: wola ui [on|off|status] (or wola tui [on|off|status])")
			return
		case "bench", "b":
			cmd := exec.Command("wola.bench", os.Args[2:]...)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			_ = cmd.Run()
			return
		case "think", "th":
			cmd := exec.Command("wola.th", os.Args[2:]...)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			_ = cmd.Run()
			return
		case "config", "t":
			cmd := exec.Command("wola.t", os.Args[2:]...)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			_ = cmd.Run()
			return
		case "store", "gt":
			cmd := exec.Command("wola.gt", os.Args[2:]...)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			_ = cmd.Run()
			return
		case "update", "u":
			cmd := exec.Command("wola.u", os.Args[2:]...)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			_ = cmd.Run()
			return
		case "uninstall", "purge":
			cmd := exec.Command("wola.purge", os.Args[2:]...)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			_ = cmd.Run()
			return
		case "dashboard", "d":
			if len(os.Args) > 2 || (!cfg.EnableTUI && !strings.HasSuffix(os.Args[0], "wola-tui")) {
				cmd := exec.Command("wola.d", os.Args[2:]...)
				cmd.Stdin = os.Stdin
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				_ = cmd.Run()
				return
			}
			initialTab = ui.TabDashboard
		case "chat", "c":
			if len(os.Args) > 2 || (!cfg.EnableTUI && !strings.HasSuffix(os.Args[0], "wola-tui")) {
				cmd := exec.Command("wola.chat", os.Args[2:]...)
				cmd.Stdin = os.Stdin
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				_ = cmd.Run()
				return
			}
			initialTab = ui.TabChat
		case "models", "m":
			if len(os.Args) > 2 || (!cfg.EnableTUI && !strings.HasSuffix(os.Args[0], "wola-tui")) {
				cmd := exec.Command("wola.m", os.Args[2:]...)
				cmd.Stdin = os.Stdin
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				_ = cmd.Run()
				return
			}
			initialTab = ui.TabModels
		case "knowledge", "i":
			if len(os.Args) > 2 || (!cfg.EnableTUI && !strings.HasSuffix(os.Args[0], "wola-tui")) {
				cmd := exec.Command("wola.i", os.Args[2:]...)
				cmd.Stdin = os.Stdin
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				_ = cmd.Run()
				return
			}
			initialTab = ui.TabKnowledge
		case "log", "l":
			if len(os.Args) > 2 || (!cfg.EnableTUI && !strings.HasSuffix(os.Args[0], "wola-tui")) {
				cmd := exec.Command("wola.log", os.Args[2:]...)
				cmd.Stdin = os.Stdin
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				_ = cmd.Run()
				return
			}
			initialTab = ui.TabLogs
		default:
			if !cfg.EnableTUI && !strings.HasSuffix(os.Args[0], "wola-tui") {
				cmd := exec.Command("wola-h", os.Args[1:]...)
				cmd.Stdin = os.Stdin
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				_ = cmd.Run()
				return
			}
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

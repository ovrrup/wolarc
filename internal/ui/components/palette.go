package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ovrrup/wolarc/internal/ui/styles"
)

type PaletteActionID string

const (
	ActionStartStack   PaletteActionID = "start_stack"
	ActionStopStack    PaletteActionID = "stop_stack"
	ActionRestartStack PaletteActionID = "restart_stack"
	ActionCopyURL      PaletteActionID = "copy_url"
	ActionLocalGateway PaletteActionID = "local_gateway"
	ActionRunIndexer   PaletteActionID = "run_indexer"
	ActionRunBench     PaletteActionID = "run_bench"
	ActionGotoDash     PaletteActionID = "goto_dash"
	ActionGotoChat     PaletteActionID = "goto_chat"
	ActionGotoModels   PaletteActionID = "goto_models"
	ActionGotoKnow     PaletteActionID = "goto_know"
	ActionGotoLogs     PaletteActionID = "goto_logs"
	ActionReloadConfig PaletteActionID = "reload_config"
	ActionQuitApp      PaletteActionID = "quit_app"
)

type PaletteItem struct {
	ID          PaletteActionID
	Title       string
	Description string
	Category    string
}

type PaletteModel struct {
	Active       bool
	FilterInput  textinput.Model
	AllItems     []PaletteItem
	Filtered     []PaletteItem
	SelectedIdx  int
	Width        int
	Height       int
}

func NewPaletteModel() PaletteModel {
	ti := textinput.New()
	ti.Placeholder = "Type a command or jump target (e.g. start, chat, tunnel)..."
	ti.CharLimit = 64
	ti.Width = 52

	items := []PaletteItem{
		{ID: ActionStartStack, Title: "Start Wola Stack", Description: "Boot chat, embed, and agent servers (wola.r)", Category: "Lifecycle"},
		{ID: ActionStopStack, Title: "Stop Wola Stack", Description: "Cleanly shut down daemons & wake-lock (wola.k)", Category: "Lifecycle"},
		{ID: ActionRestartStack, Title: "Restart Services", Description: "Reboot all engines with fresh config (wola.rs)", Category: "Lifecycle"},
		{ID: ActionCopyURL, Title: "Copy Gateway Tunnel URL", Description: "Copy live HTTPS endpoint to clipboard", Category: "Networking"},
		{ID: ActionLocalGateway, Title: "Show Local / Hotspot Gateway", Description: "Display offline IP, Web UI link & QR code (wola local)", Category: "Networking"},
		{ID: ActionRunIndexer, Title: "Trigger Note Indexer", Description: "Run background document vectorization pass (wola.i)", Category: "Knowledge"},
		{ID: ActionRunBench, Title: "Run Inference Benchmark", Description: "Measure prompt eval & generation tok/s (wola.bench)", Category: "Hardware"},
		{ID: ActionGotoDash, Title: "View: Dashboard", Description: "Switch to telemetry & hardware monitor", Category: "Navigation"},
		{ID: ActionGotoChat, Title: "View: Interactive Chat", Description: "Switch to autonomous RAG chat session", Category: "Navigation"},
		{ID: ActionGotoModels, Title: "View: Models & Storage", Description: "Manage GGUF models & import from storage", Category: "Navigation"},
		{ID: ActionGotoKnow, Title: "View: Vector Knowledge", Description: "Inspect vector database & test retrieval", Category: "Navigation"},
		{ID: ActionGotoLogs, Title: "View: Realtime Logs", Description: "Follow live multi-stream service logs", Category: "Navigation"},
		{ID: ActionReloadConfig, Title: "Reload Configuration", Description: "Re-read ~/.wolarc variables into memory", Category: "Config"},
		{ID: ActionQuitApp, Title: "Quit Wolarc TUI", Description: "Exit application back to shell", Category: "Application"},
	}

	p := PaletteModel{
		FilterInput: ti,
		AllItems:    items,
		Filtered:    items,
		SelectedIdx: 0,
		Width:       60,
		Height:      16,
	}
	return p
}

func (p *PaletteModel) Open() {
	p.Active = true
	p.FilterInput.Reset()
	p.FilterInput.Focus()
	p.filterItems("")
	p.SelectedIdx = 0
}

func (p *PaletteModel) Close() {
	p.Active = false
	p.FilterInput.Blur()
}

func (p *PaletteModel) filterItems(query string) {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		p.Filtered = p.AllItems
		p.SelectedIdx = 0
		return
	}

	var res []PaletteItem
	for _, it := range p.AllItems {
		matchTitle := strings.Contains(strings.ToLower(it.Title), q)
		matchDesc := strings.Contains(strings.ToLower(it.Description), q)
		matchCat := strings.Contains(strings.ToLower(it.Category), q)
		if matchTitle || matchDesc || matchCat {
			res = append(res, it)
		}
	}
	p.Filtered = res
	if p.SelectedIdx >= len(p.Filtered) {
		p.SelectedIdx = 0
	}
}

func (p PaletteModel) Update(msg tea.Msg) (PaletteModel, tea.Cmd, *PaletteActionID) {
	if !p.Active {
		return p, nil, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			p.Close()
			return p, nil, nil
		case "up", "ctrl+p":
			if p.SelectedIdx > 0 {
				p.SelectedIdx--
			}
			return p, nil, nil
		case "down", "ctrl+n":
			if p.SelectedIdx < len(p.Filtered)-1 {
				p.SelectedIdx++
			}
			return p, nil, nil
		case "enter":
			if len(p.Filtered) > 0 && p.SelectedIdx < len(p.Filtered) {
				action := p.Filtered[p.SelectedIdx].ID
				p.Close()
				return p, nil, &action
			}
			p.Close()
			return p, nil, nil
		}
	}

	var cmd tea.Cmd
	prevVal := p.FilterInput.Value()
	p.FilterInput, cmd = p.FilterInput.Update(msg)
	if p.FilterInput.Value() != prevVal {
		p.filterItems(p.FilterInput.Value())
	}

	return p, cmd, nil
}

func (p PaletteModel) View(screenWidth, screenHeight int) string {
	if !p.Active {
		return ""
	}

	boxWidth := 64
	if screenWidth < 70 {
		boxWidth = screenWidth - 6
	}

	var listSb strings.Builder
	maxShow := 7
	start := 0
	if p.SelectedIdx >= maxShow {
		start = p.SelectedIdx - maxShow + 1
	}
	end := start + maxShow
	if end > len(p.Filtered) {
		end = len(p.Filtered)
	}

	if len(p.Filtered) == 0 {
		listSb.WriteString(styles.DescStyle.Render("  No matching commands found.\n"))
	} else {
		for i := start; i < end; i++ {
			it := p.Filtered[i]
			cursor := "  "
			titleStyle := styles.DescStyle
			categoryBadge := lipgloss.NewStyle().Foreground(styles.ColorDim).Render("[" + it.Category + "]")

			if i == p.SelectedIdx {
				cursor = styles.KeyStyle.Render("▶ ")
				titleStyle = lipgloss.NewStyle().Bold(true).Foreground(styles.ColorCyan)
				categoryBadge = lipgloss.NewStyle().Foreground(styles.ColorCyan).Bold(true).Render("[" + it.Category + "]")
			}

			line := fmt.Sprintf("%s%-34s %s\n    %s\n",
				cursor,
				titleStyle.Render(it.Title),
				categoryBadge,
				styles.DescStyle.Render(it.Description),
			)
			listSb.WriteString(line)
		}
	}

	hints := styles.DescStyle.Render("↑/↓: Navigate  •  Enter: Execute  •  Esc: Dismiss")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		styles.ModalHeaderStyle.Render("⚡ QUICK COMMAND PALETTE"),
		p.FilterInput.View(),
		listSb.String(),
		hints,
	)

	dialog := styles.ModalBoxStyle.Width(boxWidth).Render(content)
	return lipgloss.Place(screenWidth, screenHeight, lipgloss.Center, lipgloss.Center, dialog)
}

package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/ovrrup/wolarc/internal/ui/styles"
)

type HelpModalModel struct {
	Active bool
}

func NewHelpModalModel() HelpModalModel {
	return HelpModalModel{Active: false}
}

func (h *HelpModalModel) Toggle() {
	h.Active = !h.Active
}

func (h HelpModalModel) View(screenWidth, screenHeight int) string {
	if !h.Active {
		return ""
	}

	boxWidth := 70
	if screenWidth < 76 {
		boxWidth = screenWidth - 6
	}

	renderRow := func(key, desc string) string {
		return fmt.Sprintf("  %-20s %s\n", styles.KeyStyle.Render(key), styles.DescStyle.Render(desc))
	}

	var sb strings.Builder
	sb.WriteString(styles.ModalHeaderStyle.Render("🐺 WOLARC TUI SHORTCUTS & USER GUIDE") + "\n\n")

	sb.WriteString(styles.SubheaderStyle.Render("GLOBAL NAVIGATION") + "\n")
	sb.WriteString(renderRow("Tab / Shift+Tab", "Cycle forward / backward across views"))
	sb.WriteString(renderRow("1 - 5 / F1 - F5", "Jump directly to [Dash, Chat, Models, RAG, Logs]"))
	sb.WriteString(renderRow("Ctrl+K / :", "Open Quick Command Palette"))
	sb.WriteString(renderRow("? / F1", "Toggle this Help Overlay"))
	sb.WriteString(renderRow("q / Ctrl+C", "Quit application cleanly"))
	sb.WriteString("\n")

	sb.WriteString(styles.SubheaderStyle.Render("VIEW-SPECIFIC ACTIONS") + "\n")
	sb.WriteString(renderRow("[Dash] s / r", "Start/Stop daemon stack  │  Restart node"))
	sb.WriteString(renderRow("[Chat] Enter", "Send message with automatic vector injection"))
	sb.WriteString(renderRow("[Chat] Ctrl+C", "Abort current token streaming generation"))
	sb.WriteString(renderRow("[Chat] Ctrl+L", "Clear chat history"))
	sb.WriteString(renderRow("[Models] c / e", "Activate selected model for Chat or Embedding"))
	sb.WriteString(renderRow("[Models] i / d", "Import local GGUF (Move/Copy)  │  Delete model"))
	sb.WriteString(renderRow("[Knowledge] /", "Focus semantic vector retrieval tester"))
	sb.WriteString(renderRow("[Knowledge] i", "Trigger immediate background indexer pass"))
	sb.WriteString(renderRow("[Logs] 1 - 6", "Switch between Chat, Embed, API, Tunnel, Indexer, Think"))
	sb.WriteString(renderRow("[Logs] f / c", "Toggle Auto-Follow  │  Clear current log file"))
	sb.WriteString("\n")

	sb.WriteString(styles.DescStyle.Render("Press Esc or ? to dismiss this overlay"))

	dialog := styles.ModalBoxStyle.Width(boxWidth).Render(sb.String())
	return lipgloss.Place(screenWidth, screenHeight, lipgloss.Center, lipgloss.Center, dialog)
}

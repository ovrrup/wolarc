package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/ovrrup/wolarc/internal/ui/styles"
)

func RenderHeader(width int, isOnline bool, uptime, tunnelURL string) string {
	logo := styles.TitleStyle.Render("🐺 WOLARC AI NODE")
	subtitle := styles.SubtitleStyle.Render("Autonomous Local RAG")

	var statusBadge string
	if isOnline {
		statusBadge = styles.BadgeOnline.Render("● ONLINE")
	} else {
		statusBadge = styles.BadgeOffline.Render("○ OFFLINE")
	}

	uptimeText := styles.DescStyle.Render(fmt.Sprintf("Uptime: %s", uptime))

	left := lipgloss.JoinHorizontal(lipgloss.Center, logo, "  ", subtitle)
	right := lipgloss.JoinHorizontal(lipgloss.Center, statusBadge, "  ", uptimeText)

	spaceWidth := width - lipgloss.Width(left) - lipgloss.Width(right) - 4
	if spaceWidth < 1 {
		spaceWidth = 1
	}
	spacer := strings.Repeat(" ", spaceWidth)

	topBar := lipgloss.JoinHorizontal(lipgloss.Center, left, spacer, right)

	// Tunnel gateway banner
	var tunnelBadge string
	if isOnline && strings.HasPrefix(tunnelURL, "https://") {
		tunnelBadge = fmt.Sprintf(" Gateway: %s", styles.KeyStyle.Render(tunnelURL))
	} else {
		tunnelBadge = styles.DescStyle.Render(" Gateway: Inactive (run wola.r or press 's' to boot)")
	}

	divider := lipgloss.NewStyle().
		Foreground(styles.ColorBorder).
		Render(strings.Repeat("─", max(20, width-2)))

	return lipgloss.JoinVertical(lipgloss.Left, topBar, tunnelBadge, divider)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

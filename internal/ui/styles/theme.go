package styles

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Palette
	ColorCyan    = lipgloss.Color("#00E5FF")
	ColorBlue    = lipgloss.Color("#38BDF8")
	ColorPurple  = lipgloss.Color("#C084FC")
	ColorGreen   = lipgloss.Color("#10B981")
	ColorYellow  = lipgloss.Color("#FBBF24")
	ColorRed     = lipgloss.Color("#F87171")
	ColorGray    = lipgloss.Color("#64748B")
	ColorDark    = lipgloss.Color("#0F172A")
	ColorSurface = lipgloss.Color("#1E293B")
	ColorBorder  = lipgloss.Color("#334155")
	ColorText    = lipgloss.Color("#F8FAFC")
	ColorSubtext = lipgloss.Color("#94A3B8")

	// Semantic Color Aliases
	ColorPrimary = ColorCyan
	ColorSuccess = ColorGreen
	ColorAccent  = ColorPurple
	ColorWarning = ColorYellow

	// Base Layout Styles
	DocStyle = lipgloss.NewStyle().
			Padding(0, 1)

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorCyan).
			MarginBottom(0)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(ColorSubtext).
			Italic(true)

	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorCyan)

	SubheaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorBlue)

	ValStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorText)

	ActionStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(0, 1)

	// Status Badges
	BadgeOnline = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorDark).
			Background(ColorGreen).
			Padding(0, 1)

	BadgeOffline = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorText).
			Background(ColorRed).
			Padding(0, 1)

	BadgeWarning = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorDark).
			Background(ColorYellow).
			Padding(0, 1)

	BadgeInfo = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorDark).
			Background(ColorCyan).
			Padding(0, 1)

	// Box / Card Panels
	CardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(0, 1).
			Background(ColorDark)

	CardActiveStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorCyan).
			Padding(0, 1).
			Background(ColorDark)

	CardHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorCyan).
			MarginBottom(1)

	// Tab Bar
	TabStyle = lipgloss.NewStyle().
			Foreground(ColorSubtext).
			Padding(0, 2).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(ColorBorder)

	ActiveTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorCyan).
			Padding(0, 2).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(ColorCyan)

	// Chat Specific Styles
	UserMessageStyle = lipgloss.NewStyle().
				Foreground(ColorCyan).
				Bold(true)

	AssistantMessageStyle = lipgloss.NewStyle().
				Foreground(ColorGreen).
				Bold(true)

	ThinkingBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorPurple).
				Foreground(ColorPurple).
				Padding(0, 1).
				MarginTop(1).
				MarginBottom(1)

	// Keybind Hints
	KeyStyle = lipgloss.NewStyle().
			Foreground(ColorYellow).
			Bold(true)

	DescStyle = lipgloss.NewStyle().
			Foreground(ColorSubtext)

	HelpBar = lipgloss.NewStyle().
		Foreground(ColorSubtext).
		MarginTop(1)
)

func Gauge(percent int, width int) string {
	if width < 4 {
		width = 10
	}
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	filled := (percent * width) / 100
	empty := width - filled

	fillColor := ColorGreen
	if percent >= 85 {
		fillColor = ColorRed
	} else if percent >= 65 {
		fillColor = ColorYellow
	}

	fillStyle := lipgloss.NewStyle().Foreground(fillColor)
	emptyStyle := lipgloss.NewStyle().Foreground(ColorBorder)

	return fillStyle.Render(repeat("█", filled)) + emptyStyle.Render(repeat("░", empty))
}

func repeat(s string, count int) string {
	res := ""
	for i := 0; i < count; i++ {
		res += s
	}
	return res
}

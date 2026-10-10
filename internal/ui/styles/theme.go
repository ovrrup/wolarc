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

	// Modal Overlay Styles
	ModalBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorCyan).
			Background(lipgloss.Color("#0b0f19")).
			Padding(1, 2)

	ModalHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorCyan).
				MarginBottom(1)

	// Toast Notification Styles
	ToastSuccessStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorGreen).
				Foreground(ColorGreen).
				Background(lipgloss.Color("#052e16")).
				Bold(true).
				Padding(0, 1)

	ToastInfoStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorCyan).
			Foreground(ColorCyan).
			Background(lipgloss.Color("#083344")).
			Bold(true).
			Padding(0, 1)

	ToastWarningStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorYellow).
				Foreground(ColorYellow).
				Background(lipgloss.Color("#422006")).
				Bold(true).
				Padding(0, 1)

	ToastErrorStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorRed).
			Foreground(ColorRed).
			Background(lipgloss.Color("#450a0a")).
			Bold(true).
			Padding(0, 1)

	// Command Palette Styles
	PaletteInputStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorCyan).
				Padding(0, 1).
				MarginBottom(1)

	PaletteSelectedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#0f172a")).
				Background(ColorCyan).
				Padding(0, 1)

	PaletteItemStyle = lipgloss.NewStyle().
				Foreground(ColorText).
				Padding(0, 1)
)

// Sparkline renders a high-resolution Unicode telemetry sparkline with adaptive color gradients
var sparklineGlyphs = []string{" ", "▂", "▃", "▄", "▅", "▆", "▇", "█"}

func Sparkline(values []float64, maxVal float64) string {
	if len(values) == 0 {
		return ""
	}
	if maxVal <= 0 {
		for _, v := range values {
			if v > maxVal {
				maxVal = v
			}
		}
	}
	if maxVal <= 0 {
		maxVal = 1.0
	}

	var res string
	for _, v := range values {
		ratio := v / maxVal
		if ratio < 0 {
			ratio = 0
		}
		if ratio > 1.0 {
			ratio = 1.0
		}

		idx := int(ratio * float64(len(sparklineGlyphs)-1))
		glyph := sparklineGlyphs[idx]

		var color lipgloss.TerminalColor
		if ratio > 0.8 {
			color = ColorCyan
		} else if ratio > 0.5 {
			color = ColorBlue
		} else if ratio > 0.2 {
			color = ColorGreen
		} else {
			color = ColorBorder
		}

		res += lipgloss.NewStyle().Foreground(color).Render(glyph)
	}
	return res
}

// PulseFrame returns a pulsing animated glyph based on animation tick index
func PulseFrame(tick int) (string, lipgloss.TerminalColor) {
	frames := []struct {
		glyph string
		color lipgloss.TerminalColor
	}{
		{"●", ColorGreen},
		{"◉", ColorGreen},
		{"◎", ColorCyan},
		{"○", ColorBlue},
		{"◎", ColorCyan},
		{"◉", ColorGreen},
	}
	idx := (tick / 2) % len(frames)
	return frames[idx].glyph, frames[idx].color
}

// GradientBar renders a smooth colored bar with percentage badge
func GradientBar(percent float64, width int) string {
	if width < 4 {
		width = 10
	}
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	filled := int((percent / 100.0) * float64(width))
	empty := width - filled

	var fillColor lipgloss.TerminalColor
	if percent >= 85 {
		fillColor = ColorRed
	} else if percent >= 65 {
		fillColor = ColorYellow
	} else if percent >= 35 {
		fillColor = ColorCyan
	} else {
		fillColor = ColorGreen
	}

	filledBar := lipgloss.NewStyle().Foreground(fillColor).Render(repeat("█", filled))
	emptyBar := lipgloss.NewStyle().Foreground(ColorBorder).Render(repeat("░", empty))
	return filledBar + emptyBar
}

func Gauge(percent int, width int) string {
	return GradientBar(float64(percent), width)
}

func repeat(s string, count int) string {
	res := ""
	for i := 0; i < count; i++ {
		res += s
	}
	return res
}


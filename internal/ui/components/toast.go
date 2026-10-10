package components

import (
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/ovrrup/wolarc/internal/ui/styles"
)

type ToastType int

const (
	ToastSuccess ToastType = iota
	ToastInfo
	ToastWarning
	ToastError
)

type Toast struct {
	Message   string
	Type      ToastType
	CreatedAt time.Time
	Duration  time.Duration
}

func NewToast(msg string, t ToastType) Toast {
	return Toast{
		Message:   msg,
		Type:      t,
		CreatedAt: time.Now(),
		Duration:  3500 * time.Millisecond,
	}
}

func (t *Toast) IsExpired() bool {
	return time.Since(t.CreatedAt) > t.Duration
}

func RenderToast(t *Toast, screenWidth int) string {
	if t == nil || t.IsExpired() {
		return ""
	}

	var style lipgloss.Style
	var prefix string

	switch t.Type {
	case ToastSuccess:
		style = styles.ToastSuccessStyle
		prefix = "✓ "
	case ToastInfo:
		style = styles.ToastInfoStyle
		prefix = "ℹ "
	case ToastWarning:
		style = styles.ToastWarningStyle
		prefix = "⚠ "
	case ToastError:
		style = styles.ToastErrorStyle
		prefix = "✕ "
	}

	content := style.Render(prefix + t.Message)
	toastWidth := lipgloss.Width(content)

	pad := screenWidth - toastWidth - 3
	if pad < 0 {
		pad = 0
	}

	return strings.Repeat(" ", pad) + content
}

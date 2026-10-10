package views

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ovrrup/wolarc/internal/config"
	"github.com/ovrrup/wolarc/internal/ui/styles"
)

type LocalModelItem struct {
	Name     string
	Path     string
	SizeMB   int64
	IsChat   bool
	IsEmbed  bool
	Quant    string
	Tier     string
}

type CuratedModel struct {
	Name        string
	Filename    string
	Repo        string
	Size        string
	Description string
}

type ModelsModel struct {
	Config        *config.Config
	LocalModels   []LocalModelItem
	SelectedIndex int
	Importing     bool
	ImportInput   textinput.Model
	Deleting      bool
	StatusMsg     string
	Width         int
	Height        int
}

var CuratedCatalog = []CuratedModel{
	{
		Name:        "Gemma 2 2B Instruct (Q4_K_M)",
		Filename:    "2b_it_v2.gguf",
		Repo:        "google/gemma-2-2b-it-GGUF",
		Size:        "1.6 GB",
		Description: "Google's ultra-fast mobile reasoning model",
	},
	{
		Name:        "Qwen 2.5 1.5B Instruct (Q4_K_M)",
		Filename:    "qwen2.5-1.5b-instruct-q4_k_m.gguf",
		Repo:        "Qwen/Qwen2.5-1.5B-Instruct-GGUF",
		Size:        "1.1 GB",
		Description: "High token velocity and strong multilingual logic",
	},
	{
		Name:        "Llama 3.2 1B Instruct (Q4_K_M)",
		Filename:    "Llama-3.2-1B-Instruct-Q4_K_M.gguf",
		Repo:        "bartowski/Llama-3.2-1B-Instruct-GGUF",
		Size:        "0.8 GB",
		Description: "Meta's ultra-compact low-RAM conversational model",
	},
	{
		Name:        "BGE-Small English v1.5 (Q4_K_M)",
		Filename:    "bge-small-en-v1.5-q4_k_m.gguf",
		Repo:        "bartowski/bge-small-en-v1.5-GGUF",
		Size:        "75 MB",
		Description: "Fast, accurate 384-dim semantic embedding model",
	},
	{
		Name:        "Nomic Embed Text v1.5 (Q4_K_M)",
		Filename:    "nomic-embed-text-v1.5.Q4_K_M.gguf",
		Repo:        "nomic-ai/nomic-embed-text-v1.5-GGUF",
		Size:        "85 MB",
		Description: "High-accuracy 8192 context vector encoder",
	},
}

func detectQuant(name string) string {
	upper := strings.ToUpper(name)
	quants := []string{"Q4_K_M", "Q4_K_S", "Q5_K_M", "Q5_K_S", "Q8_0", "Q4_0", "Q4_1", "Q2_K", "Q3_K_M", "FP16", "BF16"}
	for _, q := range quants {
		if strings.Contains(upper, q) {
			return q
		}
	}
	return "Standard"
}

func detectTier(sizeMB int64) string {
	if sizeMB < 1500 {
		return "🟢 Ultra-Lightweight (Optimal for 4GB devices)"
	} else if sizeMB < 3000 {
		return "🟡 Balanced (Optimal for 6GB-8GB devices)"
	}
	return "🔴 Heavyweight (Recommended for 12GB+ devices)"
}

func NewModelsModel(cfg *config.Config) ModelsModel {
	ti := textinput.New()
	ti.Placeholder = "Enter path to .gguf file or directory to import..."
	ti.CharLimit = 512
	ti.Width = 60

	m := ModelsModel{
		Config:      cfg,
		ImportInput: ti,
		Width:       80,
		Height:      24,
	}
	m.Refresh()
	return m
}

func (m *ModelsModel) Refresh() {
	var list []LocalModelItem
	_ = os.MkdirAll(m.Config.ModelsDir, 0755)

	entries, err := os.ReadDir(m.Config.ModelsDir)
	if err == nil {
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			lower := strings.ToLower(e.Name())
			if strings.HasSuffix(lower, ".gguf") {
				info, err := e.Info()
				var sizeMB int64
				if err == nil {
					sizeMB = info.Size() / (1024 * 1024)
				}
				p := filepath.Join(m.Config.ModelsDir, e.Name())
				isChat := (e.Name() == m.Config.ChatModel)
				isEmbed := (e.Name() == m.Config.EmbedModel)

				list = append(list, LocalModelItem{
					Name:    e.Name(),
					Path:    p,
					SizeMB:  sizeMB,
					IsChat:  isChat,
					IsEmbed: isEmbed,
					Quant:   detectQuant(e.Name()),
					Tier:    detectTier(sizeMB),
				})
			}
		}
	}
	m.LocalModels = list
	if m.SelectedIndex >= len(m.LocalModels) && len(m.LocalModels) > 0 {
		m.SelectedIndex = len(m.LocalModels) - 1
	}
}

func (m ModelsModel) Init() tea.Cmd {
	return nil
}

func (m *ModelsModel) Update(msg tea.Msg) (ModelsModel, tea.Cmd) {
	if m.Importing {
		var cmd tea.Cmd
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyEnter:
				path := strings.TrimSpace(m.ImportInput.Value())
				if path != "" {
					m.handleImport(path)
				}
				m.Importing = false
				m.ImportInput.Reset()
				return *m, nil
			case tea.KeyEsc:
				m.Importing = false
				m.ImportInput.Reset()
				return *m, nil
			}
		}
		m.ImportInput, cmd = m.ImportInput.Update(msg)
		return *m, cmd
	}

	if m.Deleting {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "y", "Y":
				if len(m.LocalModels) > 0 && m.SelectedIndex < len(m.LocalModels) {
					sel := m.LocalModels[m.SelectedIndex]
					_ = os.Remove(sel.Path)
					m.StatusMsg = fmt.Sprintf("Successfully deleted model: %s", sel.Name)
					m.Refresh()
				}
				m.Deleting = false
				return *m, nil
			case "n", "N", "esc":
				m.Deleting = false
				m.StatusMsg = "Deletion cancelled"
				return *m, nil
			}
		}
		return *m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.SelectedIndex > 0 {
				m.SelectedIndex--
			}
		case "down", "j":
			if m.SelectedIndex < len(m.LocalModels)-1 {
				m.SelectedIndex++
			}
		case "enter", "c":
			if len(m.LocalModels) > 0 && m.SelectedIndex < len(m.LocalModels) {
				sel := m.LocalModels[m.SelectedIndex]
				_ = config.SetConfigKey("CHAT_MODEL", sel.Name)
				m.Config.ChatModel = sel.Name
				m.StatusMsg = fmt.Sprintf("Activated Chat Engine: %s (restart wola.rs to apply)", sel.Name)
				m.Refresh()
			}
		case "e":
			if len(m.LocalModels) > 0 && m.SelectedIndex < len(m.LocalModels) {
				sel := m.LocalModels[m.SelectedIndex]
				_ = config.SetConfigKey("EMBED_MODEL", sel.Name)
				m.Config.EmbedModel = sel.Name
				m.StatusMsg = fmt.Sprintf("Activated Embedding Engine: %s (restart wola.rs to apply)", sel.Name)
				m.Refresh()
			}
		case "i":
			m.Importing = true
			m.ImportInput.Focus()
			m.StatusMsg = "Enter storage path to import local GGUF file or directory"
		case "d":
			if len(m.LocalModels) > 0 && m.SelectedIndex < len(m.LocalModels) {
				m.Deleting = true
			}
		case "r":
			m.Refresh()
			m.StatusMsg = "Refreshed model list"
		}
	}

	return *m, nil
}

func (m *ModelsModel) handleImport(srcPath string) {
	srcPath = config.ExpandHome(srcPath)
	info, err := os.Stat(srcPath)
	if err != nil {
		m.StatusMsg = fmt.Sprintf("Path does not exist: %s", srcPath)
		return
	}

	var filesToImport []string
	if info.IsDir() {
		entries, _ := os.ReadDir(srcPath)
		for _, e := range entries {
			if strings.HasSuffix(strings.ToLower(e.Name()), ".gguf") {
				filesToImport = append(filesToImport, filepath.Join(srcPath, e.Name()))
			}
		}
	} else {
		filesToImport = append(filesToImport, srcPath)
	}

	if len(filesToImport) == 0 {
		m.StatusMsg = fmt.Sprintf("No .gguf files found in %s", srcPath)
		return
	}

	count := 0
	for _, f := range filesToImport {
		dest := filepath.Join(m.Config.ModelsDir, filepath.Base(f))
		if f == dest {
			continue
		}
		if err := copyFile(f, dest); err == nil {
			count++
		}
	}

	m.StatusMsg = fmt.Sprintf("Successfully imported %d model(s) into ~/models!", count)
	m.Refresh()
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func (m ModelsModel) View() string {
	cardWidth := (m.Width - 6) / 2
	if cardWidth < 36 {
		cardWidth = 36
	}

	// 1. LOCAL INVENTORY CARD
	var inventorySb strings.Builder
	inventorySb.WriteString(styles.CardHeader.Render("📦 LOCAL GGUF REPOSITORY") + "\n\n")

	if len(m.LocalModels) == 0 {
		inventorySb.WriteString(styles.DescStyle.Render("No .gguf models found in ~/models.\nPress [i] to import a local model or download via wola.gt.\n\n"))
	} else {
		for i, mod := range m.LocalModels {
			cursor := "  "
			if i == m.SelectedIndex {
				cursor = styles.KeyStyle.Render("▶ ")
			}

			tag := "       "
			if mod.IsChat && mod.IsEmbed {
				tag = styles.BadgeInfo.Render("[CHAT+EMBD]")
			} else if mod.IsChat {
				tag = styles.BadgeOnline.Render("[CHAT]")
			} else if mod.IsEmbed {
				tag = styles.BadgeWarning.Render("[EMBD]")
			}

			nameStyle := styles.DescStyle
			if i == m.SelectedIndex {
				nameStyle = lipgloss.NewStyle().Bold(true).Foreground(styles.ColorCyan)
			}

			sizeStr := fmt.Sprintf("%d MB", mod.SizeMB)
			line := fmt.Sprintf("%s%-13s %-24s %7s\n",
				cursor,
				tag,
				nameStyle.Render(truncate(mod.Name, 24)),
				styles.DescStyle.Render(sizeStr),
			)
			inventorySb.WriteString(line)
		}
	}
	inventoryCard := styles.CardStyle.Width(cardWidth).Render(inventorySb.String())

	// 2. MODEL DETAIL INSPECTOR CARD
	var detailSb strings.Builder
	detailSb.WriteString(styles.CardHeader.Render("🔍 MODEL INSPECTOR & TELEMETRY") + "\n\n")

	if len(m.LocalModels) > 0 && m.SelectedIndex < len(m.LocalModels) {
		sel := m.LocalModels[m.SelectedIndex]
		ramReqMB := sel.SizeMB + 750
		ramBar := styles.GradientBar(float64(ramReqMB)/8192.0*100.0, 12)

		detailSb.WriteString(fmt.Sprintf(
			"• Model Name  : %s\n"+
				"• Quantization: %s\n"+
				"• Disk Space  : %s\n"+
				"• Est. RAM Req: %s  %s\n"+
				"• Device Tier : %s\n"+
				"• Active Role : %s\n\n"+
				"• Full Path   :\n  %s",
			styles.KeyStyle.Render(truncate(sel.Name, 32)),
			styles.ValStyle.Render(sel.Quant),
			styles.ValStyle.Render(fmt.Sprintf("%d MB (%.2f GB)", sel.SizeMB, float64(sel.SizeMB)/1024.0)),
			styles.KeyStyle.Render(fmt.Sprintf("~%d MB", ramReqMB)),
			ramBar,
			sel.Tier,
			getRoleBadge(sel.IsChat, sel.IsEmbed),
			styles.DescStyle.Render(sel.Path),
		))
	} else {
		detailSb.WriteString(styles.DescStyle.Render("Select a model on the left to view architecture parameters and memory footprint."))
	}
	detailCard := styles.CardStyle.Width(cardWidth).Render(detailSb.String())

	topRow := lipgloss.JoinHorizontal(lipgloss.Top, inventoryCard, "  ", detailCard)

	// 3. CURATED RECOMMENDATIONS BAR
	var catalogSb strings.Builder
	catalogSb.WriteString(styles.CardHeader.Render("🌟 CURATED MOBILE-TESTED CATALOG (Hugging Face)") + "\n")
	for _, c := range CuratedCatalog {
		catalogSb.WriteString(fmt.Sprintf(" • %-30s %s  %s\n",
			styles.KeyStyle.Render(c.Name),
			styles.BadgeInfo.Render(c.Size),
			styles.DescStyle.Render("— "+c.Description),
		))
	}
	catalogCard := styles.CardStyle.Width(m.Width - 4).Render(catalogSb.String())

	// 4. Modals & Import Bar
	var interactiveBar string
	if m.Deleting && len(m.LocalModels) > 0 {
		sel := m.LocalModels[m.SelectedIndex]
		interactiveBar = styles.ToastWarningStyle.Render(fmt.Sprintf(" ⚠ Are you sure you want to delete '%s'? [y: Yes / n: Cancel] ", sel.Name))
	} else if m.Importing {
		interactiveBar = styles.CardActiveStyle.Width(m.Width - 4).Render(
			"Import Path: " + m.ImportInput.View() + "\n" + styles.DescStyle.Render("Press Enter to copy into ~/models or Esc to abort"),
		)
	} else if m.StatusMsg != "" {
		interactiveBar = styles.DescStyle.Render("Notice: " + m.StatusMsg)
	}

	hints := fmt.Sprintf(
		" %s Set Chat  •  %s Set Embed  •  %s Import Local GGUF  •  %s Delete  •  %s Reload",
		styles.KeyStyle.Render("[Enter/c]"),
		styles.KeyStyle.Render("[e]"),
		styles.KeyStyle.Render("[i]"),
		styles.KeyStyle.Render("[d]"),
		styles.KeyStyle.Render("[r]"),
	)
	hintsBar := styles.ActionStyle.Render(hints)

	return lipgloss.JoinVertical(lipgloss.Left, topRow, catalogCard, interactiveBar, hintsBar)
}

func getRoleBadge(isChat, isEmbed bool) string {
	if isChat && isEmbed {
		return styles.BadgeInfo.Render("Active Chat + Active Embed")
	} else if isChat {
		return styles.BadgeOnline.Render("Active Chat Engine")
	} else if isEmbed {
		return styles.BadgeWarning.Render("Active Embed Engine")
	}
	return styles.DescStyle.Render("Standby (Inactive)")
}

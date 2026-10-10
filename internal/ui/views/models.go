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
}

type CuratedModel struct {
	Name        string
	Filename    string
	Repo        string
	Size        string
	Description string
}

type ModelsModel struct {
	Config       *config.Config
	LocalModels  []LocalModelItem
	SelectedIndex int
	Importing    bool
	ImportInput  textinput.Model
	ImportAction string // "choose", "move", "copy"
	StatusMsg    string
	Width        int
	Height       int
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

func NewModelsModel(cfg *config.Config) ModelsModel {
	ti := textinput.New()
	ti.Placeholder = "Enter path to .gguf file or directory..."
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
	files, err := os.ReadDir(m.Config.ModelsDir)
	if err == nil {
		for _, f := range files {
			if f.IsDir() || !strings.HasSuffix(strings.ToLower(f.Name()), ".gguf") {
				continue
			}
			info, err := f.Info()
			var sizeMB int64
			if err == nil {
				sizeMB = info.Size() / (1024 * 1024)
			}
			list = append(list, LocalModelItem{
				Name:    f.Name(),
				Path:    filepath.Join(m.Config.ModelsDir, f.Name()),
				SizeMB:  sizeMB,
				IsChat:  f.Name() == m.Config.ChatModel,
				IsEmbed: f.Name() == m.Config.EmbedModel,
			})
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
		case "c":
			if len(m.LocalModels) > 0 && m.SelectedIndex < len(m.LocalModels) {
				sel := m.LocalModels[m.SelectedIndex]
				_ = config.SetConfigKey("CHAT_MODEL", sel.Name)
				m.Config.ChatModel = sel.Name
				m.StatusMsg = fmt.Sprintf("Activated Chat LLM: %s (restart wola.rs to apply)", sel.Name)
				m.Refresh()
			}
		case "e":
			if len(m.LocalModels) > 0 && m.SelectedIndex < len(m.LocalModels) {
				sel := m.LocalModels[m.SelectedIndex]
				_ = config.SetConfigKey("EMBED_MODEL", sel.Name)
				m.Config.EmbedModel = sel.Name
				m.StatusMsg = fmt.Sprintf("Activated Embed Model: %s (restart wola.rs to apply)", sel.Name)
				m.Refresh()
			}
		case "i":
			m.Importing = true
			m.ImportInput.Focus()
			m.StatusMsg = "Enter storage path to import local GGUF file or directory"
		case "d":
			if len(m.LocalModels) > 0 && m.SelectedIndex < len(m.LocalModels) {
				sel := m.LocalModels[m.SelectedIndex]
				_ = os.Remove(sel.Path)
				m.StatusMsg = fmt.Sprintf("Deleted model: %s", sel.Name)
				m.Refresh()
			}
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
		// Copy file
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
	var sb strings.Builder

	cardWidth := m.Width - 4
	if cardWidth < 40 {
		cardWidth = 40
	}

	sb.WriteString(styles.CardHeader.Render("📦 LOCAL MODEL INVENTORY") + "\n\n")

	if len(m.LocalModels) == 0 {
		sb.WriteString(styles.DescStyle.Render("No .gguf models found in ~/models.\nPress [i] to import a local model or download using wola.gt.\n\n"))
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
			line := fmt.Sprintf("%s%-14s %-45s %10s", cursor, tag, nameStyle.Render(mod.Name), styles.DescStyle.Render(sizeStr))
			sb.WriteString(line + "\n")
		}
		sb.WriteString("\n")
	}

	// Curated catalog recommendations
	sb.WriteString(styles.CardHeader.Render("🌟 RECOMMENDED CURATED MODELS (Hugging Face)") + "\n\n")
	for _, c := range CuratedCatalog {
		sb.WriteString(fmt.Sprintf(" • %s  %s  %s\n   %s\n",
			styles.KeyStyle.Render(c.Name),
			styles.BadgeInfo.Render(c.Size),
			styles.DescStyle.Render("("+c.Repo+")"),
			styles.SubtitleStyle.Render(c.Description),
		))
	}

	if m.Importing {
		sb.WriteString("\n" + styles.KeyStyle.Render("Import Model Path: ") + m.ImportInput.View() + "\n")
	}

	if m.StatusMsg != "" {
		sb.WriteString("\n" + styles.BadgeWarning.Render("Notice:") + " " + m.StatusMsg + "\n")
	}

	inventoryCard := styles.CardStyle.Width(cardWidth).Render(sb.String())

	hints := fmt.Sprintf(
		" %s Set Chat  •  %s Set Embed  •  %s Import Local GGUF  •  %s Delete  •  %s Tab",
		styles.KeyStyle.Render("[c]"),
		styles.KeyStyle.Render("[e]"),
		styles.KeyStyle.Render("[i]"),
		styles.KeyStyle.Render("[d]"),
		styles.KeyStyle.Render("Tab"),
	)

	return lipgloss.JoinVertical(lipgloss.Left, inventoryCard, styles.DescStyle.Render(hints))
}

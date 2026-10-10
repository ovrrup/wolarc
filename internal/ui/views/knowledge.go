package views

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ovrrup/wolarc/internal/config"
	"github.com/ovrrup/wolarc/internal/ui/styles"
)

type ChunkRecord struct {
	ChunkID int       `json:"chunk_id"`
	Text    string    `json:"text"`
	Hash    string    `json:"hash"`
	Vector  []float64 `json:"vector"`
}

type FileRecord struct {
	Filename  string        `json:"filename"`
	Chunks    []ChunkRecord `json:"chunks"`
	UpdatedAt float64       `json:"updated_at"`
}

type DocItem struct {
	FilePath    string
	Filename    string
	ChunkCount  int
	LastUpdated time.Time
}

type SearchResultItem struct {
	Filename string
	ChunkID  int
	Score    float64
	Snippet  string
}

type KnowledgeModel struct {
	Config        *config.Config
	Docs          []DocItem
	TotalChunks   int
	VectorDim     int
	DBFileExists  bool
	DBFileSize    int64
	DBFileMtime   time.Time
	SearchInput   textinput.Model
	Spinner       spinner.Model
	IsSearching   bool
	SearchResults []SearchResultItem
	StatusMsg     string
	SelectedIndex int
	Width         int
	Height        int
}

func NewKnowledgeModel(cfg *config.Config) KnowledgeModel {
	ti := textinput.New()
	ti.Placeholder = "Type a search query to test RAG retrieval..."
	ti.CharLimit = 256
	ti.Width = 60

	sp := spinner.New()
	sp.Spinner = spinner.MiniDot
	sp.Style = lipgloss.NewStyle().Foreground(styles.ColorCyan)

	m := KnowledgeModel{
		Config:      cfg,
		SearchInput: ti,
		Spinner:     sp,
		Width:       80,
		Height:      24,
	}
	m.Refresh()
	return m
}

func (m *KnowledgeModel) Refresh() {
	m.Docs = []DocItem{}
	m.TotalChunks = 0
	m.VectorDim = 0
	m.DBFileExists = false
	m.DBFileSize = 0

	fi, err := os.Stat(m.Config.DBFile)
	if err != nil {
		m.StatusMsg = fmt.Sprintf("Knowledge DB file not found: %s", m.Config.DBFile)
		return
	}

	m.DBFileExists = true
	m.DBFileSize = fi.Size()
	m.DBFileMtime = fi.ModTime()

	raw, err := os.ReadFile(m.Config.DBFile)
	if err != nil {
		m.StatusMsg = fmt.Sprintf("Error reading DB: %v", err)
		return
	}

	var dbMap map[string]FileRecord
	if err := json.Unmarshal(raw, &dbMap); err != nil {
		m.StatusMsg = fmt.Sprintf("Error decoding DB JSON: %v", err)
		return
	}

	for fpath, rec := range dbMap {
		fname := rec.Filename
		if fname == "" {
			fname = filepath.Base(fpath)
		}
		cCount := len(rec.Chunks)
		m.TotalChunks += cCount

		if m.VectorDim == 0 && cCount > 0 && len(rec.Chunks[0].Vector) > 0 {
			m.VectorDim = len(rec.Chunks[0].Vector)
		}

		var updated time.Time
		if rec.UpdatedAt > 0 {
			updated = time.Unix(int64(rec.UpdatedAt), 0)
		}

		m.Docs = append(m.Docs, DocItem{
			FilePath:    fpath,
			Filename:    fname,
			ChunkCount:  cCount,
			LastUpdated: updated,
		})
	}

	sort.Slice(m.Docs, func(i, j int) bool {
		return m.Docs[i].Filename < m.Docs[j].Filename
	})

	m.StatusMsg = fmt.Sprintf("Loaded %d documents (%d total chunks)", len(m.Docs), m.TotalChunks)
}

func (m *KnowledgeModel) PerformSearch(query string) {
	if strings.TrimSpace(query) == "" {
		m.SearchResults = nil
		m.StatusMsg = "Please enter a valid search query"
		return
	}

	// 1. Fetch query vector from embedder
	embedURL := fmt.Sprintf("http://127.0.0.1:%s/embedding", m.Config.EmbedPort)
	reqBody, _ := json.Marshal(map[string]string{"content": query})
	client := http.Client{Timeout: 8 * time.Second}

	resp, err := client.Post(embedURL, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		m.StatusMsg = fmt.Sprintf("Embedder offline (%s): %v", embedURL, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		m.StatusMsg = fmt.Sprintf("Embedder returned HTTP %d", resp.StatusCode)
		return
	}

	var jsonResp interface{}
	if err := json.NewDecoder(resp.Body).Decode(&jsonResp); err != nil {
		m.StatusMsg = fmt.Sprintf("Failed to parse embedder response: %v", err)
		return
	}

	queryVec := extractVector(jsonResp)
	if len(queryVec) == 0 {
		m.StatusMsg = "Could not extract vector from embedder response"
		return
	}

	// 2. Read DB and compute cosine similarities
	raw, err := os.ReadFile(m.Config.DBFile)
	if err != nil {
		m.StatusMsg = fmt.Sprintf("Failed to read DB: %v", err)
		return
	}

	var dbMap map[string]FileRecord
	if err := json.Unmarshal(raw, &dbMap); err != nil {
		m.StatusMsg = fmt.Sprintf("Failed to decode DB: %v", err)
		return
	}

	var results []SearchResultItem
	for fpath, rec := range dbMap {
		fname := rec.Filename
		if fname == "" {
			fname = filepath.Base(fpath)
		}
		for _, ch := range rec.Chunks {
			if len(ch.Vector) == 0 {
				continue
			}
			score := cosineSimilarity(queryVec, ch.Vector)
			snippet := ch.Text
			if len(snippet) > 180 {
				snippet = snippet[:180] + "..."
			}
			// Clean newlines for snippet
			snippet = strings.ReplaceAll(snippet, "\n", " ")

			results = append(results, SearchResultItem{
				Filename: fname,
				ChunkID:  ch.ChunkID,
				Score:    score,
				Snippet:  snippet,
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > 4 {
		results = results[:4]
	}

	m.SearchResults = results
	m.StatusMsg = fmt.Sprintf("Found %d matching chunks for: '%s'", len(results), query)
}

func extractVector(data interface{}) []float64 {
	switch v := data.(type) {
	case []interface{}:
		if len(v) > 0 {
			if firstMap, ok := v[0].(map[string]interface{}); ok {
				return toFloatSlice(firstMap["embedding"])
			}
			return toFloatSlice(v)
		}
	case map[string]interface{}:
		if emb, ok := v["embedding"]; ok {
			return toFloatSlice(emb)
		}
		if dataArr, ok := v["data"].([]interface{}); ok && len(dataArr) > 0 {
			if firstMap, ok := dataArr[0].(map[string]interface{}); ok {
				return toFloatSlice(firstMap["embedding"])
			}
		}
	}
	return nil
}

func toFloatSlice(val interface{}) []float64 {
	arr, ok := val.([]interface{})
	if !ok {
		return nil
	}
	res := make([]float64, len(arr))
	for i, item := range arr {
		if f, ok := item.(float64); ok {
			res[i] = f
		}
	}
	return res
}

func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

func (m KnowledgeModel) Init() tea.Cmd {
	return m.Spinner.Tick
}

func (m KnowledgeModel) Update(msg tea.Msg) (KnowledgeModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case spinner.TickMsg:
		var spCmd tea.Cmd
		m.Spinner, spCmd = m.Spinner.Update(msg)
		return m, spCmd

	case tea.KeyMsg:
		if m.IsSearching {
			switch msg.String() {
			case "esc":
				m.IsSearching = false
				m.SearchInput.Blur()
				return m, nil
			case "enter":
				m.PerformSearch(m.SearchInput.Value())
				return m, nil
			default:
				m.SearchInput, cmd = m.SearchInput.Update(msg)
				return m, cmd
			}
		}

		switch msg.String() {
		case "/", "s":
			m.IsSearching = true
			m.SearchInput.Focus()
			return m, nil
		case "r":
			m.Refresh()
			return m, nil
		case "i":
			// Trigger indexer in background
			m.StatusMsg = "Triggering indexer pass..."
			go func() {
				cmd := exec.Command("wola.i", "--once")
				_ = cmd.Run()
			}()
			return m, nil
		case "up", "k":
			if m.SelectedIndex > 0 {
				m.SelectedIndex--
			}
		case "down", "j":
			if m.SelectedIndex < len(m.Docs)-1 {
				m.SelectedIndex++
			}
		}
	}

	return m, nil
}

func (m KnowledgeModel) View() string {
	cardWidth := (m.Width - 6) / 2
	if cardWidth < 34 {
		cardWidth = 34
	}

	// 1. STATS CARD
	dbStatusBadge := styles.BadgeOffline.Render("NOT FOUND")
	if m.DBFileExists {
		dbStatusBadge = styles.BadgeOnline.Render("ACTIVE")
	}

	sizeStr := fmt.Sprintf("%.2f KB", float64(m.DBFileSize)/1024.0)
	if m.DBFileSize > 1024*1024 {
		sizeStr = fmt.Sprintf("%.2f MB", float64(m.DBFileSize)/(1024.0*1024.0))
	}

	statsContent := fmt.Sprintf(
		"%s  %s\n\n"+
			"• Store File   : %s\n"+
			"• File Size    : %s\n"+
			"• Total Docs   : %s\n"+
			"• Total Chunks : %s\n"+
			"• Vector Dim   : %s\n"+
			"• Watch Folder : %s\n"+
			"• Chunk Size   : %d chars",
		styles.HeaderStyle.Render("VECTOR DATABASE"),
		dbStatusBadge,
		styles.ValStyle.Render(filepath.Base(m.Config.DBFile)),
		styles.KeyStyle.Render(sizeStr),
		styles.KeyStyle.Render(fmt.Sprintf("%d", len(m.Docs))),
		styles.KeyStyle.Render(fmt.Sprintf("%d", m.TotalChunks)),
		styles.KeyStyle.Render(fmt.Sprintf("%d-D", m.VectorDim)),
		styles.DescStyle.Render(m.Config.KnowledgeDir),
		m.Config.ChunkSize,
	)
	statsCard := styles.CardStyle.Width(cardWidth).Render(statsContent)

	// 2. DOCUMENT INVENTORY CARD
	var docList strings.Builder
	docList.WriteString(styles.HeaderStyle.Render("INDEXED DOCUMENTS") + "\n\n")

	if len(m.Docs) == 0 {
		docList.WriteString(styles.DescStyle.Render("No documents indexed yet.\nPlace .txt, .md, or .pdf files into:\n" + m.Config.KnowledgeDir + "\nthen press 'i' to index."))
	} else {
		maxItems := 8
		start := 0
		if m.SelectedIndex >= maxItems {
			start = m.SelectedIndex - maxItems + 1
		}
		end := start + maxItems
		if end > len(m.Docs) {
			end = len(m.Docs)
		}

		for i := start; i < end; i++ {
			doc := m.Docs[i]
			cursor := "  "
			itemStyle := lipgloss.NewStyle().Foreground(styles.ColorDim)
			if i == m.SelectedIndex {
				cursor = "❯ "
				itemStyle = lipgloss.NewStyle().Bold(true).Foreground(styles.ColorPrimary)
			}

			timeAgo := ""
			if !doc.LastUpdated.IsZero() {
				timeAgo = doc.LastUpdated.Format("Jan 02 15:04")
			}

			line := fmt.Sprintf("%s%-24s %s (%s)",
				cursor,
				itemStyle.Render(truncate(doc.Filename, 24)),
				styles.KeyStyle.Render(fmt.Sprintf("%3d chunks", doc.ChunkCount)),
				styles.DescStyle.Render(timeAgo),
			)
			docList.WriteString(line + "\n")
		}
		if len(m.Docs) > maxItems {
			docList.WriteString(styles.DescStyle.Render(fmt.Sprintf("\n  (Showing %d-%d of %d docs, scroll ↑/↓)", start+1, end, len(m.Docs))))
		}
	}

	docsCard := styles.CardStyle.Width(cardWidth).Render(docList.String())
	topRow := lipgloss.JoinHorizontal(lipgloss.Top, statsCard, "  ", docsCard)

	// 3. SEMANTIC SEARCH TESTER CARD
	var searchSection strings.Builder
	searchSection.WriteString(styles.HeaderStyle.Render("SEMANTIC RETRIEVAL TESTER") + "\n")
	searchSection.WriteString(styles.DescStyle.Render("Verify embedding accuracy and cosine similarity matching against local vector index") + "\n\n")

	searchSection.WriteString(m.SearchInput.View() + "\n\n")

	if len(m.SearchResults) > 0 {
		searchSection.WriteString(styles.SubheaderStyle.Render("Top Matched Chunks:") + "\n")
		for idx, res := range m.SearchResults {
			matchPercent := res.Score * 100.0
			matchBar := styles.GradientBar(matchPercent, 10)
			scoreBadge := styles.BadgeOnline.Render(fmt.Sprintf("%s %.1f%%", matchBar, matchPercent))
			if res.Score < 0.6 {
				scoreBadge = styles.BadgeOffline.Render(fmt.Sprintf("%s %.1f%%", matchBar, matchPercent))
			}

			searchSection.WriteString(fmt.Sprintf(
				"  %d. %s %s [chunk #%d]\n     %s\n\n",
				idx+1,
				scoreBadge,
				styles.KeyStyle.Render(res.Filename),
				res.ChunkID,
				styles.DescStyle.Render(res.Snippet),
			))
		}
	} else if m.IsSearching {
		searchSection.WriteString(styles.DescStyle.Render(m.Spinner.View() + " Type query and press Enter to execute semantic search..."))
	}

	searchCard := styles.CardStyle.Width(m.Width - 4).Render(searchSection.String())

	// Actions Footer
	var actions []string
	if m.IsSearching {
		actions = append(actions, styles.KeyStyle.Render("Enter")+": Search", styles.KeyStyle.Render("Esc")+": Cancel")
	} else {
		actions = append(actions,
			styles.KeyStyle.Render("/")+": Test Query",
			styles.KeyStyle.Render("i")+": Run Indexer",
			styles.KeyStyle.Render("r")+": Refresh DB",
			styles.KeyStyle.Render("↑/↓")+": Navigate Docs",
		)
	}

	statusBar := styles.DescStyle.Render(m.StatusMsg)
	actionBar := styles.ActionStyle.Render(strings.Join(actions, "  │  "))

	return lipgloss.JoinVertical(lipgloss.Left, topRow, searchCard, statusBar, actionBar)
}

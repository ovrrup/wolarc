package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	LlamaBin            string
	ModelsDir           string
	KnowledgeDir        string
	ChatsDir            string
	DBFile              string
	MetricsFile         string
	ChatModel           string
	EmbedModel          string
	ChatPort            string
	EmbedPort           string
	AgentPort           string
	ChatThreads         int
	EmbedThreads        int
	ChatCtx             int
	EmbedCtx            int
	ChatParallel        int
	FlashAttention      string
	MLock               bool
	TunnelProvider      string
	Temperature         float64
	SimilarityThreshold float64
	TopK                int
	SystemPrompt        string
	SourceDir           string
	ChunkSize           int
	EnableThinking      bool
	EnableTUI           bool
}

func ExpandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[2:])
		}
	} else if path == "~" {
		home, err := os.UserHomeDir()
		if err == nil {
			return home
		}
	}
	return os.ExpandEnv(path)
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func DefaultConfig() *Config {
	home, _ := os.UserHomeDir()
	defaultLlama := "llama-server"
	if lp, err := exec.LookPath("llama-server"); err == nil {
		defaultLlama = lp
	} else if p := filepath.Join(home, "llama.cpp", "build", "bin", "llama-server"); fileExists(p) {
		defaultLlama = p
	}
	wolaData := filepath.Join(home, ".wola", "data")
	wolaRun := filepath.Join(home, ".wola", "run")
	dbFile := filepath.Join(wolaData, "knowledge_db.json")
	if !fileExists(dbFile) && fileExists(filepath.Join(home, "knowledge_db.json")) {
		dbFile = filepath.Join(home, "knowledge_db.json")
	}
	metricsFile := filepath.Join(wolaRun, "wola_metrics.json")
	if !fileExists(metricsFile) && fileExists(filepath.Join(home, ".wola_metrics.json")) {
		metricsFile = filepath.Join(home, ".wola_metrics.json")
	}

	docBase := filepath.Join(home, "storage", "shared", "Documents")
	if !fileExists(docBase) {
		docBase = filepath.Join(home, "Documents")
	}
	kDir := filepath.Join(docBase, "wola", "knowledge")
	cDir := filepath.Join(docBase, "wola", "chats")
	if fileExists(filepath.Join(docBase, "agent_knowledge")) {
		kDir = filepath.Join(docBase, "agent_knowledge")
	}
	if fileExists(filepath.Join(docBase, "agent_chats")) {
		cDir = filepath.Join(docBase, "agent_chats")
	}

	return &Config{
		LlamaBin:            defaultLlama,
		ModelsDir:           filepath.Join(home, "models"),
		KnowledgeDir:        kDir,
		ChatsDir:            cDir,
		DBFile:              dbFile,
		MetricsFile:         metricsFile,
		ChatModel:           "",
		EmbedModel:          "",
		ChatPort:            "8080",
		EmbedPort:           "8081",
		AgentPort:           "5000",
		ChatThreads:         4,
		EmbedThreads:        2,
		ChatCtx:             4096,
		EmbedCtx:            2048,
		ChatParallel:        2,
		FlashAttention:      "auto",
		MLock:               false,
		TunnelProvider:      "auto",
		Temperature:         0.6,
		SimilarityThreshold: 0.50,
		TopK:                2,
		SystemPrompt:        "You are an intelligent, perceptive assistant. Answer using your own rich knowledge base, reasoning, and creativity. When personal background notes or memories are available, weave them naturally into your response to personalize it without robotic quotes or disclaimers.",
		SourceDir:           "",
		ChunkSize:           1200,
		EnableThinking:      true,
		EnableTUI:           true,
	}
}

func ConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".wolarc")
}

func LoadConfig() *Config {
	cfg := DefaultConfig()
	path := ConfigPath()

	file, err := os.Open(path)
	if err != nil {
		return cfg
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.Trim(strings.TrimSpace(parts[1]), "\"'")

		switch key {
		case "LLAMA_BIN":
			expanded := ExpandHome(val)
			if fileExists(expanded) {
				cfg.LlamaBin = expanded
			} else if lp, err := exec.LookPath("llama-server"); err == nil {
				cfg.LlamaBin = lp
			} else {
				cfg.LlamaBin = expanded
			}
		case "MODELS_DIR":
			cfg.ModelsDir = ExpandHome(val)
		case "KNOWLEDGE_DIR":
			cfg.KnowledgeDir = ExpandHome(val)
		case "CHATS_DIR":
			cfg.ChatsDir = ExpandHome(val)
		case "DB_FILE":
			cfg.DBFile = ExpandHome(val)
		case "METRICS_FILE":
			cfg.MetricsFile = ExpandHome(val)
		case "CHAT_MODEL":
			cfg.ChatModel = val
		case "EMBED_MODEL":
			cfg.EmbedModel = val
		case "CHAT_PORT":
			cfg.ChatPort = val
		case "EMBED_PORT":
			cfg.EmbedPort = val
		case "AGENT_PORT":
			cfg.AgentPort = val
		case "CHAT_THREADS":
			if n, err := strconv.Atoi(val); err == nil {
				cfg.ChatThreads = n
			}
		case "EMBED_THREADS":
			if n, err := strconv.Atoi(val); err == nil {
				cfg.EmbedThreads = n
			}
		case "CHAT_CTX":
			if n, err := strconv.Atoi(val); err == nil {
				cfg.ChatCtx = n
			}
		case "EMBED_CTX":
			if n, err := strconv.Atoi(val); err == nil {
				cfg.EmbedCtx = n
			}
		case "CHAT_PARALLEL":
			if n, err := strconv.Atoi(val); err == nil {
				cfg.ChatParallel = n
			}
		case "FLASH_ATTENTION":
			cfg.FlashAttention = val
		case "MLOCK":
			cfg.MLock = (strings.ToLower(val) == "true")
		case "TUNNEL_PROVIDER":
			cfg.TunnelProvider = val
		case "TEMPERATURE":
			if f, err := strconv.ParseFloat(val, 64); err == nil {
				cfg.Temperature = f
			}
		case "SIMILARITY_THRESHOLD":
			if f, err := strconv.ParseFloat(val, 64); err == nil {
				cfg.SimilarityThreshold = f
			}
		case "TOP_K":
			if n, err := strconv.Atoi(val); err == nil {
				cfg.TopK = n
			}
		case "SYSTEM_PROMPT":
			cfg.SystemPrompt = val
		case "WOLA_SOURCE_DIR":
			cfg.SourceDir = ExpandHome(val)
		case "CHUNK_SIZE":
			if n, err := strconv.Atoi(val); err == nil {
				cfg.ChunkSize = n
			}
		case "ENABLE_THINKING":
			valLower := strings.ToLower(val)
			cfg.EnableThinking = !(valLower == "false" || valLower == "0" || valLower == "off")
		case "ENABLE_TUI":
			valLower := strings.ToLower(val)
			cfg.EnableTUI = !(valLower == "false" || valLower == "0" || valLower == "off")
		}
	}

	return cfg
}

func SetConfigKey(key, value string) error {
	path := ConfigPath()
	var lines []string

	if data, err := os.ReadFile(path); err == nil {
		lines = strings.Split(string(data), "\n")
	}

	found := false
	prefix := key + "="
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, prefix) {
			lines[i] = fmt.Sprintf(`%s="%s"`, key, value)
			found = true
			break
		}
	}

	if !found {
		lines = append(lines, fmt.Sprintf(`%s="%s"`, key, value))
	}

	content := strings.Join(lines, "\n")
	return os.WriteFile(path, []byte(content), 0644)
}

func getNodeLockPath() string {
	home, _ := os.UserHomeDir()
	p := filepath.Join(home, ".wola", "run", "wola.lock")
	if fileExists(p) {
		return p
	}
	return filepath.Join(home, ".wola.lock")
}

func IsNodeRunning() bool {
	lockPath := getNodeLockPath()
	if _, err := os.Stat(lockPath); err == nil {
		return true
	}
	return false
}

func NodeUptime() string {
	lockPath := getNodeLockPath()
	info, err := os.Stat(lockPath)
	if err != nil {
		return "Offline"
	}
	diff := time.Since(info.ModTime())
	hours := int(diff.Hours())
	mins := int(diff.Minutes()) % 60
	secs := int(diff.Seconds()) % 60
	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, mins, secs)
	}
	return fmt.Sprintf("%dm %ds", mins, secs)
}

func GetTunnelURL() string {
	home, _ := os.UserHomeDir()
	for _, p := range []string{filepath.Join(home, ".wola", "run", "wola_url"), filepath.Join(home, ".wola_url")} {
		if data, err := os.ReadFile(p); err == nil {
			u := strings.TrimSpace(string(data))
			if strings.HasPrefix(u, "http") {
				return u
			}
		}
	}

	for _, p := range []string{filepath.Join(home, ".wola", "logs", "tunnel.log"), filepath.Join(home, "tunnel.log")} {
		if data, err := os.ReadFile(p); err == nil {
			text := string(data)
			lines := strings.Split(text, "\n")
			for i := len(lines) - 1; i >= 0; i-- {
				l := strings.TrimSpace(lines[i])
				if (strings.Contains(l, ".trycloudflare.com") || strings.Contains(l, ".lhr.life") || strings.Contains(l, ".serveo.net") || strings.Contains(l, ".pinggy.link")) && strings.Contains(l, "https://") {
					for _, word := range strings.Fields(l) {
						if strings.HasPrefix(word, "https://") && !strings.Contains(word, "admin") {
							return word
						}
					}
				}
			}
		}
	}
	return "No active tunnel"
}

type MetricsData struct {
	PromptTokens         int     `json:"prompt_tokens"`
	PredictedTokens      int     `json:"predicted_tokens"`
	PromptPerSecond      float64 `json:"prompt_per_second"`
	PredictedPerSecond   float64 `json:"predicted_per_second"`
	TotalPredictedTokens int     `json:"total_predicted_tokens"`
	TotalPromptTokens    int     `json:"total_prompt_tokens"`
	RequestsCompleted    int     `json:"requests_completed"`
}

func GetMetrics(cfg *Config) MetricsData {
	var m MetricsData
	targetFile := cfg.MetricsFile
	if !fileExists(targetFile) {
		home, _ := os.UserHomeDir()
		alt := filepath.Join(home, ".wola", "run", "wola_metrics.json")
		if fileExists(alt) {
			targetFile = alt
		} else {
			targetFile = filepath.Join(home, ".wola_metrics.json")
		}
	}
	if data, err := os.ReadFile(targetFile); err == nil {
		_ = json.Unmarshal(data, &m)
	}
	return m
}

# Go Bubble Tea TUI User Manual 🎛️

Wolarc includes an industry-standard Terminal User Interface (TUI) built in Go using the Charm libraries: **Bubble Tea**, **Lip Gloss**, and **Bubbles**.

---

## 🚀 Launching the Interface

Run the casual or technical command:
```bash
wola          # Unified launch command
# Or directly:
wola-tui
```

> [!TIP]
> **Switching Default UI Modes**:
> - `wola.ui on` (or `wola.tui on`): Running `wola` opens the full-screen Go TUI.
> - `wola.ui off` (or `wola.tui off`): Running `wola` uses the classic lightweight CLI suite.

---

## 🗂️ The 5 Navigation Tabs

Navigate between views using **Numbers `1`–`5`** or **`Tab` / `Shift+Tab`**:

```
[1: ⚡ Dashboard]  [2: 💬 Chat]  [3: 📦 Models]  [4: 📚 Knowledge]  [5: 📋 Logs]
```

### 1. `⚡ Dashboard` (Key: `1`)
The main operational cockpit showing live hardware statistics and daemon states:
- **Service Status Indicators**: Glowing animated indicators for Chat Server (8080), Embed Server (8081), Agent Gateway (5000), Indexer, and Remote Tunnel.
- **Hardware Telemetry**: CPU load bars, RAM usage (Model vs. System), battery percentage, and thermal temperatures.
- **Velocity Sparklines**: Real-time token generation throughput rendered as Unicode sparklines (` ▂▃▄▅▆▇█`).
- **Active Model Profile**: Displays the active chat model, context size, and quantization format.

### 2. `💬 Chat` (Key: `2`)
Interactive conversational interface with live streaming tokens:
- **Server-Sent Events (SSE) Streaming**: Smooth token-by-token rendering.
- **Collapsible Deep Thinking Blocks (`Ctrl+T`)**: When reasoning is enabled, the model's `<think>` cognitive traces are rendered in a dedicated collapsible container.
- **Generation Speedometer**: Live generation badges displaying token throughput and total tokens generated: `[⚡ 22.4 tok/s • 385 tokens]`.
- **RAG Injection Badges**: Visual indicator when personal knowledge is retrieved and injected into the turn.

### 3. `📦 Models` (Key: `3`)
Local GGUF model manager:
- **Model Inventory**: Scans `~/models` and displays file size, quantization format (e.g. `Q4_K_M`), and last modified date.
- **Role Assignment**: Press `Enter` or `Space` to assign a model as the active **Chat Model** or **Embedding Model**.
- **Local Importer**: Import local `.gguf` files directly from internal storage (`/sdcard/Download/`).

### 4. `📚 Knowledge` (Key: `4`)
Vector database inspector and retrieval test bench:
- **Vector Stats**: Displays total indexed documents, total chunks, database file size, and last sync timestamp.
- **Interactive Search Tester**: Type a query and hit Enter to test vector retrieval live. Displays matched text chunks alongside gradient similarity score bars (e.g. `[████████░░] 84%`).

### 5. `📋 Logs` (Key: `5`)
Multiplexed live log viewer:
- **Stream Selector**: Use `[` and `]` to switch between log streams:
  - `chat`: `llama-server` Chat LLM engine
  - `embed`: `llama-server` Embedding engine
  - `agent`: Agent API gateway
  - `tunnel`: Remote watchdog tunnel
  - `indexer`: Background file indexer
  - `thinking`: Model reasoning `<think>` stream
- **Auto-Scroll**: Automatically pins to the bottom as new lines arrive.

---

## ⌨️ Global Keyboard Shortcuts

| Shortcut | Action Description |
|:---|:---|
| `Tab` / `Shift+Tab` | Cycle forward / backward through tabs |
| `1`, `2`, `3`, `4`, `5` | Jump directly to Dashboard, Chat, Models, Knowledge, or Logs |
| `Ctrl+K` or `:` | Open the **Command Palette** |
| `?` or `F1` | Open the **Help Cheat Sheet Modal** |
| `Ctrl+T` | Toggle visibility of `<think>` reasoning blocks in Chat |
| `Ctrl+L` | Clear active chat history or log view |
| `Esc` | Close open modals or overlays |
| `q` or `Ctrl+C` | Gracefully exit the TUI |

---

## ⚡ Command Palette (`Ctrl+K` / `:`)

Press `Ctrl+K` from any screen to open the fuzzy action launcher:
- `start` / `stop` / `restart`: Manage background daemons without leaving the TUI.
- `copy url`: Copy the active remote tunnel URL to the clipboard.
- `thinking on` / `thinking off`: Toggle reasoning tokens dynamically.
- `bench`: Launch in-terminal hardware benchmark.

# Configuration Knobs & Environment 🎚️

Wolarc configuration is centralized in `~/.wolarc`. Parameters are cleanly divided into **Dynamic Runtime Knobs** (instant effect, no restart needed) and **Engine Boot Knobs** (applied on daemon reboot).

---

## ⚡ Tuning CLI: `wola.config` / `wola.t`

```bash
# Display entire active configuration
wola.config
wola.t

# View the value of a specific parameter
wola.t TEMPERATURE

# Update a parameter
wola.t TEMPERATURE 0.65
wola.t TOP_K 3
```

---

## 🚀 1. Dynamic Runtime Knobs (Instant Effect)

These parameters take effect immediately on your next query without restarting any daemons:

| Parameter | Default | Range / Values | Description |
|:---|:---|:---|:---|
| `TEMPERATURE` | `0.6` | `0.0` – `1.5` | Sampling randomness. Lower is more deterministic; higher is more creative. `0.6` is optimal for natural reasoning. |
| `SIMILARITY_THRESHOLD`| `0.50` | `0.00` – `1.00`| Minimum cosine similarity score required for a knowledge chunk to be injected. |
| `TOP_K` | `2` | `1` – `10` | Maximum number of top-matching knowledge chunks injected into the prompt context. |
| `SYSTEM_PROMPT` | *(Default)* | String | Custom system persona and overarching behavioral guidelines. |
| `ENABLE_THINKING` | `true` | `true` / `false` | Dynamically enables or suppresses cognitive reasoning `<think>` tokens. Set `false` for instant response speed. |
| `ENABLE_TUI` | `true` | `true` / `false` | Sets default interface mode for `wola` (`true`: Go TUI, `false`: classic script suite). |

---

## ⚙️ 2. Engine Boot Knobs (Requires `wola.restart`)

These parameters configure `llama-server` allocations and network listeners. Applied upon restarting daemons (`wola.restart` / `wola.rs`):

| Parameter | Default | Recommended | Description |
|:---|:---|:---|:---|
| `CHAT_MODEL` | *(Starter)* | `~/models/xxx.gguf` | Path to the active conversation LLM GGUF file. |
| `EMBED_MODEL` | *(Starter)* | `~/models/xxx.gguf` | Path to the active vector embedding model GGUF file. |
| `CHAT_THREADS` | `4` | Big CPU cores | Number of CPU execution threads dedicated to chat inference. |
| `EMBED_THREADS` | `2` | `2` | Number of CPU execution threads dedicated to vector embedding. |
| `CHAT_CTX` | `4096` | `2048` – `8192` | Total context window in tokens for chat generation. |
| `EMBED_CTX` | `2048` | `1024` – `2048` | Context window for vector embedding ingestion. |
| `CHAT_PORT` | `8080` | `8080` | Local HTTP port for the Chat LLM server. |
| `EMBED_PORT` | `8081` | `8081` | Local HTTP port for the Embedding server. |
| `AGENT_PORT` | `5000` | `5000` | Port for the Agent API Gateway and Standalone Web Chat UI. |
| `FLASH_ATTENTION` | `auto` | `auto` / `on` | Toggles Flash Attention in `llama-server`. |
| `MLOCK` | `false` | `false` | Locks model weights into physical RAM to prevent OS swapping (use caution on mobile). |
| `CHAT_PARALLEL` | `2` | `2` | Number of parallel evaluation slots in `llama-server`. |
| `TUNNEL_PROVIDER`| `auto` | `auto`, `cloudflare`, `localhost.run`, `serveo`, `pinggy` | Preferred remote tunnel provider. |

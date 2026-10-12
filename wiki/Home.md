# Welcome to the Wolarc Documentation Wiki 🐺⚡

> **Wolarc** is an autonomous, fully private, local RAG (Retrieval-Augmented Generation) AI node engineered for Android (Termux) and Linux edge environments. It orchestrates high-performance quantized GGUF models via `llama.cpp`, maintains an automated vector knowledge indexer over local files and conversations, exposes an OpenAI-compatible API gateway, provides an air-gapped zero-dependency Web UI, and establishes encrypted remote mesh and tunnel connectivity.

---

## 🧭 Wiki Navigation

Explore the comprehensive guides and technical documentation:

| Section | Description | Key Pages |
|:---|:---|:---|
| 🚀 **Getting Started** | Fast onboarding, prerequisites, and automated installation | [[Quick Start Guide\|Quick-Start-Guide]] • [[Installation on Termux & Linux\|Installation-on-Termux-&-Linux]] |
| 🧠 **Architecture & Engine** | System topology, dual-engine isolation, and RAG dataflow | [[Architecture & Topology\|Architecture-&-Topology]] • [[RAG & Vector Engine\|RAG-&-Vector-Engine]] • [[Daemon Lifecycle & Storage\|Daemon-Lifecycle-&-Storage]] |
| ⌨️ **Commands & Control** | Strict 1-to-1 Technical vs. Casual command pairing and TUI | [[Complete Command Reference\|Complete-Command-Reference]] • [[Go TUI User Manual\|Go-TUI-User-Manual]] • [[Configuration Knobs\|Configuration-Knobs]] |
| 🌐 **Networking & Access** | Remote pairing, Tailscale mesh VPN, and offline gateways | [[Networking & Remote Access\|Networking-&-Remote-Access]] • [[Tailscale Mesh VPN Guide\|Tailscale-Mesh-VPN-Guide]] • [[Offline Hotspot & LAN\|Offline-Hotspot-&-LAN]] |
| 🛠️ **Optimization & Support**| Hardware profiling, troubleshooting, and upcoming roadmap | [[Hardware Tuning & Benchmarks\|Hardware-Tuning-&-Benchmarks]] • [[Troubleshooting & FAQ\|Troubleshooting-&-FAQ]] • [[Roadmap & Future Milestones\|Roadmap-&-Future-Milestones]] |

---

## ⚡ System Architecture at a Glance

```
                      ┌──────────────────────────────────────────────┐
                      │              Remote Client Apps              │
                      │  (Chatbox, Obsidian, NextChat, OpenWebUI)    │
                      └──────────────────────┬───────────────────────┘
                                             │
                       ┌─────────────────────┼─────────────────────┐
                       │                     │                     │
                Tailscale Mesh VPN     Cloudflare / SSH      Local Network
                 (Encrypted P2P)       (Watchdog Tunnel)    (Hotspot / LAN)
                       │                     │                     │
                       └─────────────────────┼─────────────────────┘
                                             ▼
                               ┌───────────────────────────┐
                               │    Agent API Gateway      │
                               │  (agent.py / Port 5000)   │
                               └──────┬─────────────┬──────┘
                                      │             │
                Query Vector / Cosine │             │ Augmented Prompt
                                      ▼             ▼
  ┌────────────────────────┐   ┌──────────────┐   ┌──────────────┐
  │ Local Notes & Docs     │──>│ Embed Server │   │  Chat LLM    │
  │ (indexer.py background)│   │ (Port 8081)  │   │ (Port 8080)  │
  │ └─ SHA-256 Deduplication   │ (Port 8081)  │   │              │
  └────────────────────────┘   └──────────────┘   └──────────────┘
```

---

## 🌟 Core Architectural Highlights

### 1. Dual llama-server Engine Isolation
Unlike monolithic local AI setups that share a single model context for both chat generation and vectorization, Wolarc runs two dedicated instances of `llama-server`:
- **Chat Server (`127.0.0.1:8080`)**: Dedicated to text generation, multi-turn dialogue, tool calls, and cognitive reasoning (`<think>`). Configured with Flash Attention, multi-slot parallelism, and tunable context windows.
- **Embedding Server (`127.0.0.1:8081`)**: Dedicated lightweight embedding model (e.g. `bge-small-en-v1.5`) serving `/v1/embeddings` concurrently without locking or evicting the chat model's KV cache.

### 2. Universal Multi-Format RAG & Dialogue Memory Blending
The continuous background indexer (`indexer.py`) monitors both your knowledge folder (`~/Documents/wola/knowledge`) and conversation history (`~/Documents/wola/chats`):
- **Universal Formats**: Ingests `.pdf`, `.docx`, `.epub`, `.md`, `.txt`, `.mp3`, `.wav`, `.m4a`, `.mp4` video metadata, images with EXIF/OCR, and code.
- **Blended Cognitive Memory**: Retrieved context is infused as natural long-term memory rather than verbatim quoting. The model reasons natively with your personal knowledge without awkward robotic disclaimers.
- **Atomic Multi-Process Safety**: Employs `fcntl.flock` cross-process advisory locking to protect `knowledge_db.json` from race conditions during concurrent chat logging and document indexing.

### 3. Industry-Standard Go Bubble Tea TUI (`wola`)
Built in Go using Charm Bubble Tea, Lip Gloss, and Bubbles:
- **Dashboard**: Real-time CPU, RAM, battery, thermal temperature, and token velocity sparklines.
- **Chat**: Live SSE streaming with expandable `<think>` reasoning blocks (`Ctrl+T`) and token meters.
- **Models**: Inventory manager, active role switcher, and local file importer.
- **Knowledge**: Live vector browser with interactive semantic similarity tester.
- **Logs**: Multiplexed live tailer for Chat, Embed, Agent, Tunnel, Indexer, and Thinking streams.
- **Command Palette (`Ctrl+K` / `:`)**: Instant fuzzy action launcher and keyboard shortcuts.

### 4. Resilient Multi-Route Connectivity
- **Air-Gapped Offline / Hotspot**: Connect over local Wi-Fi, portable hotspot (`192.168.43.1`), or USB tethering with in-terminal QR codes and `/v1/local_info`.
- **Tailscale WireGuard Mesh VPN**: Peer-to-peer encrypted mesh traversal with MagicDNS, Serve, and Funnel.
- **Autonomous Tunnel Watchdog Daemon**: Background supervisor monitoring tunnel health every 20s, auto-cascading across Cloudflare Quick Tunnels, `localhost.run`, `serveo.net`, and `pinggy.io`.

### 5. Strict 1-to-1 Technical vs. Casual Command Model
Every action has exactly **one Casual Name** (e.g. `wola.start`, `wola.chat`, `wola.models`) and **one Technical Name** (e.g. `wola.r`, `wola.c`, `wola.m`). No confusing clutter or redundant aliases.

---

## ⚡ Quick Start (Termux & Linux)

Install with a single command:
```bash
curl -fsSL https://raw.githubusercontent.com/ovrrup/wolarc/main/install.sh | bash
```

Launch the interface:
```bash
wola          # Opens Go Bubble Tea TUI
# Or boot services in the background:
wola.start    # or wola.r
```

For complete instructions, see the [[Quick Start Guide\|Quick-Start-Guide]].

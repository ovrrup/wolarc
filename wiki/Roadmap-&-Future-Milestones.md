# Roadmap & Future Milestones 🗺️

Wolarc is evolving from a standalone terminal-centric node on Android and Linux into a fully integrated, distributed local AI ecosystem.

---

## 🚀 Upcoming Strategic Milestones

```
┌─────────────────────────────────────────────────────────────────────────┐
│               Milestone 2: Native Cross-Platform Flutter App            │
│        (Android APK • iOS • Linux Desktop • Windows • macOS Desktop)    │
│  • Streaming Chat & Thinking UI (<think>) • Mobile Knowledge Uploader   │
│  • Real-Time Telemetry Dashboard • One-Tap QR Pairing & Model Manager   │
└────────────────────────────────────┬────────────────────────────────────┘
                                     │
            ◄── Reliable & Resilient Connection Fabric ──►
                     (Milestone 1: The Wolarc Link)
                                     │
    ┌───────────────────────┬────────┴──────────────┬───────────────────┐
    ▼                       ▼                       ▼                   ▼
mDNS Auto-Discovery    Tailscale P2P WireGuard   Named Cloudflare   WebRTC
(Zero-Config LAN)      (Direct Mesh / STUN)      (Static Tunnels)   (DataChan)
    │                       │                       │                   │
    └───────────────────────┼───────────────────────┴───────────────────┘
                            ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                 Wolarc Autonomous Host Node (Termux / Linux)            │
│  • llama-server Chat Engine (8080)   • llama-server Embed Engine (8081) │
│  • Multi-Format Indexer (indexer.py) • Agent API Gateway (5000)         │
│  • Unified Storage (~/.wola/)        • Bubble Tea TUI / Classic CLI     │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Milestone 1: Reliable & Effective Client-Host Connection Fabric

### 1. The Challenge
Edge nodes running on Android or edge Linux machines face significant connectivity hurdles:
- Cellular IP churn when switching between Wi-Fi and mobile 4G/5G.
- Carrier-Grade NAT (CGNAT) preventing direct incoming connections.
- Ephemeral tunnel timeouts dropping long-running streaming inference requests.

### 2. Architectural Pillars
1. **Zero-Config Local Auto-Discovery (mDNS / Zeroconf)**:
   - Host broadcasts `_wolarc._tcp.local` on port 5000.
   - Client applications discover active Wolarc nodes on the local subnet without typing manual IP addresses.
2. **Resilient Multi-Transport Mesh with Auto-Failover**:
   - **Tier 1 (LAN)**: Sub-2ms low-latency connection for local Wi-Fi and offline Android hotspots.
   - **Tier 2 (Tailscale WireGuard Mesh)**: Direct peer-to-peer encrypted WireGuard tunnels across NATs via STUN/DERP.
   - **Tier 3 (Managed Cloudflare Named Tunnels)**: Authenticated persistent tunnels (`cloudflared`) with custom static hostnames.
   - **Tier 4 (WebRTC DataChannels)**: Peer-to-peer data channels bypassing CGNAT for direct client-host streaming.
3. **Resumable Token Streaming & Session Continuity**:
   - In-memory event sequence buffer (`Last-Event-ID`).
   - If a mobile client transitions from Wi-Fi to cellular data during active generation, it reconnects with its last event ID and resumes token streaming without re-running inference or dropping context.
4. **Cryptographic Out-of-Band Pairing Protocol**:
   - Ephemeral Ed25519 pairing key exchange rendered as an ASCII QR code or numeric PIN (`wola.pair`).
   - Scoped authorization tokens (`chat`, `knowledge`, `admin`) stored in `~/.wola/run/authorized_clients.json`.

---

## Milestone 2: Official Native Flutter Companion App (`wola-flutter`)

An official, high-performance native application engineered in Flutter (Dart) targeting Android (APK), iOS, Linux, Windows, and macOS:

### Core Capabilities:
1. **Dynamic Connection Hub**:
   - Integrated camera QR scanner for instant node pairing.
   - Auto-discovers local nodes via mDNS.
   - Multi-node manager: switch seamlessly between phone nodes, local workstations, and remote home servers.
2. **Deep Thinking & Streaming Conversational UI**:
   - Smooth SSE streaming tokens with real-time velocity metrics (`tok/s`).
   - Expandable, styled `<think>` cognitive reasoning blocks with duration counters and step logs.
   - Rich Markdown, syntax highlighting across 60+ programming languages, and LaTeX math.
3. **Universal Document & Media Uploader (RAG)**:
   - Native OS share sheet and document picker to upload PDFs, Word docs, notes, audio, and images directly into `~/Documents/wola/knowledge/`.
   - Real-time indexing progress and vector database inspection.
   - Response citation inspector: tap in-chat sources to view original document chunks and similarity match scores.
4. **Hardware Telemetry & Node Remote Control**:
   - Live gauges for CPU load, RAM allocation, battery levels, and thermal conditions.
   - Remote daemon management (Start, Stop, Restart) and live multiplexed log streams.
   - Remote model store browser with one-tap Hugging Face GGUF downloads and runtime parameter sliders.

---

## 📅 Phased Implementation Schedule

| Phase | Milestone | Focus Areas | Deliverables |
|:---|:---|:---|:---|
| **Phase 1** | **M1: Connection** | Protocol Specification & Discovery | • mDNS service broadcast daemon on host (`_wolarc._tcp.local`)<br>• Pairing protocol design & Ed25519 handshake<br>• Host endpoints: `/v1/pair`, `/v1/peers`, `/v1/stream_resume` |
| **Phase 2** | **M1: Connection** | Resilient Streaming & Named Tunnels | • Resumable SSE stream buffer with `Last-Event-ID`<br>• Keep-alive ping/pong heartbeat<br>• Persistent Cloudflare Named Tunnel integration (`wola.t TUNNEL_TOKEN`)<br>• `wola.pair` and `wola.conn` CLI tools |
| **Phase 3** | **M2: Flutter App** | Project Scaffold & Connection Hub | • Flutter project structure (`apps/wola_flutter`)<br>• Cyber-Wolf Dark Material 3 design system<br>• One-Tap QR Scanner & mDNS Auto-Discovery UI<br>• Multi-host manager & credentials vault |
| **Phase 4** | **M2: Flutter App** | Chat Experience & Thinking Engine | • SSE streaming chat interface with auto-scroll<br>• Expandable `<think>` cognitive reasoning block<br>• Markdown & code syntax highlighter with copy buttons<br>• Live token velocity speedometer |
| **Phase 5** | **M2: Flutter App** | RAG Management & Telemetry | • Document picker & remote knowledge base uploader<br>• Vector database inspector & retrieval test bench<br>• Live hardware telemetry gauges (CPU, RAM, Battery, Thermal)<br>• Daemon log follower view |
| **Phase 6** | **M2: Flutter App** | Model Store & Cross-Platform Release | • Remote GGUF model browser & downloader<br>• Runtime hyperparameter tuner<br>• Automated GitHub Actions CI/CD for Android APK, Linux, Windows, macOS builds |

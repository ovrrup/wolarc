# Wolarc Roadmap & Future Milestones 🐺⚡

> **Wolarc** is evolving from a standalone terminal-centric node on Android and Linux into a fully integrated, distributed local AI ecosystem. This roadmap documents our core upcoming engineering milestones, architectural designs, technical specifications, and delivery phases.

---

## Strategic Goals Overview

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                           THE WOLARC DISTRIBUTED ECOSYSTEM                      │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│   ┌─────────────────────────────────────────────────────────────────────────┐   │
│   │               Milestone 2: Native Cross-Platform Flutter App            │   │
│   │        (Android APK • iOS • Linux Desktop • Windows • macOS Desktop)    │   │
│   │  • Streaming Chat & Thinking UI (<think>) • Mobile Knowledge Uploader   │   │
│   │  • Real-Time Telemetry Dashboard • One-Tap QR Pairing & Model Manager   │   │
│   └────────────────────────────────────┬────────────────────────────────────┘   │
│                                        │                                        │
│               ◄── Reliable & Resilient Connection Fabric ──►                    │
│                        (Milestone 1: The Wolarc Link)                           │
│                                        │                                        │
│       ┌───────────────────────┬────────┴──────────────┬───────────────────┐     │
│       ▼                       ▼                       ▼                   ▼     │
│   mDNS Auto-Discovery    Tailscale P2P WireGuard   Named Cloudflare   WebRTC    │
│   (Zero-Config LAN)      (Direct Mesh / STUN)      (Static Tunnels)   (DataChan)│
│       │                       │                       │                   │     │
│       └───────────────────────┼───────────────────────┴───────────────────┘     │
│                               ▼                                                 │
│   ┌─────────────────────────────────────────────────────────────────────────┐   │
│   │                 Wolarc Autonomous Host Node (Termux / Linux)            │   │
│   │  • llama-server Chat Engine (8080)   • llama-server Embed Engine (8081) │   │
│   │  • Multi-Format Indexer (indexer.py) • Agent API Gateway (5000)         │   │
│   │  • Unified Storage (~/.wola/)        • Bubble Tea TUI / Classic CLI     │   │
│   └─────────────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

## Milestone 1: Reliable & Effective Client-Host Connection Fabric

### 1. Problem Statement & Challenges
Edge AI nodes running on Android (Termux) or Linux workstations face unique networking hurdles:
- **Cellular & Dynamic IP Churn**: Devices frequently transition between home Wi-Fi, office LANs, cellular data (4G/5G), and offline hotspots, causing IP addresses to change and breaking ongoing streaming requests.
- **Carrier-Grade NAT (CGNAT) & Firewalls**: Mobile networks and home ISPs do not provide public IPv4 addresses, preventing direct inbound connections without complex port forwarding.
- **Tunnel Instability & Relay Latency**: Free SSH relay services (e.g. `localhost.run`, `serveo.net`, ephemeral Quick Tunnels) suffer from rate limits, connection dropouts, and geographic latency hops.
- **Stream Interruptions During Inference**: Large language model inference generates tokens over seconds or minutes; an intermittent network blip currently aborts the HTTP SSE connection and discards generated tokens.
- **Manual Discovery Friction**: Users must manually read terminal outputs, copy URLs, or scan QR codes each time the host IP changes.

### 2. Architecture & Technical Solutions

#### A. Multi-Tier Transport Mesh with Autonomous Failover
Wolarc will establish a four-tier hierarchical transport layer that automatically negotiates the fastest, most secure, and most resilient route available:

```
                  ┌─────────────────────────────────────┐
                  │          Client Application         │
                  └──────────────────┬──────────────────┘
                                     │
                             Route Negotiation
                                     │
           ┌─────────────────────────┼─────────────────────────┐
           ▼                         ▼                         ▼
   [Tier 1: LAN / Wi-Fi]     [Tier 2: Tailscale P2P]   [Tier 3: Cloudflare Named]
   • mDNS: wola.local        • WireGuard direct mesh   • Persistent custom domain
   • Zero latency (<2ms)     • NAT traversal (STUN)    • Cloudflare edge TLS
   • Air-gapped offline      • Encrypted worldwide     • Global access
           │                         │                         │
           └─────────────────────────┼─────────────────────────┘
                                     │ (All fallback to)
                                     ▼
                        [Tier 4: WebRTC DataChannels]
                        • Direct peer-to-peer data pipe
                        • Immune to CGNAT via STUN/DERP
                        • Low-overhead bidirectional streams
```

1. **Tier 1: Zero-Config Local Discovery (mDNS / Zeroconf / Avahi)**:
   - Host daemon announces service via mDNS as `_wolarc._tcp.local` on port 5000.
   - Client apps automatically discover active Wolarc hosts on the same Wi-Fi or hotspot without manual IP entry.
   - Service record includes node hostname, API version, active model name, and device architecture.
2. **Tier 2: Direct Peer-to-Peer Encrypted Mesh (Tailscale WireGuard)**:
   - Deep integration with Tailscale MagicDNS (e.g., `http://wola-phone.tailnet:5000/v1`).
   - Direct end-to-end encrypted WireGuard tunnels traversing symmetric NATs via STUN/DERP.
   - Preserves static hostnames across Wi-Fi/cellular transitions with zero relay bottleneck.
3. **Tier 3: Managed Cloudflare Named Tunnels (`cloudflared`)**:
   - Upgrade from ephemeral Quick Tunnels (`trycloudflare.com`) to authenticated persistent Named Tunnels.
   - Allows users to bind their personal domain (e.g. `ai.myname.dev`) or free tunnel token with automatic DNS updates and WebSocket pass-through.
4. **Tier 4: WebRTC DataChannels (P2P Fallback)**:
   - WebRTC DataChannel connection layer for high-speed streaming even when both client and host are behind strict Carrier-Grade NATs.

#### B. Resumable Token Streaming & Session Continuity Engine
- **Resumable Server-Sent Events (SSE)**: Implement event sequence identifiers (`Last-Event-ID`) and an in-memory sliding token buffer on the host. If the client experiences a transient network drop (e.g., walking out of Wi-Fi range), it reconnects with `Last-Event-ID: <n>` and receives buffered tokens without restarting generation.
- **Adaptive Heartbeat & Keep-Alive**: Host and client exchange lightweight JSON heartbeats (`ping`/`pong`) every 15s with jittered backoff. Detects dead connections in <3s instead of waiting for TCP socket timeouts.
- **Stateful Request Journal**: Long-running generation jobs persist their progress in `~/.wola/run/active_jobs.json`. Clients can detach and reattach to existing generations at will.

#### C. Cryptographic Handshake & Zero-Trust Pairing Protocol
- **Out-of-Band Pairing**:
  - Host generates an ephemeral Ed25519 pairing key rendered as a cryptographically signed QR code or a 6-digit numeric OTP.
  - Client scans QR or enters PIN to complete a Diffie-Hellman key exchange.
  - Host issues a scoped, rotating access token (PASETO or HMAC-SHA256) stored in `~/.wola/run/authorized_clients.json`.
- **Role-Based Token Scopes**:
  - `chat`: Can query `/v1/chat/completions` and `/v1/models`.
  - `knowledge`: Can upload documents and query vector database.
  - `admin`: Can start/stop daemons, download models, and edit configuration parameters.
- **Automatic IP Whitelisting & Rate Limiting**: Token bucket rate limiter protects the edge node against brute force and resource exhaustion.

#### D. New Host CLI Tooling
- Technical Commands: `wola.pair`, `wola.peers`, `wola.conn`
- Casual Commands: `wola.pair`, `wola.clients`, `wola.connection`
- Features: List authorized clients, revoke client access tokens, display active transport routes, and benchmark connection latency.

---

## Milestone 2: Official Native Flutter Client App (`Wolarc App`)

### 1. Vision & Objectives
Provide an elegant, native, high-performance cross-platform application for Android, iOS, Linux, macOS, and Windows that unlocks the full power of the Wolarc node without relying on generic third-party chat software.

### 2. Core Functional Modules

#### A. Dynamic Connection Hub & Host Management
- **One-Tap QR Code Pairing**: Scan the host's terminal pairing code (`wola.pair` / `wola.link`) using the mobile camera for immediate zero-config setup.
- **Local Network Auto-Discovery**: Automatically detect Wolarc instances on the local network via mDNS without typing IP addresses.
- **Multi-Node Manager**: Save and switch between multiple Wolarc nodes (e.g., "Termux Phone", "Home Desktop", "Office Server").
- **Smart Transport Switching**: Visually displays current transport mode (`Local Wi-Fi • 3ms`, `Tailscale • 28ms`, or `Cloudflare • 85ms`) and automatically switches routes when the device changes networks.

#### B. Streaming Conversational Chat & Cognitive Reasoning
- **Live SSE Token Streaming**: Smooth token-by-token text streaming with zero UI stutter.
- **Native Deep Thinking UI (`<think>`)**:
  - Collapsible, styled reasoning container with expandable cognitive step metrics.
  - Real-time reasoning token stream viewer.
  - Thinking duration timer (`Thinking for 4.2s...`) and step count.
  - Dynamic thinking toggle switch in the chat input bar (`wola think on/off`).
- **Rich Markdown & Code Rendering**:
  - Syntax highlighting across 60+ programming languages.
  - One-tap code copy button and full-screen code viewer.
  - Mathematical formulas via LaTeX / KaTeX.
- **Token Velocity & Telemetry Badges**:
  - Real-time generation speedometer (`⚡ 22.4 tok/s • 412 tokens`).
  - Active model badge (e.g. `Gemma 2 2B Q4_K_M`).

#### C. Universal Knowledge Base & Document Explorer (RAG)
- **Direct File Ingestion**:
  - Share sheet and document picker to upload files (`.pdf`, `.docx`, `.epub`, `.md`, `.txt`, `.mp3`, `.jpg`) directly from phone storage into the host's knowledge repository (`~/Documents/wola/knowledge/`).
  - Instant background indexing progress indicator.
- **Vector Knowledge Inspector**:
  - Browse indexed documents, chunk counts, SHA-256 signatures, and file sizes.
  - Live vector similarity test bench: type a test query and inspect semantic match percentages with visual gradient score bars.
  - Delete or prune documents and vectors with one tap.
- **Chat Retrieval Source Inspector**:
  - Inspect the exact knowledge chunks and similarity scores injected into each LLM response.
  - Tap citations to view original source text in a dedicated sheet.

#### D. Live Hardware Telemetry & Remote Node Controls
- **Real-Time Telemetry Dashboard**:
  - CPU utilization gauges with thread breakdowns.
  - RAM allocation breakdown (Model RAM, System RAM, Free RAM).
  - Mobile battery percentage, charging state, and thermal temperature.
  - Interactive token throughput history sparklines.
- **Daemon Lifecycle Controls**:
  - Remote Start, Stop, and Restart node services (`wola.start`, `wola.stop`, `wola.restart`).
  - Real-time daemon log streamer with filter tabs (`Chat`, `Embed`, `Agent`, `Tunnel`, `Indexer`).

#### E. Remote Model Store & Tuning Center
- **Interactive Model Store**:
  - Browse verified GGUF models directly from Hugging Face.
  - Remote trigger download on host node with live progress bar and download speed.
  - Assign active Chat and Embedding model roles.
- **Visual Hyperparameter Tuner**:
  - Temperature slider (0.0 to 1.5).
  - Context window size selector (2048, 4096, 8192, 16384).
  - RAG Similarity Threshold & Top-K chunk injection sliders.
  - CPU thread count stepper.
  - System persona editor.

### 3. Technical Architecture & Tech Stack

```
┌────────────────────────────────────────────────────────────────────────┐
│                        FLUTTER CLIENT ARCHITECTURE                     │
├────────────────────────────────────────────────────────────────────────┤
│                                                                        │
│   ┌────────────────────────────────────────────────────────────────┐   │
│   │                         Presentation Layer                     │   │
│   │   • Material 3 / Cyber-Wolf Dark Theme (Custom Design System)  │   │
│   │   • Chat View • Knowledge Manager • Telemetry Dashboard        │   │
│   │   • Model Store • Host Switcher • Settings & Parameter Tuner   │   │
│   └───────────────────────────────┬────────────────────────────────┘   │
│                                   │                                    │
│   ┌───────────────────────────────▼────────────────────────────────┐   │
│   │                        State Management                        │   │
│   │   • Riverpod 2.x (AsyncNotifier, StateNotifier, StateProvider) │   │
│   │   • Streams for SSE Token Delivery & Telemetry WebSockets      │   │
│   └───────────────────────────────┬────────────────────────────────┘   │
│                                   │                                    │
│   ┌───────────────────────────────▼────────────────────────────────┐   │
│   │                        Repository Layer                        │   │
│   │   • ChatRepository • KnowledgeRepository • NodeConfigRepo      │   │
│   │   • TelemetryRepository • PeerConnectionRepository             │   │
│   └───────────────────────────────┬────────────────────────────────┘   │
│                                   │                                    │
│   ┌───────────────────────────────▼────────────────────────────────┐   │
│   │                    Network & Transport Engine                  │   │
│   │   • Dio HTTP / SSE Streaming Client (Resumable Streams)        │   │
│   │   • WebSocket Client (Live Telemetry & Logs)                   │   │
│   │   • mDNS Multicast Discovery Engine (bonsoir / multicast_dns)  │   │
│   │   • Mobile QR Camera Scanner (mobile_scanner)                  │   │
│   │   • WebRTC / WireGuard Transport Adapters                      │   │
│   └───────────────────────────────┬────────────────────────────────┘   │
│                                   │                                    │
│   ┌───────────────────────────────▼────────────────────────────────┐   │
│   │                      Local Persistence Layer                   │   │
│   │   • Isar / Hive Database (Offline Chat History & Cached Nodes) │   │
│   │   • Flutter Secure Storage (Cryptographic Pairing Tokens)      │   │
│   └────────────────────────────────────────────────────────────────┘   │
└────────────────────────────────────────────────────────────────────────┘
```

- **Framework**: Flutter 3.24+ (Dart 3.5+)
- **Target Platforms**: Android (APK / Google Play / F-Droid), iOS, Linux (.deb / AppImage), Windows (.msi / .zip), macOS (.dmg).
- **Core Dependencies**:
  - `flutter_riverpod`: Reactive, declarative state management.
  - `dio` & `flutter_client_sse`: Resumable HTTP and SSE token streaming.
  - `web_socket_channel`: Bidirectional telemetry and log streaming.
  - `bonsoir` / `nsd`: Zero-config mDNS service discovery.
  - `mobile_scanner`: High-performance camera QR pairing.
  - `flutter_markdown` & `flutter_highlight`: Markdown and syntax highlighting.
  - `fl_chart`: Responsive telemetry graphs and velocity sparklines.
  - `isar`: Blazing-fast embedded NoSQL database for local chat cache.
  - `flutter_secure_storage`: Keyring-backed storage for node pairing secrets.

---

## Phased Implementation Roadmap

| Phase | Milestone | Focus Areas | Deliverables |
|:---|:---|:---|:---|
| **Phase 1** | **M1: Connection** | Protocol Specification & Discovery | • mDNS service broadcast daemon on host (`_wolarc._tcp.local`)<br>• Pairing protocol design & Ed25519 handshake<br>• Host endpoints: `/v1/pair`, `/v1/peers`, `/v1/stream_resume` |
| **Phase 2** | **M1: Connection** | Resilient Streaming & Named Tunnels | • Resumable SSE stream buffer with `Last-Event-ID`<br>• Keep-alive ping/pong heartbeat<br>• Persistent Cloudflare Named Tunnel integration (`wola.t TUNNEL_TOKEN`)<br>• `wola.pair` and `wola.conn` CLI tools |
| **Phase 3** | **M2: Flutter App** | Project Scaffold & Connection Hub | • Flutter project structure (`apps/wola_flutter`)<br>• Cyber-Wolf Dark Material 3 design system<br>• One-Tap QR Scanner & mDNS Auto-Discovery UI<br>• Multi-host manager & credentials vault |
| **Phase 4** | **M2: Flutter App** | Chat Experience & Thinking Engine | • SSE streaming chat interface with auto-scroll<br>• Expandable `<think>` cognitive reasoning block<br>• Markdown & code syntax highlighter with copy buttons<br>• Live token velocity speedometer |
| **Phase 5** | **M2: Flutter App** | RAG Management & Telemetry | • Document picker & remote knowledge base uploader<br>• Vector database inspector & retrieval test bench<br>• Live hardware telemetry gauges (CPU, RAM, Battery, Thermal)<br>• Daemon log follower view |
| **Phase 6** | **M2: Flutter App** | Model Store & Cross-Platform Release | • Remote GGUF model browser & downloader<br>• Runtime hyperparameter tuner<br>• Automated GitHub Actions CI/CD for Android APK, Linux, Windows, macOS builds |

---

## Alignment with Wolarc Design Principles

1. **Strict 1-to-1 Technical vs. Casual Pairing**:
   - Connection features follow the pairing standard:
     - `wola.pair` (Technical) / `wola.pair` (Casual)
     - `wola.peers` (Technical) / `wola.clients` (Casual)
     - `wola.conn` (Technical) / `wola.connection` (Casual)
2. **100% Offline & Air-Gapped Capable**:
   - Both the connection method and Flutter app work flawlessly over local offline Wi-Fi, Android portable hotspots, and USB reverse tethering without requiring active internet connectivity.
3. **Zero Third-Party Cloud Telemetry**:
   - The Flutter app will communicate exclusively with the user's private Wolarc node without any third-party analytics, crash reporting, or external cloud dependencies.

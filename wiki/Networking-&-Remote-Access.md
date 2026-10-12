# Networking & Remote Access 🌐

Wolarc provides multiple secure, resilient networking pathways to connect client applications across local networks, air-gapped hotspots, and the internet.

---

## 🗺️ Multi-Route Connectivity Matrix

| Route Mode | Access Scope | Latency | Security | Command |
|:---|:---|:---|:---|:---|
| **Localhost** | Same device | <0.5 ms | Device sandboxed | `http://127.0.0.1:5000` |
| **Wi-Fi LAN** | Same local network | 1–3 ms | Local subnet | `wola.hotspot` / `wola.local` |
| **Mobile AP Hotspot** | Air-gapped offline | 2–5 ms | WPA2/WPA3 Hotspot | `wola.hotspot` (`192.168.43.1`) |
| **Tailscale Mesh VPN** | Worldwide peers | 15–40 ms | WireGuard E2EE | `wola.tailscale` / `wola.ts` |
| **Public Watchdog Tunnel**| Internet | 50–120 ms | HTTPS TLS | `wola.link` / `wola.url` |

---

## 🛡️ Autonomous Tunnel Watchdog Daemon

Remote access over public networks is supervised by the background watchdog daemon (`bin/wola-tunnel-daemon`):

```
┌─────────────────────────────────────────────────────────────┐
│                 TUNNEL WATCHDOG SUPERVISOR                  │
├─────────────────────────────────────────────────────────────┤
│ 1. Every 20s: Probes active tunnel URL with HTTP GET.       │
│ 2. If endpoint responds 200/404 -> HEALTHY (sleep 20s).     │
│ 3. If endpoint fails -> Checks WAN connectivity first.      │
│    ├─ If offline -> Pauses cascade (prevents battery drain) │
│    └─ If online -> Initiates provider fallback cascade.     │
│ 4. Updates ~/.wola/run/wola_tunnel_url with new URL.        │
└─────────────────────────────────────────────────────────────┘
```

### Supported Providers & Fallback Priority:
1. **Cloudflare Quick Tunnels** (`trycloudflare.com`): High speed, global CDN edge, zero account required.
2. **`localhost.run`**: Direct SSH reverse tunnel with automatic HTTPS.
3. **`serveo.net`**: Robust SSH forwarding fallback.
4. **`pinggy.io`**: Lightweight alternative SSH tunnel.

Switch providers manually:
```bash
wola.url switch cloudflare
wola.url switch localhost.run
wola.url switch serveo
wola.url switch pinggy
```

---

## ⚡ Connection Latency Benchmark (`wola.link test`)

Run a real-time latency sweep across all routes:
```bash
wola.link test        # Technical: wola.url test
```

Sample output:
```
[ROUTE BENCHMARK]
  • Localhost (127.0.0.1:5000)       : 0.42 ms  [OPTIMAL]
  • Wi-Fi LAN (192.168.1.45:5000)    : 2.15 ms  [FAST]
  • Mobile Hotspot (192.168.43.1:5000): 3.80 ms  [RELIABLE]
  • Tailscale Mesh                   : 22.40 ms [DIRECT P2P]
  • Public Tunnel                    : 74.10 ms [REMOTE HTTPS]
```

---

## 📱 Connecting Third-Party Client Apps

Wolarc's API Gateway on port 5000 is 100% OpenAI API compatible.

### Common Settings:
- **API Host / Base URL**: `https://<subdomain>.trycloudflare.com/v1` or `http://192.168.1.X:5000/v1`
- **API Key**: `sk-test` (or any non-empty string)
- **Model**: Auto-detected or enter active model name

### 1. Chatbox (Mobile & Desktop)
1. Open Chatbox -> **Settings** -> **Model Provider**.
2. Select **OpenAI API**.
3. Set **API Host**: paste your tunnel or LAN URL (e.g. `https://xxx.trycloudflare.com`).
4. Set **API Key**: `sk-test`.
5. Tap **Check** to verify connectivity.

### 2. Obsidian (Smart Connections & Copilot Plugins)
1. In Obsidian Settings -> Community Plugins -> **Smart Connections** or **Copilot**.
2. Select **Custom OpenAI Compatible** provider.
3. Set **Chat Endpoint**: `https://<subdomain>.trycloudflare.com/v1/chat/completions`.
4. Set **API Key**: `sk-test`.

### 3. NextChat / OpenWebUI
1. Go to Settings -> Model -> Custom Models.
2. Endpoint: `https://<subdomain>.trycloudflare.com/v1`.
3. Model: Select or enter active model name.

### 4. Python OpenAI SDK
```python
from openai import OpenAI

client = OpenAI(
    base_url="https://<subdomain>.trycloudflare.com/v1",  # or http://127.0.0.1:5000/v1
    api_key="sk-test"
)

response = client.chat.completions.create(
    model="default",
    messages=[{"role": "user", "content": "Explain quantum entanglement in 2 sentences."}],
    stream=True
)

for chunk in response:
    content = chunk.choices[0].delta.content or ""
    print(content, end="", flush=True)
print()
```

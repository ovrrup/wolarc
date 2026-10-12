# Offline Air-Gapped Hotspot & LAN Gateway 🏕️

Wolarc is engineered to operate 100% offline without requiring internet access, cloud servers, or cellular data plans.

---

## 📴 Air-Gapped Operational Modes

### 1. Android Portable Hotspot Mode
Turn your Android device running Wolarc into a private, portable AI router:
1. On your Android phone, enable **Personal Hotspot** (mobile data can remain **OFF**).
2. Connect your laptop, tablet, or another phone to this Wi-Fi network.
3. Android portable hotspots automatically assign the host gateway IP `192.168.43.1`.
4. Run in Termux:
   ```bash
   wola.hotspot      # Technical: wola.local
   ```
5. On connected devices, open:
   ```
   http://192.168.43.1:5000/
   ```

### 2. Local Wi-Fi LAN Mode
When connected to your home or office Wi-Fi network:
- Run `wola.hotspot` or `wola.local`.
- Wolarc scans local network interfaces (`wlan0`, `eth0`) and outputs the exact local subnet URL (e.g. `http://192.168.1.150:5000`).
- Any device on the same local Wi-Fi can immediately access the Web Chat UI or point OpenAI-compatible clients to `http://192.168.1.150:5000/v1`.

### 3. USB Tethering / RNDIS Mode
For zero-wireless, high-speed, interference-free wired communication:
1. Connect your Android phone to your PC via a USB cable.
2. In Android Settings, turn on **USB Tethering**.
3. Android assigns the virtual interface IP `192.168.42.129`.
4. Point your PC client app to `http://192.168.42.129:5000/v1`.

---

## 📷 Instant In-Terminal QR Code Pairing

Running `wola.hotspot` or `wola.local` renders a high-contrast ASCII QR code directly inside your terminal session:

```bash
wola.hotspot
```

- Open the camera on your iPhone, iPad, or secondary Android device.
- Scan the terminal QR code to open the Standalone Web Chat UI instantly in your mobile browser.
- No typing lengthy IP addresses or port numbers.

---

## 🌐 The Standalone Zero-Dependency Web Chat UI

When you navigate to `http://<ip>:5000/`, Wolarc serves an embedded browser client:
- **Zero External Dependencies**: Contains zero CDN links, external Google Fonts, or remote JavaScript bundles. Operates completely in pure offline environments.
- **Rich Markdown & Syntax Highlighting**: Clean rendering of code blocks, tables, and lists.
- **Collapsible Cognitive Reasoning Boxes**: Displays model reasoning traces (`<think>`) in expandable dark containers with real-time toggle controls.
- **Live SSE Token Streaming**: Ultra-smooth real-time generation.

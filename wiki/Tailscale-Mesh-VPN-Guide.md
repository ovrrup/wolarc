# Tailscale WireGuard Mesh VPN Guide 🔒

Tailscale provides an encrypted WireGuard mesh network connecting your devices directly without opening ports on your router, configuring DDNS, or relying on public third-party relay tunnels.

---

## 🌟 Why Tailscale for Wolarc?

1. **Persistent MagicDNS Hostnames**: Your node receives a permanent hostname (e.g. `http://wola-phone.your-tailnet.ts.net:5000`). Unlike temporary SSH or Cloudflare tunnel URLs that change on reconnect, Tailscale URLs **never change**.
2. **Direct Peer-to-Peer WireGuard Encryption**: Traffic flows directly between client and host over UDP hole-punching (STUN/DERP) with end-to-end cryptographic authentication.
3. **Seamless Network Roaming**: Move freely between home Wi-Fi, office networks, and cellular 5G without breaking active sessions.
4. **Zero Battery Overhead**: Runs efficiently in Android's native VPN subsystem.

---

## 📱 Android Setup (Termux)

Because Android isolates VPN network permissions, Tailscale on Android runs through the official Android application:

### Step 1: Install Tailscale App
1. Install **Tailscale** from [Google Play](https://play.google.com/store/apps/details?id=com.tailscale.ipn) or [F-Droid](https://f-droid.org/).
2. Open the app, log in with your identity provider (Google, GitHub, Microsoft, etc.), and toggle the connection **ON**.

### Step 2: Verify in Wolarc
Open Termux and run:
```bash
wola.tailscale status     # Technical: wola.ts status
```

Wolarc detects the active Android VPN interface, extracts your assigned 100.x.y.z IP address, and verifies connectivity:
```
=================================================================
               WOLA TAILSCALE MESH INTEGRATION
=================================================================
  Status       : CONNECTED
  Tailscale IP : 100.84.120.45
  Endpoint     : http://100.84.120.45:5000
  OpenAI Base  : http://100.84.120.45:5000/v1
  MagicDNS     : http://wola-phone.tailnet-xyz.ts.net:5000/v1
=================================================================
```

---

## 🐧 Linux Setup

On Linux systems, Wolarc interfaces directly with the `tailscale` CLI daemon:

### Step 1: Install Tailscale CLI
```bash
curl -fsSL https://tailscale.com/install.sh | sh
```

### Step 2: Authenticate Node
```bash
wola.tailscale up         # Technical: wola.ts up
```
Follow the printed authentication URL in your browser to authorize the machine on your tailnet.

---

## 🔗 Connecting Remote Devices

Once both your Wolarc host and your client device (Mac, Windows PC, iPhone, iPad) are logged into the same Tailscale account:

1. Open your client app (Chatbox, Obsidian, or browser).
2. Set API Host to:
   ```
   http://100.84.120.45:5000/v1
   ```
   Or using MagicDNS:
   ```
   http://wola-phone:5000/v1
   ```
3. Set API Key to `sk-test`.
4. Chat with complete privacy across any network worldwide!

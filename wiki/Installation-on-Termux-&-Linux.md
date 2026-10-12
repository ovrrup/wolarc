# Installation on Termux & Linux 🐧📱

Detailed platform setup instructions, prerequisite configurations, and dependency resolution for Android (Termux) and Linux edge environments.

---

## 📱 Part 1: Android Setup (Termux)

### Step 1: Install Termux & Termux:API
1. Download **Termux** from [F-Droid](https://f-droid.org/en/packages/com.termux/) or the [Termux GitHub Releases](https://github.com/termux/termux-app/releases).
   > [!CAUTION]
   > Do NOT install Termux from Google Play Store. It has been unmaintained since 2020 and cannot update packages.
2. (Optional but recommended) Install **Termux:API** from F-Droid to enable battery telemetry and wake-locks.

### Step 2: Grant Storage Permissions
Open Termux and run:
```bash
termux-setup-storage
```
Tap **Allow** when Android requests storage access. This maps shared internal storage to `~/storage/shared`.

### Step 3: Update Package Repositories
```bash
pkg update -y && pkg upgrade -y
```

### Step 4: Run the Wolarc Installer
```bash
curl -fsSL https://raw.githubusercontent.com/ovrrup/wolarc/main/install.sh | bash
```

### Termux Battery Optimization Whitelist
To keep Wolarc running continuously in the background when the phone screen is turned off:
1. Long-press the **Termux app icon** -> **App Info** -> **Battery**.
2. Select **Unrestricted** (or "Don't optimize").
3. In Termux, swipe down from the notification drawer and tap **Acquire Wakelock**. (Wolarc also acquires and releases this automatically via `termux-wake-lock` when `wola.start` and `wola.stop` are run).

---

## 🐧 Part 2: Linux Setup (Debian / Ubuntu / Arch / Fedora)

### Step 1: Install System Dependencies

#### Debian / Ubuntu / Mint:
```bash
sudo apt update
sudo apt install -y curl git python3 python3-pip python3-venv golang-go libgomp1
```

#### Arch Linux / Manjaro:
```bash
sudo pacman -Syu --noconfirm curl git python python-pip go gcc
```

#### Fedora / RHEL:
```bash
sudo dnf install -y curl git python3 python3-pip golang gcc
```

### Step 2: Run the Installer
```bash
curl -fsSL https://raw.githubusercontent.com/ovrrup/wolarc/main/install.sh | bash
```

The installer detects your Linux distribution and architecture (`x86_64` or `aarch64`), acquires the prebuilt `llama-server` binary, compiles the Go TUI, and configures the environment.

---

## ⚙️ Part 3: Installer Execution Modes

`install.sh` provides 3 distinct setup workflows:

### 1. Clean Slate Mode (`--clean` / `-c`)
```bash
bash install.sh --clean -y
```
- Halts all running Wolarc processes.
- Cleans stale socket locks, PID records, and log files.
- Refreshes all binaries in `/data/data/com.termux/files/usr/bin/` (or `/usr/local/bin/`).
- **Preserves downloaded GGUF models in `~/models`**.

### 2. Adjust & Merge BETA (`--adjust` / `-a`)
```bash
bash install.sh --adjust -y
```
- Reconciles configuration differences between older setups and new versions.
- Automatically updates file paths in `~/.wolarc` to match the centralized `~/.wola/` layout.
- Migrates documents, chats, and vector embeddings without data loss.

### 3. Safe Upgrade Mode (`--upgrade` / `-u`)
```bash
bash install.sh --upgrade -y
```
- Pulls latest updates from the Git repository.
- Recompiles the Go TUI (`wola`) and deploys new CLI scripts to `~/.wola/lib/` and `$PATH`.
- Preserves all user configurations, databases, models, and personal knowledge files.

---

## 🔍 Part 4: Verifying the Installation

Run the diagnostic check:
```bash
wola.check        # Technical: wola.status
```

Expected output:
```
=================================================================
                    WOLA SERVICE STATUS
=================================================================
  Switch State : ON / OFF
  • Chat Server  (8080): Stopped / Running (PID: xxxxx)
  • Embed Server (8081): Stopped / Running (PID: xxxxx)
  • Agent API    (5000): Stopped / Running (PID: xxxxx)
  • Indexer Daemon     : Stopped / Running (PID: xxxxx)
  • Tunnel Daemon      : Stopped / Running (PID: xxxxx)
  • Tunnel URL         : https://...
=================================================================
```

Verify the Go TUI binary:
```bash
wola --help
```

Verify python modules:
```bash
python3 -c "import flask, requests; print('Python environment OK')"
```

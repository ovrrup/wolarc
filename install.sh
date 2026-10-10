#!/data/data/com.termux/files/usr/bin/bash
set -e

# ANSI Colors
C_RESET="\033[0m"
C_BOLD="\033[1m"
C_DIM="\033[2m"
C_RED="\033[1;31m"
C_GREEN="\033[1;32m"
C_YELLOW="\033[1;33m"
C_CYAN="\033[1;36m"
C_MAGENTA="\033[1;35m"

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" 2>/dev/null && pwd)"

# One-command installer bootstrap: If run via curl pipe or outside repo, fetch repository first
REPO_URL="https://github.com/ovrrup/wolarc.git"
INSTALL_DIR="$HOME/wola"

if [ ! -f "$DIR/config/wolarc.example" ] || [ ! -d "$DIR/cmd/wola-tui" ]; then
    echo -e "\033[1;36m[*] One-command installation detected. Bootstrapping Wolarc repository...\033[0m"
    if ! command -v git >/dev/null 2>&1; then
        echo -e "  [*] Installing git..."
        if [ -d "/data/data/com.termux/files/usr/bin" ]; then
            pkg update -y && pkg install -y git
        elif command -v apt-get >/dev/null 2>&1; then
            sudo apt-get update -qq && sudo apt-get install -y git
        elif command -v pacman >/dev/null 2>&1; then
            sudo pacman -Sy --noconfirm git
        elif command -v dnf >/dev/null 2>&1; then
            sudo dnf install -y git
        fi
    fi

    if [ -d "$INSTALL_DIR/.git" ]; then
        echo -e "  [*] Updating existing repository at $INSTALL_DIR..."
        git -C "$INSTALL_DIR" pull --quiet || true
    else
        echo -e "  [*] Cloning repository into $INSTALL_DIR..."
        mkdir -p "$HOME"
        git clone --quiet "$REPO_URL" "$INSTALL_DIR"
    fi

    DIR="$INSTALL_DIR"
fi

# Detect environment and appropriate installation prefix
if [ -d "/data/data/com.termux/files/usr/bin" ]; then
    IS_TERMUX=true
    TARGET_BIN="${PREFIX:-/data/data/com.termux/files/usr}/bin"
    DEFAULT_KNOWLEDGE="$HOME/storage/shared/Documents/agent_knowledge"
    DEFAULT_CHATS="$HOME/storage/shared/Documents/agent_chats"
else
    IS_TERMUX=false
    TARGET_BIN="${HOME}/.local/bin"
    [ -w "/usr/local/bin" ] && TARGET_BIN="/usr/local/bin"
    DEFAULT_KNOWLEDGE="$HOME/Documents/agent_knowledge"
    DEFAULT_CHATS="$HOME/Documents/agent_chats"
fi

MODE="interactive"
INSTALL_LLAMA=""
INSTALL_GO_TUI=""

while [ $# -gt 0 ]; do
    case "$1" in
        "--upgrade"|"-u") MODE="upgrade" ;;
        "--clean") MODE="clean" ;;
        "--adjust") MODE="adjust" ;;
        "-y"|"--yes"|"--unattended") MODE="unattended" ;;
        "--prefix") TARGET_BIN="$2"; shift ;;
        "--tui"|"--with-tui") INSTALL_GO_TUI="true" ;;
        "--no-tui"|"--without-tui") INSTALL_GO_TUI="false" ;;
        "--llama"|"--with-llama") INSTALL_LLAMA="true" ;;
        "--no-llama"|"--without-llama") INSTALL_LLAMA="false" ;;
        "--help"|"-h")
            echo "Usage: bash install.sh [--upgrade | --clean | --adjust | -y | --prefix <path> | --[no-]tui | --[no-]llama]"
            exit 0
            ;;
    esac
    shift
done

print_banner() {
    clear 2>/dev/null || true
    echo -e "${C_CYAN}=================================================================${C_RESET}"
    echo -e " ${C_BOLD}██╗    ██╗ ██████╗ ██╗      █████╗ ██████╗  ██████╗${C_RESET}"
    echo -e " ${C_BOLD}██║    ██║██╔═══██╗██║     ██╔══██╗██╔══██╗██╔════╝${C_RESET}"
    echo -e " ${C_BOLD}██║ █╗ ██║██║   ██║██║     ███████║██████╔╝██║     ${C_RESET}"
    echo -e " ${C_BOLD}██║███╗██║██║   ██║██║     ██╔══██║██╔══██╗██║     ${C_RESET}"
    echo -e " ${C_BOLD}╚███╔███╔╝╚██████╔╝███████╗██║  ██║██║  ██║╚██████╗${C_RESET}"
    echo -e "  ${C_BOLD}╚══╝╚══╝  ╚═════╝ ╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝ ╚═════╝${C_RESET}"
    echo -e "      ${C_YELLOW}Autonomous Local RAG AI Node Suite for Android/Linux${C_RESET}"
    echo -e "${C_CYAN}=================================================================${C_RESET}"
}

print_step() {
    echo -e "\n${C_BOLD}${C_CYAN}[Step $1/6]${C_RESET} ${C_BOLD}$2${C_RESET}"
}

# -------------------------------------------------------------
# STEP 1: Hardware & Environment Diagnostics
# -------------------------------------------------------------
diagnose_hardware() {
    print_step "1" "Hardware & Environment Check"
    
    local tot_ram=0
    local free_ram=0
    if [ -f "/proc/meminfo" ]; then
        tot_ram=$(grep MemTotal /proc/meminfo | awk '{print int($2/1024)}')
        free_ram=$(grep MemAvailable /proc/meminfo | awk '{print int($2/1024)}')
    fi

    local free_disk=$(df -h "$HOME" | tail -1 | awk '{print $4}')

    echo -e "  • Platform       : $( [ "$IS_TERMUX" = true ] && echo "${C_GREEN}Termux (Android)${C_RESET}" || echo "${C_CYAN}Linux${C_RESET}" )"
    echo -e "  • Total RAM      : ${C_YELLOW}${tot_ram} MB${C_RESET} (Available: ${free_ram} MB)"
    echo -e "  • Storage Free   : ${C_GREEN}${free_disk}${C_RESET} in $HOME"
    echo -e "  • Target Bin Dir : ${C_DIM}${TARGET_BIN}${C_RESET}"

    if [ "$tot_ram" -gt 0 ] && [ "$tot_ram" -lt 3000 ]; then
        echo -e "  ${C_YELLOW}ℹ Notice: Compact RAM device detected. Recommended models: Qwen 2.5 1.5B or Llama 3.2 1B.${C_RESET}"
    fi
}

# -------------------------------------------------------------
# STEP 2: System Dependencies & llama-server Check
# -------------------------------------------------------------
check_dependencies() {
    print_step "2" "System Dependencies & Engine Check"

    # Llama-server engine detection
    local llama_path=""
    # Clean up stale/broken symlink if present
    if [ -L "$TARGET_BIN/llama-server" ] && [ ! -e "$TARGET_BIN/llama-server" ]; then
        rm -f "$TARGET_BIN/llama-server"
    fi

    if command -v llama-server >/dev/null 2>&1; then
        llama_path=$(command -v llama-server)
    elif [ -x "$TARGET_BIN/llama-server" ]; then
        llama_path="$TARGET_BIN/llama-server"
    elif [ -x "$HOME/llama.cpp/build/bin/llama-server" ]; then
        llama_path="$HOME/llama.cpp/build/bin/llama-server"
        mkdir -p "$TARGET_BIN"
        ln -sf "$llama_path" "$TARGET_BIN/llama-server" 2>/dev/null || true
    fi

    # 1. Permission for Llama Inference Engine
    if [ -z "$INSTALL_LLAMA" ]; then
        if [ -n "$llama_path" ]; then
            INSTALL_LLAMA="skip"
            echo -e "  ${C_GREEN}✓ Existing llama-server engine verified:${C_RESET} ${llama_path}"
        elif [ "$MODE" = "interactive" ]; then
            echo -e "\n  ${C_BOLD}${C_CYAN}─── [Permission 1/2] Llama Inference Engine ───${C_RESET}"
            echo -e "  Wolarc requires 'llama-server' (llama.cpp) to run GGUF neural models locally."
            echo -e "  Would you like the setup wizard to install llama-server and required llama engine files?"
            echo -e "    ${C_GREEN}[y] Yes${C_RESET} - Install llama-server (${IS_TERMUX:+via pkg install llama-cpp}${IS_TERMUX:-via official prebuilt release})"
            echo -e "    ${C_YELLOW}[n] No${C_RESET}  - Skip llama-server (configure LLAMA_BIN manually in ~/.wolarc later)"
            read -p "  Install llama-server engine now? [Y/n]: " llama_ans
            llama_ans=${llama_ans:-y}
            if [[ "$llama_ans" =~ ^[Yy] ]]; then
                INSTALL_LLAMA="true"
            else
                INSTALL_LLAMA="false"
            fi
        else
            INSTALL_LLAMA="true"
        fi
    fi

    # 2. Permission for Experimental Go Designed TUI
    if [ -z "$INSTALL_GO_TUI" ]; then
        if [ "$MODE" = "interactive" ]; then
            echo -e "\n  ${C_BOLD}${C_CYAN}─── [Permission 2/2] Experimental Go Designed TUI ───${C_RESET}"
            echo -e "  Wolarc includes an experimental modern Go Bubble Tea terminal UI ('wola-tui')."
            echo -e "  ${C_YELLOW}Notice: This is an experimental design that can be removed completely.${C_RESET}"
            echo -e "  You can choose to install the Go TUI, or continue with the classic lightweight CLI suite without it."
            echo -e "    ${C_GREEN}[1] Yes${C_RESET} - Install Go Bubble Tea TUI & Go files (Full modern interface) [Recommended]"
            echo -e "    ${C_YELLOW}[2] No${C_RESET}  - Continue without Go TUI (Use classic lightweight CLI suite only)"
            read -p "  Install Go TUI and required Go files? [1/2] (default: 1): " tui_ans
            tui_ans=${tui_ans:-1}
            if [[ "$tui_ans" == "1" ]] || [[ "$tui_ans" =~ ^[Yy] ]]; then
                INSTALL_GO_TUI="true"
            else
                INSTALL_GO_TUI="false"
            fi
        else
            INSTALL_GO_TUI="true"
        fi
    fi

    local missing=()
    for cmd in python3 git curl; do
        if ! command -v "$cmd" >/dev/null 2>&1; then
            missing+=("$cmd")
        fi
    done
    if ! command -v ssh >/dev/null 2>&1; then
        missing+=("openssh")
    fi

    # Only check/install Go compiler if user wants the Go TUI
    if [ "$INSTALL_GO_TUI" = "true" ] && ! command -v go >/dev/null 2>&1; then
        if [ "$IS_TERMUX" = true ]; then
            missing+=("golang")
        elif command -v apt-get >/dev/null 2>&1; then
            missing+=("golang-go")
        else
            missing+=("golang")
        fi
    fi

    # Auto-include llama-cpp for Termux only if user permitted
    if [ "$IS_TERMUX" = true ] && [ "$INSTALL_LLAMA" = "true" ] && ! command -v llama-server >/dev/null 2>&1; then
        missing+=("llama-cpp")
    fi

    if [ ${#missing[@]} -gt 0 ]; then
        echo -e "\n  ${C_RED}[-] Installing missing system packages: ${missing[*]}${C_RESET}"
        if [ "$IS_TERMUX" = true ]; then
            pkg install -y "${missing[@]}"
        elif command -v apt-get >/dev/null 2>&1; then
            sudo apt-get update -qq && sudo apt-get install -y "${missing[@]}"
        elif command -v pacman >/dev/null 2>&1; then
            sudo pacman -Sy --noconfirm "${missing[@]}"
        elif command -v dnf >/dev/null 2>&1; then
            sudo dnf install -y "${missing[@]}"
        else
            echo -e "  ${C_RED}[!] Please install missing packages using your package manager.${C_RESET}"
            exit 1
        fi
    fi

    echo -e "  ${C_GREEN}✓ Core utilities verified (python3, git, curl, ssh)${C_RESET}"
    if [ "$INSTALL_GO_TUI" = "true" ]; then
        echo -e "  ${C_GREEN}✓ Go compiler verified for TUI build${C_RESET}"
    else
        echo -e "  ${C_DIM}• Go compiler skipped (Classic CLI mode selected)${C_RESET}"
    fi

    # Optional PDF support
    if ! command -v pdftotext >/dev/null 2>&1; then
        echo -e "  ${C_DIM}• 'pdftotext' not found (PDF parsing disabled). Install poppler to enable.${C_RESET}"
        if [ "$IS_TERMUX" = true ] && [ "$MODE" = "interactive" ]; then
            read -p "    Install poppler now for PDF support? (y/N): " pdf_ans
            if [[ "$pdf_ans" =~ ^[Yy] ]]; then
                pkg install -y poppler
            fi
        fi
    else
        echo -e "  ${C_GREEN}✓ PDF document extractor available (pdftotext)${C_RESET}"
    fi

    # Install llama-server prebuilt if user approved and still missing
    if [ "$INSTALL_LLAMA" = "true" ] && [ -z "$llama_path" ]; then
        echo -e "  ${C_YELLOW}[*] Installing prebuilt llama-server engine...${C_RESET}"
        if [ "$IS_TERMUX" = true ]; then
            pkg install -y llama-cpp >/dev/null 2>&1 || true
            if command -v llama-server >/dev/null 2>&1; then
                llama_path=$(command -v llama-server)
            fi
        fi

        if [ -z "$llama_path" ]; then
            local arch
            arch=$(uname -m)
            local asset_pattern=""
            case "$arch" in
                "x86_64"|"amd64") asset_pattern="llama-.*-bin-ubuntu-x64.tar.gz" ;;
                "aarch64"|"arm64")
                    if [ "$IS_TERMUX" = true ]; then
                        asset_pattern="llama-.*-bin-android-arm64.tar.gz"
                    else
                        asset_pattern="llama-.*-bin-ubuntu-arm64.tar.gz"
                    fi
                    ;;
            esac

            if [ -n "$asset_pattern" ]; then
                echo -e "  [*] Querying official prebuilt binary for ${arch}..."
                local dl_url
                dl_url=$(curl -sL "https://api.github.com/repos/ggerganov/llama.cpp/releases?per_page=5" 2>/dev/null | grep "browser_download_url" | grep -E "$asset_pattern" | head -n 1 | cut -d '"' -f 4)
                if [ -n "$dl_url" ]; then
                    echo -e "  [*] Downloading prebuilt engine: ${dl_url}..."
                    local tmp_tar="${TMPDIR:-/tmp}/llama_server_dl.tar.gz"
                    if curl -sL "$dl_url" -o "$tmp_tar"; then
                        mkdir -p "$TARGET_BIN"
                        tar -xzf "$tmp_tar" -C "$TARGET_BIN" --wildcards "*llama-server*" --strip-components=2 2>/dev/null || \
                        tar -xzf "$tmp_tar" -C "$TARGET_BIN" --wildcards "*llama-server*" --strip-components=1 2>/dev/null || \
                        tar -xzf "$tmp_tar" -C "$TARGET_BIN" 2>/dev/null || true
                        rm -f "$tmp_tar"
                        chmod +x "$TARGET_BIN/llama-server" 2>/dev/null || true
                        [ -x "$TARGET_BIN/llama-server" ] && llama_path="$TARGET_BIN/llama-server"
                    fi
                fi
            fi
        fi
    fi

    if [ -n "$llama_path" ]; then
        echo -e "  ${C_GREEN}✓ llama-server verified & ready:${C_RESET} ${llama_path}"
    elif [ "$INSTALL_LLAMA" = "false" ]; then
        echo -e "  ${C_YELLOW}• llama-server skipped by user permission.${C_RESET}"
        echo -e "    ${C_DIM}Configure LLAMA_BIN later in ~/.wolarc or run 'pkg install llama-cpp'.${C_RESET}"
    else
        echo -e "  ${C_YELLOW}⚠ 'llama-server' binary not found.${C_RESET}"
        echo -e "    ${C_DIM}Run 'pkg install llama-cpp' on Termux or install llama-server into PATH.${C_RESET}"
    fi
}

# -------------------------------------------------------------
# STEP 3: Pre-existing Setup & Conflict Inspection
# -------------------------------------------------------------
detect_and_handle_conflicts() {
    print_step "3" "Pre-existing Setup & Conflict Inspection"

    local conflicts_found=false
    local proc_list=()
    local file_list=()
    local bin_list=()

    # 1. Active interfering processes
    if pgrep -f "llama-server" >/dev/null 2>&1; then
        local pids
        pids=$(pgrep -f "llama-server" | tr '\n' ' ')
        proc_list+=("llama-server (PIDs: $pids)")
        conflicts_found=true
    fi
    if pgrep -f "agent.py" >/dev/null 2>&1; then
        local pids
        pids=$(pgrep -f "agent.py" | tr '\n' ' ')
        proc_list+=("agent.py (PIDs: $pids)")
        conflicts_found=true
    fi
    if pgrep -f "ssh.*(localhost.run|serveo.net|pinggy.io)" >/dev/null 2>&1; then
        local pids
        pids=$(pgrep -f "ssh.*(localhost.run|serveo.net|pinggy.io)" | tr '\n' ' ')
        proc_list+=("Wola SSH tunnel daemon (PIDs: $pids)")
        conflicts_found=true
    fi

    # 2. Existing configuration & daemons in $HOME
    [ -f "$HOME/.wolarc" ] && file_list+=("$HOME/.wolarc (configuration)") && conflicts_found=true
    [ -f "$HOME/agent.py" ] && file_list+=("$HOME/agent.py (API server script)") && conflicts_found=true
    [ -f "$HOME/indexer.py" ] && file_list+=("$HOME/indexer.py (RAG indexer script)") && conflicts_found=true
    [ -f "$HOME/.wola.lock" ] && file_list+=("$HOME/.wola.lock (runtime lock)") && conflicts_found=true
    [ -f "$HOME/knowledge_db.json" ] && file_list+=("$HOME/knowledge_db.json (vector database)") && conflicts_found=true

    for logf in "$HOME/agent.log" "$HOME/tunnel.log" "$HOME/llama_chat.log" "$HOME/llama_embed.log" "$HOME/.wola_url" "$HOME/.wola_bench_cache.json" "$HOME/.wola_update_cache.json"; do
        if [ -f "$logf" ]; then
            file_list+=("$logf")
            conflicts_found=true
        fi
    done

    # 3. Installed binaries / symlinks in $TARGET_BIN
    if [ -d "$TARGET_BIN" ]; then
        for b in "$TARGET_BIN"/wola* "$TARGET_BIN"/wolarc; do
            if [ -e "$b" ] || [ -L "$b" ]; then
                bin_list+=("$b")
                conflicts_found=true
            fi
        done
    fi

    if [ "$conflicts_found" = false ]; then
        echo -e "  ${C_GREEN}✓ Clean environment verified.${C_RESET} No conflicting processes or existing files found."
        return 0
    fi

    echo -e "  ${C_YELLOW}⚠ Pre-existing Wola setup or potentially interfering files/processes detected!${C_RESET}"

    if [ ${#proc_list[@]} -gt 0 ]; then
        echo -e "\n  ${C_BOLD}${C_RED}[Active Processes]${C_RESET}"
        for p in "${proc_list[@]}"; do
            echo -e "    • $p"
        done
    fi

    if [ ${#file_list[@]} -gt 0 ]; then
        echo -e "\n  ${C_BOLD}${C_YELLOW}[Existing Config, Daemons & Runtime Files]${C_RESET}"
        for f in "${file_list[@]}"; do
            echo -e "    • $f"
        done
    fi

    if [ ${#bin_list[@]} -gt 0 ]; then
        echo -e "\n  ${C_BOLD}${C_CYAN}[Existing Binaries in $TARGET_BIN (${#bin_list[@]} items)]${C_RESET}"
        local sample_bins=("${bin_list[@]:0:5}")
        for b in "${sample_bins[@]}"; do
            echo -e "    • $(basename "$b")"
        done
        if [ ${#bin_list[@]} -gt 5 ]; then
            echo -e "    • ... and $((${#bin_list[@]} - 5)) more"
        fi
    fi

    echo ""
    echo -e "${C_CYAN}-----------------------------------------------------------------${C_RESET}"
    echo -e "How would you like to handle existing files and directories?"
    echo -e "  ${C_BOLD}[1] Delete / Clean Slate${C_RESET} (Remove interfering files & stop processes)"
    echo -e "      ${C_DIM}Stops running instances, clears stale locks/logs, removes old binaries.${C_RESET}"
    echo -e "      ${C_GREEN}Note: Downloaded models in ~/models are safely preserved.${C_RESET}"
    echo ""
    echo -e "  ${C_BOLD}[2] Adjust & Merge (BETA)${C_RESET} (Reconcile paths, adapt config & databases)"
    echo -e "      ${C_YELLOW}⚠️  BETA FEATURE: May modify or harm existing files & directories!${C_RESET}"
    echo ""
    echo -e "  ${C_BOLD}[3] Safe Upgrade${C_RESET} (Retain existing configuration and proceed)"
    echo -e "${C_CYAN}-----------------------------------------------------------------${C_RESET}"

    local choice=""
    if [ "$MODE" = "clean" ]; then
        choice="1"
    elif [ "$MODE" = "adjust" ]; then
        choice="2"
    elif [ "$MODE" = "upgrade" ] || [ "$MODE" = "unattended" ]; then
        choice="3"
    else
        read -p "Select option [1=Clean / 2=Adjust (BETA) / 3=Safe Upgrade]: " choice
    fi

    case "$choice" in
        1|"clean")
            echo -e "\n[*] Cleaning interfering processes and files..."
            pkill -f "llama-server" 2>/dev/null || true
            pkill -f "agent.py" 2>/dev/null || true
            pkill -f "ssh.*(localhost.run|serveo.net|pinggy.io)" 2>/dev/null || true
            sleep 1

            rm -f "$HOME/.wola.lock" "$HOME/.wola_url" "$HOME/.wola_version" "$HOME/.wola_bench_cache.json" "$HOME/.wola_update_cache.json"
            rm -f "$HOME/agent.log" "$HOME/tunnel.log" "$HOME/llama_chat.log" "$HOME/llama_embed.log"
            rm -f "$HOME/agent.py" "$HOME/indexer.py"

            if [ -d "$TARGET_BIN" ]; then
                for b in "$TARGET_BIN"/wola* "$TARGET_BIN"/wolarc; do
                    rm -f "$b" 2>/dev/null || true
                done
            fi

            if [ -f "$HOME/.wolarc" ]; then
                local bak="$HOME/.wolarc.bak.$(date +%s)"
                cp "$HOME/.wolarc" "$bak"
                echo -e "  ${C_DIM}Backed up previous ~/.wolarc to $bak${C_RESET}"
                rm -f "$HOME/.wolarc"
            fi

            echo -e "  ${C_GREEN}✓ Cleaned interfering files and stopped active processes.${C_RESET}"
            echo -e "  ${C_GREEN}✓ ~/models directory and .gguf files preserved intact.${C_RESET}"
            ;;

        2|"adjust")
            echo -e "\n${C_RED}╔══════════════════════════════════════════════════════════════════╗${C_RESET}"
            echo -e "${C_RED}║                 ⚠️  WARNING: BETA FEATURE                        ║${C_RESET}"
            echo -e "${C_RED}╠══════════════════════════════════════════════════════════════════╣${C_RESET}"
            echo -e "${C_YELLOW}║ You have chosen to let Wolarc ADJUST the existing setup.         ║"
            echo -e "║                                                                  ║"
            echo -e "║ This is an experimental BETA feature. Automatic adjustment will  ║"
            echo -e "║ inspect, adapt, and rewrite existing configuration files,        ║"
            echo -e "║ directories, and runtime environments.                           ║"
            echo -e "║                                                                  ║"
            echo -e "║ ${C_BOLD}${C_RED}WARNING: THIS MAY RESULT IN MODIFYING OR HARMING YOUR FILES      ${C_RESET}${C_YELLOW}║"
            echo -e "║ ${C_BOLD}${C_RED}AND DIRECTORIES, INCLUDING CUSTOM PATHS AND DATABASES!           ${C_RESET}${C_YELLOW}║"
            echo -e "╚══════════════════════════════════════════════════════════════════╝${C_RESET}"

            local confirm_adjust="y"
            if [ "$MODE" != "adjust" ] && [ "$MODE" != "unattended" ]; then
                read -p "Are you sure you want Wolarc to adjust existing setup? (y/N): " confirm_adjust
            fi

            if [[ ! "$confirm_adjust" =~ ^[Yy] ]]; then
                echo -e "Adjustment cancelled. Proceeding with Safe Upgrade."
                return 0
            fi

            echo -e "\n[*] Running Wolarc Beta Adjustment & Reconciliation engine..."
            pkill -f "llama-server" 2>/dev/null || true
            pkill -f "agent.py" 2>/dev/null || true
            pkill -f "ssh.*(localhost.run|serveo.net|pinggy.io)" 2>/dev/null || true
            rm -f "$HOME/.wola.lock"

            python3 -c "
import os, sys, re, json, shutil, time

home = os.path.expanduser('~')
cfg_path = os.path.join(home, '.wolarc')
repo_dir = '$DIR'
default_knowledge = '$DEFAULT_KNOWLEDGE'
default_chats = '$DEFAULT_CHATS'

print('  • Reconciling directories and paths...')
models_dir = os.path.join(home, 'models')
knowledge_dir = default_knowledge
chats_dir = default_chats

if os.path.exists(cfg_path):
    bak = f'{cfg_path}.bak.{int(time.time())}'
    shutil.copy2(cfg_path, bak)
    print(f'    - Backed up ~/.wolarc to {bak}')

    existing_vars = {}
    with open(cfg_path, 'r') as f:
        for line in f:
            line = line.strip()
            if line and not line.startswith('#') and '=' in line:
                k, v = line.split('=', 1)
                existing_vars[k.strip()] = v.strip('\"\' ')

    if 'MODELS_DIR' in existing_vars:
        models_dir = os.path.expanduser(existing_vars['MODELS_DIR'])
    if 'KNOWLEDGE_DIR' in existing_vars:
        knowledge_dir = os.path.expanduser(existing_vars['KNOWLEDGE_DIR'])
    if 'CHATS_DIR' in existing_vars:
        chats_dir = os.path.expanduser(existing_vars['CHATS_DIR'])

os.makedirs(models_dir, exist_ok=True)
os.makedirs(knowledge_dir, exist_ok=True)
os.makedirs(chats_dir, exist_ok=True)
print(f'    - Verified Models Dir: {models_dir}')
print(f'    - Verified Knowledge Dir: {knowledge_dir}')
print(f'    - Verified Chats Dir: {chats_dir}')

existing_ggufs = [f for f in os.listdir(models_dir) if f.lower().endswith('.gguf')] if os.path.exists(models_dir) else []
chat_candidate = None
embed_candidate = None

for g in existing_ggufs:
    g_low = g.lower()
    if ('embed' in g_low or 'bge' in g_low) and not embed_candidate:
        embed_candidate = g
    elif ('embed' not in g_low and 'bge' not in g_low) and not chat_candidate:
        chat_candidate = g

example_cfg = os.path.join(repo_dir, 'config', 'wolarc.example')
template_lines = []
if os.path.exists(example_cfg):
    with open(example_cfg, 'r') as f:
        template_lines = f.readlines()

new_config = []
if not os.path.exists(cfg_path):
    new_config = template_lines
else:
    with open(cfg_path, 'r') as f:
        current_lines = f.readlines()

    keys_present = set()
    for l in current_lines:
        m = re.match(r'^\s*([A-Za-z0-9_]+)=', l)
        if m:
            k = m.group(1)
            keys_present.add(k)
            if k == 'CHAT_MODEL' and chat_candidate:
                val = l.split('=', 1)[1].strip('\"\' \n')
                if not val or not os.path.exists(os.path.join(models_dir, val)):
                    l = f'CHAT_MODEL=\"{chat_candidate}\"\n'
                    print(f'    - Adjusted CHAT_MODEL -> {chat_candidate}')
            elif k == 'EMBED_MODEL' and embed_candidate:
                val = l.split('=', 1)[1].strip('\"\' \n')
                if not val or not os.path.exists(os.path.join(models_dir, val)):
                    l = f'EMBED_MODEL=\"{embed_candidate}\"\n'
                    print(f'    - Adjusted EMBED_MODEL -> {embed_candidate}')
            elif k == 'SIMILARITY_THRESHOLD':
                try:
                    s_val = float(l.split('=', 1)[1].strip('\"\' \n'))
                    if s_val < 0.45 or s_val > 0.70:
                        l = 'SIMILARITY_THRESHOLD=\"0.50\"\n'
                        print('    - Adjusted SIMILARITY_THRESHOLD -> 0.50 (optimal high-confidence RAG)')
                except Exception:
                    pass
        new_config.append(l)

    for tl in template_lines:
        tm = re.match(r'^\s*([A-Za-z0-9_]+)=', tl)
        if tm and tm.group(1) not in keys_present:
            new_config.append(tl)
            print(f'    - Injected missing parameter: {tm.group(1)}')

with open(cfg_path, 'w') as f:
    f.writelines(new_config)
print('    - ~/.wolarc successfully reconciled and updated.')

db_path = os.path.join(home, 'knowledge_db.json')
if os.path.exists(db_path):
    try:
        with open(db_path, 'r') as f:
            data = json.load(f)
        dirty = False
        for k, doc in data.items():
            for ch in doc.get('chunks', []):
                v = ch.get('vector', [])
                while isinstance(v, list) and len(v) > 0 and isinstance(v[0], list):
                    v = v[0]
                    ch['vector'] = v
                    dirty = True
        if dirty:
            with open(db_path, 'w') as f:
                json.dump(data, f, indent=2)
            print('    - Flattened 2D vectors in knowledge_db.json to clean 1D vectors.')
        print('    - knowledge_db.json verified valid JSON.')
    except Exception:
        corrupt_bak = f'{db_path}.corrupted.{int(time.time())}'
        shutil.move(db_path, corrupt_bak)
        with open(db_path, 'w') as f:
            json.dump({}, f)
        print(f'    - [!] Repaired corrupted knowledge_db.json (backed up to {corrupt_bak})')

for lf in ['agent.log', 'tunnel.log', 'llama_chat.log', 'llama_embed.log']:
    lp = os.path.join(home, lf)
    if os.path.exists(lp) and os.path.getsize(lp) > 10 * 1024 * 1024:
        with open(lp, 'w') as f:
            pass
        print(f'    - Truncated bloated log: {lf}')
"
            echo -e "  ${C_GREEN}✓ Environment and configurations adjusted.${C_RESET}"
            ;;

        *)
            echo -e "\n[*] Proceeding with Safe Upgrade (keeping existing files and configs)..."
            ;;
    esac
}

# -------------------------------------------------------------
# STEP 4: Python Environment & Dependencies
# -------------------------------------------------------------
setup_python() {
    print_step "4" "Python Dependencies Setup"

    if [ -f "$DIR/requirements.txt" ]; then
        echo -e "  [*] Installing requirements (flask, flask-cors, requests, numpy, watchdog, gunicorn, qrcode)..."
        pip install -r "$DIR/requirements.txt" --quiet
        echo -e "  ${C_GREEN}✓ Python packages installed successfully.${C_RESET}"
    fi
}

# -------------------------------------------------------------
# STEP 5: Configuration & Knowledge Directories
# -------------------------------------------------------------
setup_config() {
    print_step "5" "Configuration Setup"

    local models_dir="$HOME/models"
    local knowledge_dir="$DEFAULT_KNOWLEDGE"
    local chats_dir="$DEFAULT_CHATS"

    if [ "$MODE" = "interactive" ] && [ ! -f "$HOME/.wolarc" ]; then
        echo -e "  ${C_BOLD}Configure default directories:${C_RESET}"
        read -p "  Models directory [default: $models_dir]: " input_models
        [ -n "$input_models" ] && models_dir=$(eval echo "$input_models")

        read -p "  Knowledge base directory [default: $knowledge_dir]: " input_knowledge
        [ -n "$input_knowledge" ] && knowledge_dir=$(eval echo "$input_knowledge")

        read -p "  Chat history directory [default: $chats_dir]: " input_chats
        [ -n "$input_chats" ] && chats_dir=$(eval echo "$input_chats")
    fi

    mkdir -p "$models_dir"
    mkdir -p "$knowledge_dir"
    mkdir -p "$chats_dir"

    # Create config file if not existing
    if [ ! -f "$HOME/.wolarc" ]; then
        cp "$DIR/config/wolarc.example" "$HOME/.wolarc"
        # Update configured directories
        python3 -c "
import sys, re
def set_val(k, v, path):
    with open(path, 'r') as f: lines = f.readlines()
    pat = re.compile(rf'^\s*{k}=.*$')
    lines = [f'{k}=\"{v}\"\n' if pat.match(l) else l for l in lines]
    with open(path, 'w') as f: f.writelines(lines)

set_val('MODELS_DIR', '$models_dir', '$HOME/.wolarc')
set_val('KNOWLEDGE_DIR', '$knowledge_dir', '$HOME/.wolarc')
set_val('CHATS_DIR', '$chats_dir', '$HOME/.wolarc')
" 2>/dev/null
        echo -e "  ${C_GREEN}✓ Initialized ~/.wolarc template with custom paths.${C_RESET}"
    else
        echo -e "  ${C_GREEN}✓ Preserving existing ~/.wolarc.${C_RESET}"
    fi

    # Save repository source path for auto-updates & reconcile LLAMA_BIN
    python3 -c "
import sys, os, re, shutil
src_dir, target_bin, user_tui = sys.argv[1], sys.argv[2], sys.argv[3]
path = os.path.expanduser('~/.wolarc')
if os.path.exists(path):
    with open(path, 'r') as f: lines = f.readlines()
    pat = re.compile(r'^\s*WOLA_SOURCE_DIR=.*$')
    has_it = any(pat.match(l) for l in lines)
    if has_it:
        lines = [f'WOLA_SOURCE_DIR=\"{src_dir}\"\n' if pat.match(l) else l for l in lines]
    else:
        lines.append(f'WOLA_SOURCE_DIR=\"{src_dir}\"\n')

    # Reconcile LLAMA_BIN to verified binary if missing or invalid
    llama_pat = re.compile(r'^\s*LLAMA_BIN\s*=\s*[\"\'\s]*(.*?)[\"\'\s]*$')
    llama_val = None
    for l in lines:
        m = llama_pat.match(l)
        if m:
            llama_val = os.path.expanduser(m.group(1).strip())
            break
    if not llama_val or (llama_val != 'llama-server' and not os.path.exists(llama_val)):
        resolved_bin = 'llama-server' if shutil.which('llama-server') else os.path.join(target_bin, 'llama-server')
        lines = [f'LLAMA_BIN=\"{resolved_bin}\"\n' if llama_pat.match(l) else l for l in lines]

    # Sync ENABLE_TUI according to installer choice
    tui_pat = re.compile(r'^\s*ENABLE_TUI=.*$')
    if user_tui in ('true', 'false'):
        if any(tui_pat.match(l) for l in lines):
            lines = [f'ENABLE_TUI=\"{user_tui}\"\n' if tui_pat.match(l) else l for l in lines]
        else:
            lines.append(f'ENABLE_TUI=\"{user_tui}\"\n')
    else:
        if not any(tui_pat.match(l) for l in lines):
            lines.append('ENABLE_TUI=\"true\"\n')

    with open(path, 'w') as f: f.writelines(lines)
" "$DIR" "$TARGET_BIN" "$INSTALL_GO_TUI" 2>/dev/null
}

# -------------------------------------------------------------
# STEP 6: Deploy Binaries & Create Symlinks
# -------------------------------------------------------------
deploy_binaries() {
    print_step "6" "Installing Executables into ${TARGET_BIN}"

    mkdir -p "$TARGET_BIN"

    # Deploy python background scripts
    cp "$DIR/agent.py" "$HOME/agent.py"
    [ -f "$DIR/indexer.py" ] && cp "$DIR/indexer.py" "$HOME/indexer.py"
    [ -f "$DIR/web_ui.py" ] && cp "$DIR/web_ui.py" "$HOME/web_ui.py"

    # Deploy all commands in bin/
    for f in "$DIR/bin"/*; do
        [ -d "$f" ] && continue
        [ -e "$f" ] || continue
        fname=$(basename "$f")
        [ "$fname" = "wola.gh" ] && continue
        cp "$f" "$TARGET_BIN/$fname"
        chmod +x "$TARGET_BIN/$fname"
    done

    # Remove any stale legacy aliases
    rm -f "$TARGET_BIN/wola.destroy" "$TARGET_BIN/wola.selfdestruct" "$TARGET_BIN/wola.nuke" \
          "$TARGET_BIN/wola.gh" "$TARGET_BIN/wola.upgrade" "$TARGET_BIN/wola.logs" 2>/dev/null || true

    # Create unified Technical <-> Casual command pairs (exactly one alias per command)
    # Start: wola.r (technical) <-> wola.start (casual)
    ln -sf "$TARGET_BIN/wolarc" "$TARGET_BIN/wola.r"
    ln -sf "$TARGET_BIN/wolarc" "$TARGET_BIN/wola.start"

    # Stop: wola.k (technical) <-> wola.stop (casual)
    ln -sf "$TARGET_BIN/wolarc" "$TARGET_BIN/wola.k"
    ln -sf "$TARGET_BIN/wolarc" "$TARGET_BIN/wola.stop"

    # Restart: wola.rs (technical) <-> wola.restart (casual)
    ln -sf "$TARGET_BIN/wolarc" "$TARGET_BIN/wola.rs"
    ln -sf "$TARGET_BIN/wolarc" "$TARGET_BIN/wola.restart"

    # Status: wola.status (technical) <-> wola.check (casual)
    ln -sf "$TARGET_BIN/wolarc" "$TARGET_BIN/wola.status"
    ln -sf "$TARGET_BIN/wolarc" "$TARGET_BIN/wola.check"

    # Dashboard: wola.d (technical) <-> wola.dashboard (casual)
    ln -sf "$TARGET_BIN/wola.d" "$TARGET_BIN/wola.dashboard"

    # Chat: wola.c (technical) <-> wola.chat (casual)
    ln -sf "$TARGET_BIN/wola.chat" "$TARGET_BIN/wola.c"

    # Models: wola.m (technical) <-> wola.models (casual)
    ln -sf "$TARGET_BIN/wola.m" "$TARGET_BIN/wola.models"

    # Store: wola.gt (technical) <-> wola.store (casual)
    ln -sf "$TARGET_BIN/wola.gt" "$TARGET_BIN/wola.store"

    # Knowledge: wola.i (technical) <-> wola.knowledge (casual)
    ln -sf "$TARGET_BIN/wola.i" "$TARGET_BIN/wola.knowledge"

    # Benchmark: wola.b (technical) <-> wola.bench (casual)
    ln -sf "$TARGET_BIN/wola.bench" "$TARGET_BIN/wola.b"

    # Logs: wola.l (technical) <-> wola.log (casual)
    ln -sf "$TARGET_BIN/wola.log" "$TARGET_BIN/wola.l"

    # Thinking: wola.th (technical) <-> wola.think (casual)
    ln -sf "$TARGET_BIN/wola.th" "$TARGET_BIN/wola.think"

    # Remote Link: wola.url (technical) <-> wola.link (casual)
    ln -sf "$TARGET_BIN/wola.url" "$TARGET_BIN/wola.link"

    # Offline Hotspot: wola.local (technical) <-> wola.hotspot (casual)
    ln -sf "$TARGET_BIN/wolarc" "$TARGET_BIN/wola.local"
    ln -sf "$TARGET_BIN/wolarc" "$TARGET_BIN/wola.hotspot"

    # Tailscale Mesh: wola.ts (technical) <-> wola.tailscale (casual)
    ln -sf "$TARGET_BIN/wolarc" "$TARGET_BIN/wola.ts"
    ln -sf "$TARGET_BIN/wolarc" "$TARGET_BIN/wola.tailscale"

    # Settings / Config: wola.t (technical) <-> wola.config (casual)
    ln -sf "$TARGET_BIN/wola.t" "$TARGET_BIN/wola.config"

    # Updates: wola.u (technical) <-> wola.update (casual)
    ln -sf "$TARGET_BIN/wola.u" "$TARGET_BIN/wola.update"

    # Uninstall: wola.purge (technical) <-> wola.uninstall (casual)
    ln -sf "$TARGET_BIN/wola.purge" "$TARGET_BIN/wola.uninstall"

    # Help: wola.h (technical) <-> wola.help (casual)
    ln -sf "$TARGET_BIN/wola-h" "$TARGET_BIN/wola.h"
    ln -sf "$TARGET_BIN/wola-h" "$TARGET_BIN/wola.help"

    # Build and deploy Go Bubble Tea TUI if permitted
    if [ "$INSTALL_GO_TUI" = "true" ]; then
        echo -e "  [*] Building Go Bubble Tea TUI (wola-tui)..."
        local tui_installed=false

        if command -v go >/dev/null 2>&1 && [ -d "$DIR/cmd/wola-tui" ]; then
            export TMPDIR="${TMPDIR:-/data/data/com.termux/files/usr/tmp}"
            [ ! -d "$TMPDIR" ] && export TMPDIR="/tmp"
            mkdir -p "$DIR/bin"
            if (cd "$DIR" && go build -ldflags="-s -w" -o "$DIR/bin/wola-tui" ./cmd/wola-tui); then
                cp "$DIR/bin/wola-tui" "$TARGET_BIN/wola-tui"
                chmod +x "$TARGET_BIN/wola-tui"
                tui_installed=true
                echo -e "  ${C_GREEN}✓ Successfully built and installed Go TUI (${TARGET_BIN}/wola-tui)${C_RESET}"
            fi
        fi

        if [ "$tui_installed" = false ]; then
            if [ -f "$DIR/dist/wola-tui" ]; then
                cp "$DIR/dist/wola-tui" "$TARGET_BIN/wola-tui"
                chmod +x "$TARGET_BIN/wola-tui"
                tui_installed=true
            elif [ -f "$DIR/dist/wola-tui-linux-arm64" ] && [ "$IS_TERMUX" = "true" ]; then
                cp "$DIR/dist/wola-tui-linux-arm64" "$TARGET_BIN/wola-tui"
                chmod +x "$TARGET_BIN/wola-tui"
                tui_installed=true
            elif [ -f "$DIR/bin/wola-tui" ]; then
                cp "$DIR/bin/wola-tui" "$TARGET_BIN/wola-tui"
                chmod +x "$TARGET_BIN/wola-tui"
                tui_installed=true
            fi
        fi

        if [ -f "$TARGET_BIN/wola-tui" ]; then
            ln -sf "$TARGET_BIN/wola-tui" "$TARGET_BIN/wola.tui"
            ln -sf "$TARGET_BIN/wola-tui" "$TARGET_BIN/wola.ui"
            ln -sf "$TARGET_BIN/wola-tui" "$TARGET_BIN/wola"
            echo -e "  ${C_GREEN}✓ Integrated Go TUI: 'wola', 'wola.tui' & 'wola.ui' -> ${TARGET_BIN}/wola-tui${C_RESET}"
        else
            ln -sf "$TARGET_BIN/wolarc" "$TARGET_BIN/wola.tui"
            ln -sf "$TARGET_BIN/wolarc" "$TARGET_BIN/wola.ui"
            ln -sf "$TARGET_BIN/wolarc" "$TARGET_BIN/wola"
        fi
    else
        echo -e "  ${C_DIM}• Skipping Go TUI compilation & deployment (Classic CLI mode active).${C_RESET}"
        rm -f "$TARGET_BIN/wola.tui" "$TARGET_BIN/wola.ui" "$TARGET_BIN/wola-tui" 2>/dev/null || true
        ln -sf "$TARGET_BIN/wolarc" "$TARGET_BIN/wola.tui"
        ln -sf "$TARGET_BIN/wolarc" "$TARGET_BIN/wola.ui"
        ln -sf "$TARGET_BIN/wolarc" "$TARGET_BIN/wola"
        echo -e "  ${C_GREEN}✓ Linked classic CLI suite: 'wola' -> ${TARGET_BIN}/wolarc${C_RESET}"
    fi

    # Save commit version if in git repo
    if [ -d "$DIR/.git" ]; then
        git -C "$DIR" rev-parse --short HEAD > "$HOME/.wola_version" 2>/dev/null || true
    fi

    echo -e "  ${C_GREEN}✓ Installed command pairs into ${TARGET_BIN}:${C_RESET}"
    echo -e "    Start      : wola.start     / wola.r"
    echo -e "    Stop       : wola.stop      / wola.k"
    echo -e "    Restart    : wola.restart   / wola.rs"
    echo -e "    Status     : wola.check     / wola.status"
    echo -e "    Dashboard  : wola.dashboard / wola.d"
    echo -e "    Chat       : wola.chat      / wola.c"
    echo -e "    Models     : wola.models    / wola.m"
    echo -e "    Store      : wola.store     / wola.gt"
    echo -e "    Knowledge  : wola.knowledge / wola.i"
    echo -e "    Benchmark  : wola.bench     / wola.b"
    echo -e "    Logs       : wola.log       / wola.l"
    echo -e "    Thinking   : wola.think     / wola.th"
    echo -e "    Gateway    : wola.link      / wola.url"
    echo -e "    Hotspot    : wola.hotspot   / wola.local"
    echo -e "    Tailscale  : wola.tailscale / wola.ts"
    echo -e "    Config     : wola.config    / wola.t"
    echo -e "    Update     : wola.update    / wola.u"
    echo -e "    Uninstall  : wola.uninstall / wola.purge"
    echo -e "    Help       : wola.help      / wola.h"
}

# -------------------------------------------------------------
# Summary & Next Action Wizard
# -------------------------------------------------------------
post_install() {
    echo -e "\n${C_CYAN}=================================================================${C_RESET}"
    echo -e "  ${C_BOLD}${C_GREEN}✓ WOLARC INSTALLATION COMPLETE!${C_RESET}"
    echo -e "${C_CYAN}=================================================================${C_RESET}"
    echo -e "  ${C_BOLD}Command Cheatsheet (Casual / Technical):${C_RESET}"
    echo -e "    ${C_GREEN}wola.start${C_RESET}     / ${C_DIM}wola.r${C_RESET}        • Start all services (chat, embed, agent API, tunnel)"
    echo -e "    ${C_RED}wola.stop${C_RESET}      / ${C_DIM}wola.k${C_RESET}        • Stop all services cleanly"
    echo -e "    ${C_YELLOW}wola.restart${C_RESET}   / ${C_DIM}wola.rs${C_RESET}       • Relaunch all background daemons"
    echo -e "    ${C_CYAN}wola.check${C_RESET}     / ${C_DIM}wola.status${C_RESET}   • Service process PIDs & port status"
    echo -e "    ${C_CYAN}wola.dashboard${C_RESET} / ${C_DIM}wola.d${C_RESET}        • Open real-time Telemetry Dashboard"
    echo -e "    ${C_GREEN}wola.chat${C_RESET}      / ${C_DIM}wola.c${C_RESET}        • Interactive CLI chat client & one-shot query"
    echo -e "    ${C_YELLOW}wola.models${C_RESET}    / ${C_DIM}wola.m${C_RESET}        • Model manager & GGUF switcher"
    echo -e "    ${C_GREEN}wola.store${C_RESET}     / ${C_DIM}wola.gt${C_RESET}       • Model Store & Local GGUF Importer"
    echo -e "    ${C_YELLOW}wola.knowledge${C_RESET} / ${C_DIM}wola.i${C_RESET}        • Vector knowledge base & note indexer"
    echo -e "    ${C_YELLOW}wola.bench${C_RESET}     / ${C_DIM}wola.b${C_RESET}        • Measure token speed & auto-tune threads"
    echo -e "    ${C_DIM}wola.log${C_RESET}       / ${C_DIM}wola.l${C_RESET}        • Unified live multi-stream log monitor"
    echo -e "    ${C_MAGENTA}wola.think${C_RESET}     / ${C_DIM}wola.th${C_RESET}       • Real-time Reasoning Stream Viewer"
    echo -e "    ${C_CYAN}wola.link${C_RESET}      / ${C_DIM}wola.url${C_RESET}      • Remote Gateway URL, latency & QR code"
    echo -e "    ${C_CYAN}wola.hotspot${C_RESET}   / ${C_DIM}wola.local${C_RESET}    • Air-Gapped Hotspot & LAN Web UI (Offline)"
    echo -e "    ${C_YELLOW}wola.config${C_RESET}    / ${C_DIM}wola.t${C_RESET}        • Configuration parameter tuner"
    echo -e "    ${C_GREEN}wola.update${C_RESET}    / ${C_DIM}wola.u${C_RESET}        • Built-in Update Checker and Installer"
    echo -e "    ${C_RED}wola.uninstall${C_RESET} / ${C_DIM}wola.purge${C_RESET}    • Discrete uninstaller & system reset"
    echo -e "    ${C_CYAN}wola.help${C_RESET}      / ${C_DIM}wola.h${C_RESET}        • Interactive Command Center"
    echo -e "${C_CYAN}=================================================================${C_RESET}"

    # Check if models exist in ~/models
    local models_count=$(find "$HOME/models" -maxdepth 1 -name "*.gguf" 2>/dev/null | wc -l)
    if [ "$models_count" -eq 0 ]; then
        echo -e "\n${C_YELLOW}[!] No GGUF models detected in ~/models.${C_RESET}"
        local store_ans="y"
        if [ "$MODE" = "interactive" ]; then
            read -p "Would you like to auto-download verified lightweight starter models now? (Y/n): " store_ans
            store_ans=${store_ans:-y}
        fi
        if [[ "$store_ans" =~ ^[Yy] ]]; then
            echo -e "  [*] Running automated starter model downloader..."
            "$TARGET_BIN/wola.gt" --quick-start || python3 "$DIR/bin/wola.gt" --quick-start
        fi
    fi

    if [ "$MODE" = "interactive" ]; then
        echo ""
        if [ "$INSTALL_GO_TUI" = "true" ] && [ -x "$TARGET_BIN/wola-tui" ]; then
            read -p "Would you like to launch Wolarc TUI (wola) now? (Y/n): " launch_ans
            launch_ans=${launch_ans:-y}
            if [[ "$launch_ans" =~ ^[Yy] ]]; then
                exec "$TARGET_BIN/wola-tui"
            fi
        else
            read -p "Would you like to launch Wolarc Command Center (wola-h) now? (Y/n): " launch_ans
            launch_ans=${launch_ans:-y}
            if [[ "$launch_ans" =~ ^[Yy] ]]; then
                exec "$TARGET_BIN/wola-h"
            fi
        fi
    fi
}

# Execute sequence
print_banner
diagnose_hardware
check_dependencies
detect_and_handle_conflicts
setup_python
setup_config
deploy_binaries
post_install

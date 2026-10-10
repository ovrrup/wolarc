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

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Detect environment and appropriate installation prefix
if [ -d "/data/data/com.termux/files/usr/bin" ]; then
    IS_TERMUX=true
    TARGET_BIN="${PREFIX:-/data/data/com.termux/files/usr}/bin"
    DEFAULT_KNOWLEDGE="$HOME/storage/shared/Documents/agent_knowledge"
else
    IS_TERMUX=false
    TARGET_BIN="${HOME}/.local/bin"
    [ -w "/usr/local/bin" ] && TARGET_BIN="/usr/local/bin"
    DEFAULT_KNOWLEDGE="$HOME/Documents/agent_knowledge"
fi

MODE="interactive"
while [ $# -gt 0 ]; do
    case "$1" in
        "--upgrade"|"-u") MODE="upgrade" ;;
        "--clean") MODE="clean" ;;
        "--adjust") MODE="adjust" ;;
        "-y"|"--yes"|"--unattended") MODE="unattended" ;;
        "--prefix") TARGET_BIN="$2"; shift ;;
        "--help"|"-h")
            echo "Usage: bash install.sh [--upgrade | --clean | --adjust | -y | --prefix <path>]"
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

    local missing=()
    for cmd in python3 git curl; do
        if ! command -v "$cmd" >/dev/null 2>&1; then
            missing+=("$cmd")
        fi
    done
    if ! command -v ssh >/dev/null 2>&1; then
        missing+=("openssh")
    fi

    if [ ${#missing[@]} -gt 0 ]; then
        echo -e "  ${C_RED}[-] Missing packages: ${missing[*]}${C_RESET}"
        if [ "$IS_TERMUX" = true ]; then
            echo -e "  [*] Installing missing packages via pkg..."
            pkg install -y "${missing[@]}"
        else
            echo -e "  ${C_RED}[!] Please install missing packages using your package manager.${C_RESET}"
            exit 1
        fi
    else
        echo -e "  ${C_GREEN}✓ Core utilities verified (python3, git, curl, ssh)${C_RESET}"
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

    # Llama-server detection
    local llama_path=""
    if command -v llama-server >/dev/null 2>&1; then
        llama_path=$(command -v llama-server)
    elif [ -f "$HOME/llama.cpp/build/bin/llama-server" ]; then
        llama_path="$HOME/llama.cpp/build/bin/llama-server"
    elif [ -f "$TARGET_BIN/llama-server" ]; then
        llama_path="$TARGET_BIN/llama-server"
    fi

    if [ -n "$llama_path" ]; then
        echo -e "  ${C_GREEN}✓ llama-server found:${C_RESET} ${llama_path}"
    else
        echo -e "  ${C_YELLOW}⚠ 'llama-server' binary not found.${C_RESET}"
        echo -e "    ${C_DIM}You can configure LLAMA_BIN later in ~/.wolarc or via 'wola.t LLAMA_BIN <path>'.${C_RESET}"
        echo -e "    ${C_DIM}Build guide: git clone https://github.com/ggerganov/llama.cpp && cmake -B build && cmake --build build -j4${C_RESET}"
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

print('  • Reconciling directories and paths...')
models_dir = os.path.join(home, 'models')
knowledge_dir = default_knowledge

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

os.makedirs(models_dir, exist_ok=True)
os.makedirs(knowledge_dir, exist_ok=True)
print(f'    - Verified Models Dir: {models_dir}')
print(f'    - Verified Knowledge Dir: {knowledge_dir}')

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
            json.load(f)
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
        echo -e "  [*] Installing requirements (flask, flask-cors, requests, numpy, watchdog)..."
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

    if [ "$MODE" = "interactive" ] && [ ! -f "$HOME/.wolarc" ]; then
        echo -e "  ${C_BOLD}Configure default directories:${C_RESET}"
        read -p "  Models directory [default: $models_dir]: " input_models
        [ -n "$input_models" ] && models_dir=$(eval echo "$input_models")

        read -p "  Knowledge base directory [default: $knowledge_dir]: " input_knowledge
        [ -n "$input_knowledge" ] && knowledge_dir=$(eval echo "$input_knowledge")
    fi

    mkdir -p "$models_dir"
    mkdir -p "$knowledge_dir"

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
" 2>/dev/null
        echo -e "  ${C_GREEN}✓ Initialized ~/.wolarc template with custom paths.${C_RESET}"
    else
        echo -e "  ${C_GREEN}✓ Preserving existing ~/.wolarc.${C_RESET}"
    fi

    # Save repository source path for auto-updates
    python3 -c "
import sys, re
path = '$HOME/.wolarc'
with open(path, 'r') as f: lines = f.readlines()
pat = re.compile(r'^\s*WOLA_SOURCE_DIR=.*$')
has_it = any(pat.match(l) for l in lines)
if has_it:
    lines = [f'WOLA_SOURCE_DIR=\"$DIR\"\n' if pat.match(l) else l for l in lines]
else:
    lines.append(f'WOLA_SOURCE_DIR=\"$DIR\"\n')
with open(path, 'w') as f: f.writelines(lines)
" 2>/dev/null
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

    # Deploy all commands in bin/
    for f in "$DIR/bin"/*; do
        [ -d "$f" ] && continue
        [ -e "$f" ] || continue
        fname=$(basename "$f")
        [ "$fname" = "wola.gh" ] && continue
        cp "$f" "$TARGET_BIN/$fname"
        chmod +x "$TARGET_BIN/$fname"
    done

    # Create convenient command symlinks
    ln -sf "$TARGET_BIN/wolarc" "$TARGET_BIN/wola.r"
    ln -sf "$TARGET_BIN/wolarc" "$TARGET_BIN/wola.k"
    ln -sf "$TARGET_BIN/wolarc" "$TARGET_BIN/wola.rs"
    ln -sf "$TARGET_BIN/wola-h" "$TARGET_BIN/wola.h"
    ln -sf "$TARGET_BIN/wola.u" "$TARGET_BIN/wola.upgrade"
    ln -sf "$TARGET_BIN/wola.url" "$TARGET_BIN/wola.link"
    ln -sf "$TARGET_BIN/wola.chat" "$TARGET_BIN/wola.c"
    ln -sf "$TARGET_BIN/wola.bench" "$TARGET_BIN/wola.b"
    ln -sf "$TARGET_BIN/wola.log" "$TARGET_BIN/wola.logs"
    ln -sf "$TARGET_BIN/wola.gt" "$TARGET_BIN/wola.gh"

    # Save commit version if in git repo
    if [ -d "$DIR/.git" ]; then
        git -C "$DIR" rev-parse --short HEAD > "$HOME/.wola_version" 2>/dev/null || true
    fi

    echo -e "  ${C_GREEN}✓ Installed commands into ${TARGET_BIN}:${C_RESET}"
    echo -e "    wola.r  wola.k  wola.rs  wola.d  wola.url  wola.chat  wola.bench  wola.log  wola.th  wola.gt  wola.gh  wola.t  wola.m  wola.i  wola.u  wola-h"
}

# -------------------------------------------------------------
# Summary & Next Action Wizard
# -------------------------------------------------------------
post_install() {
    echo -e "\n${C_CYAN}=================================================================${C_RESET}"
    echo -e "  ${C_BOLD}${C_GREEN}✓ WOLARC INSTALLATION COMPLETE!${C_RESET}"
    echo -e "${C_CYAN}=================================================================${C_RESET}"
    echo -e "  ${C_BOLD}Commands Cheatsheet:${C_RESET}"
    echo -e "    ${C_GREEN}wola.r${C_RESET}     • Start all services (chat, embed, agent API, tunnel)"
    echo -e "    ${C_RED}wola.k${C_RESET}     • Stop all services cleanly"
    echo -e "    ${C_CYAN}wola.d${C_RESET}     • Open real-time Telemetry Dashboard"
    echo -e "    ${C_CYAN}wola.url${C_RESET}   • Remote Gateway URL, latency ping & QR code"
    echo -e "    ${C_GREEN}wola.chat${C_RESET}  • Interactive CLI chat client with note citations"
    echo -e "    ${C_YELLOW}wola.bench${C_RESET} • Measure token speed & auto-tune hardware threads"
    echo -e "    ${C_DIM}wola.log${C_RESET}   • Unified live log monitor & diagnostics"
    echo -e "    ${C_MAGENTA}wola.th${C_RESET}    • Real-time Thinking Stream Viewer"
    echo -e "    ${C_GREEN}wola.gt${C_RESET}    • Model Store & Local GGUF Importer (wola.gh)"
    echo -e "    ${C_YELLOW}wola.u${C_RESET}     • Built-in Update Checker and Installer"
    echo -e "    ${C_YELLOW}wola-h${C_RESET}     • Interactive Command Center"
    echo -e "${C_CYAN}=================================================================${C_RESET}"

    # Check if models exist in ~/models
    local models_count=$(find "$HOME/models" -maxdepth 1 -name "*.gguf" 2>/dev/null | wc -l)
    if [ "$models_count" -eq 0 ] && [ "$MODE" = "interactive" ]; then
        echo -e "\n${C_YELLOW}[!] No GGUF models detected in ~/models.${C_RESET}"
        read -p "Would you like to launch the Model Store (wola.gt / wola.gh) now to download or import one? (Y/n): " store_ans
        store_ans=${store_ans:-y}
        if [[ "$store_ans" =~ ^[Yy] ]]; then
            exec "$TARGET_BIN/wola.gt"
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

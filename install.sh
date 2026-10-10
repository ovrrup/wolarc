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
        "-y"|"--yes"|"--unattended") MODE="unattended" ;;
        "--prefix") TARGET_BIN="$2"; shift ;;
        "--help"|"-h")
            echo "Usage: bash install.sh [--upgrade | -y | --prefix <path>]"
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
    echo -e "\n${C_BOLD}${C_CYAN}[Step $1/5]${C_RESET} ${C_BOLD}$2${C_RESET}"
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
# STEP 3: Python Environment & Dependencies
# -------------------------------------------------------------
setup_python() {
    print_step "3" "Python Dependencies Setup"

    if [ -f "$DIR/requirements.txt" ]; then
        echo -e "  [*] Installing requirements (flask, flask-cors, requests, numpy, watchdog)..."
        pip install -r "$DIR/requirements.txt" --quiet
        echo -e "  ${C_GREEN}✓ Python packages installed successfully.${C_RESET}"
    fi
}

# -------------------------------------------------------------
# STEP 4: Configuration & Knowledge Directories
# -------------------------------------------------------------
setup_config() {
    print_step "4" "Configuration Setup"

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
# STEP 5: Deploy Binaries & Create Symlinks
# -------------------------------------------------------------
deploy_binaries() {
    print_step "5" "Installing Executables into ${TARGET_BIN}"

    mkdir -p "$TARGET_BIN"

    # Deploy python background scripts
    cp "$DIR/agent.py" "$HOME/agent.py"
    [ -f "$DIR/indexer.py" ] && cp "$DIR/indexer.py" "$HOME/indexer.py"

    # Deploy all commands in bin/
    for f in "$DIR/bin"/*; do
        [ -e "$f" ] || continue
        fname=$(basename "$f")
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

    # Save commit version if in git repo
    if [ -d "$DIR/.git" ]; then
        git -C "$DIR" rev-parse --short HEAD > "$HOME/.wola_version" 2>/dev/null || true
    fi

    echo -e "  ${C_GREEN}✓ Installed commands into ${TARGET_BIN}:${C_RESET}"
    echo -e "    wola.r  wola.k  wola.rs  wola.d  wola.url  wola.chat  wola.bench  wola.log  wola.th  wola.gt  wola.t  wola.m  wola.i  wola.u  wola-h"
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
    echo -e "    ${C_GREEN}wola.gt${C_RESET}    • Model Store (Download GGUF models)"
    echo -e "    ${C_YELLOW}wola.u${C_RESET}     • Built-in Update Checker and Installer"
    echo -e "    ${C_YELLOW}wola-h${C_RESET}     • Interactive Command Center"
    echo -e "${C_CYAN}=================================================================${C_RESET}"

    # Check if models exist in ~/models
    local models_count=$(find "$HOME/models" -maxdepth 1 -name "*.gguf" 2>/dev/null | wc -l)
    if [ "$models_count" -eq 0 ] && [ "$MODE" = "interactive" ]; then
        echo -e "\n${C_YELLOW}[!] No GGUF models detected in ~/models.${C_RESET}"
        read -p "Would you like to launch the Model Store (wola.gt) now to download one? (Y/n): " store_ans
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
setup_python
setup_config
deploy_binaries
post_install

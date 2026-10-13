#!/data/data/com.termux/files/usr/bin/bash

# ANSI Colors
C_RESET="\033[0m"
C_BOLD="\033[1m"
C_DIM="\033[2m"
C_RED="\033[1;31m"
C_GREEN="\033[1;32m"
C_YELLOW="\033[1;33m"
C_CYAN="\033[1;36m"
C_MAGENTA="\033[1;35m"

CONFIG="$HOME/.wolarc"
[ -f "$CONFIG" ] && source "$CONFIG"

MODELS_DIR="${MODELS_DIR:-$HOME/models}"

show_models() {
    echo "=================================================="
    echo "               WOLA MODEL INVENTORY"
    echo "=================================================="
    echo "[*] Directory: $MODELS_DIR"
    echo ""
    local count=0
    for f in "$MODELS_DIR"/*.gguf; do
        [ -e "$f" ] || continue
        count=$((count + 1))
        fname=$(basename "$f")
        size=$(du -h "$f" 2>/dev/null | cut -f1)
        tag="      "
        if [ "$fname" == "$CHAT_MODEL" ]; then
            tag="[CHAT] "
        elif [ "$fname" == "$EMBED_MODEL" ]; then
            tag="[EMBD] "
        fi
        echo "  $tag $fname ($size)"
    done

    if [ "$count" -eq 0 ]; then
        echo "  [-] No .gguf models found in $MODELS_DIR"
    fi
    echo "=================================================="
    echo "Usage:"
    echo "  wola.m chat <filename.gguf>   -> Set Chat LLM"
    echo "  wola.m embed <filename.gguf>  -> Set Embedding Model"
    echo "  wola.m import [path]          -> Import local GGUF model"
    echo "  wola.m rm [model.gguf]        -> Delete model (interactive or by filename)"
    echo "  wola.m                        -> List models"
    echo "=================================================="
}

set_model() {
    local type="$1"
    local target="$2"

    if [ ! -f "$MODELS_DIR/$target" ]; then
        echo "[-] Error: File '$target' not found in $MODELS_DIR"
        exit 1
    fi

    local key="CHAT_MODEL"
    [ "$type" == "embed" ] && key="EMBED_MODEL"

    python3 -c '
import sys, re
key, val, config_path = sys.argv[1], sys.argv[2], sys.argv[3]
try:
    with open(config_path, "r") as f:
        lines = f.readlines()
except FileNotFoundError:
    lines = []
pattern = re.compile(rf"^\s*{re.escape(key)}\s*=.*$")
found = False
new_lines = []
for line in lines:
    if pattern.match(line):
        new_lines.append(f"{key}=\"{val}\"\n")
        found = True
    else:
        new_lines.append(line)
if not found:
    new_lines.append(f"{key}=\"{val}\"\n")
with open(config_path, "w") as f:
    f.writelines(new_lines)
' "$key" "$target" "$CONFIG"
    echo "[+] Switched $type Model -> $target"
    echo "[*] Restart with 'wola.k && wola.r' to apply."
}

delete_model() {
    local target="$1"
    local force=false
    shift || true

    for arg in "$@"; do
        if [ "$arg" = "-y" ] || [ "$arg" = "--yes" ] || [ "$arg" = "-f" ]; then
            force=true
        fi
    done

    local models=()
    for f in "$MODELS_DIR"/*.gguf; do
        [ -e "$f" ] && models+=("$(basename "$f")")
    done

    if [ ${#models[@]} -eq 0 ]; then
        echo -e "${C_YELLOW}[-] No .gguf models found in ${MODELS_DIR}.${C_RESET}"
        return 0
    fi

    if [ -z "$target" ] || [ "$target" = "-y" ] || [ "$target" = "--yes" ]; then
        [ "$target" = "-y" ] || [ "$target" = "--yes" ] && force=true
        echo -e "${C_CYAN}==================================================${C_RESET}"
        echo -e "           ${C_BOLD}SELECT MODEL TO DELETE${C_RESET}"
        echo -e "${C_CYAN}==================================================${C_RESET}"
        local idx=1
        for m in "${models[@]}"; do
            local fpath="$MODELS_DIR/$m"
            local sz=$(du -h "$fpath" 2>/dev/null | cut -f1)
            local tag=""
            [ "$m" = "$CHAT_MODEL" ] && tag=" ${C_YELLOW}[ACTIVE CHAT]${C_RESET}"
            [ "$m" = "$EMBED_MODEL" ] && tag=" ${C_YELLOW}[ACTIVE EMBED]${C_RESET}"
            echo -e "  [${C_BOLD}$idx${C_RESET}] $m ($sz)$tag"
            idx=$((idx + 1))
        done
        echo -e "  [${C_DIM}0${C_RESET}] Cancel"
        echo -e "${C_CYAN}==================================================${C_RESET}"
        read -p "Select model to delete [1-${#models[@]}/0]: " sel
        if [[ "$sel" =~ ^[0-9]+$ ]] && [ "$sel" -ge 1 ] && [ "$sel" -le "${#models[@]}" ]; then
            target="${models[$((sel - 1))]}"
        else
            echo "Model deletion aborted."
            return 0
        fi
    fi

    target=$(basename "$target")
    if [[ "$target" != *.gguf ]]; then
        echo -e "${C_RED}[-] Error: '$target' is not a GGUF model file. Model deletion only applies to .gguf files.${C_RESET}"
        return 1
    fi
    local full_path="$MODELS_DIR/$target"

    if [ ! -f "$full_path" ]; then
        echo -e "${C_RED}[-] Error: Model '$target' does not exist in ${MODELS_DIR}.${C_RESET}"
        return 1
    fi

    local size_str=$(du -h "$full_path" 2>/dev/null | cut -f1)
    local is_active_chat=false
    local is_active_embed=false
    [ "$target" = "$CHAT_MODEL" ] && is_active_chat=true
    [ "$target" = "$EMBED_MODEL" ] && is_active_embed=true

    if [ "$is_active_chat" = true ] || [ "$is_active_embed" = true ]; then
        echo -e "${C_YELLOW}[!] WARNING: '$target' is currently configured as an ACTIVE model!${C_RESET}"
        [ "$is_active_chat" = true ] && echo -e "    • Configured as CHAT_MODEL in ~/.wolarc"
        [ "$is_active_embed" = true ] && echo -e "    • Configured as EMBED_MODEL in ~/.wolarc"
        echo -e "    Deleting it will clear its assignment in ~/.wolarc."
    fi

    if [ "$force" = false ]; then
        read -p "Are you sure you want to delete '$target' ($size_str)? (y/N): " conf
        if [[ ! "$conf" =~ ^[Yy] ]]; then
            echo "Deletion cancelled."
            return 0
        fi
    fi

    rm -f "$full_path"
    echo -e "${C_GREEN}[+] Deleted model: '$target' (freed $size_str).${C_RESET}"

    if [ "$is_active_chat" = true ]; then
        python3 -c "
import re, os
path = os.path.expanduser('~/.wolarc')
if os.path.exists(path):
    with open(path) as f: lines = f.readlines()
    pat = re.compile(r'^\s*CHAT_MODEL=.*$')
    lines = [f'CHAT_MODEL=\"\"\n' if pat.match(l) else l for l in lines]
    with open(path, 'w') as f: f.writelines(lines)
"
        echo -e "${C_YELLOW}[*] Unassigned CHAT_MODEL in ~/.wolarc. Please configure a new chat model before starting the node.${C_RESET}"
    fi

    if [ "$is_active_embed" = true ]; then
        python3 -c "
import re, os
path = os.path.expanduser('~/.wolarc')
if os.path.exists(path):
    with open(path) as f: lines = f.readlines()
    pat = re.compile(r'^\s*EMBED_MODEL=.*$')
    lines = [f'EMBED_MODEL=\"\"\n' if pat.match(l) else l for l in lines]
    with open(path, 'w') as f: f.writelines(lines)
"
        echo -e "${C_YELLOW}[*] Unassigned EMBED_MODEL in ~/.wolarc. Please configure a new embed model before starting the node.${C_RESET}"
    fi
}

case "$1" in
    "-h"|"--help"|"help")
        echo "=================================================="
        echo "               WOLA MODEL INVENTORY"
        echo "=================================================="
        echo "Usage:"
        echo "  wola.models                   (Casual) -> List models"
        echo "  wola.m                        (Technical)"
        echo "  wola.m chat <filename.gguf>   -> Set Chat LLM"
        echo "  wola.m embed <filename.gguf>  -> Set Embedding Model"
        echo "  wola.m import [path]          -> Import local GGUF model"
        echo "  wola.m rm [model.gguf]        -> Delete model (interactive or by filename)"
        echo "  wola.m delete [model.gguf]    -> Delete model"
        echo "=================================================="
        exit 0
        ;;
    "chat")
        [ -z "$2" ] && echo "[-] Specify model filename." && exit 1
        set_model "chat" "$2"
        ;;
    "embed")
        [ -z "$2" ] && echo "[-] Specify model filename." && exit 1
        set_model "embed" "$2"
        ;;
    "rm"|"delete"|"remove")
        shift
        delete_model "$@"
        ;;
    "import"|"add")
        shift
        if command -v wola.gt >/dev/null 2>&1; then
            exec wola.gt import "$@"
        else
            m_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
            exec "$m_dir/wola.gt" import "$@"
        fi
        ;;
    *)
        show_models
        ;;
esac

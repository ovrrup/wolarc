#!/data/data/com.termux/files/usr/bin/bash

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

case "$1" in
    "chat")
        [ -z "$2" ] && echo "[-] Specify model filename." && exit 1
        set_model "chat" "$2"
        ;;
    "embed")
        [ -z "$2" ] && echo "[-] Specify model filename." && exit 1
        set_model "embed" "$2"
        ;;
    "import"|"add")
        shift
        if command -v wola.gt >/dev/null 2>&1; then
            exec wola.gt import "$@"
        elif command -v wola.gh >/dev/null 2>&1; then
            exec wola.gh import "$@"
        else
            m_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
            exec "$m_dir/wola.gt" import "$@"
        fi
        ;;
    *)
        show_models
        ;;
esac

#!/data/data/com.termux/files/usr/bin/bash

CONFIG="$HOME/.wolarc"
[ ! -f "$CONFIG" ] && touch "$CONFIG"

if [ "$1" = "-h" ] || [ "$1" = "--help" ] || [ "$1" = "help" ]; then
    echo "=================================================="
    echo "             WOLA CONFIGURATION TUNER"
    echo "=================================================="
    echo "Usage:"
    echo "  wola.config [KEY] [VALUE]   (Casual)"
    echo "  wola.t [KEY] [VALUE]        (Technical)"
    echo ""
    echo "Examples:"
    echo "  wola.config                 -> Show all configuration variables"
    echo "  wola.t TEMPERATURE 0.7      -> Set sampling temperature to 0.7"
    echo "  wola.t CHAT_THREADS 4       -> Set CPU inference threads"
    echo "=================================================="
    exit 0
fi

if [ -z "$1" ]; then
    echo "=================================================="
    echo "             CURRENT WOLA CONFIGURATION"
    echo "=================================================="
    while IFS= read -r line || [ -n "$line" ]; do
        [[ "$line" =~ ^[[:space:]]*# ]] && continue
        [[ -z "${line// }" ]] && continue
        echo "  $line"
    done < "$CONFIG"
    echo "=================================================="
    echo "Usage:"
    echo "  wola.config [KEY] [VALUE]   (Casual)"
    echo "  wola.t [KEY] [VALUE]        (Technical)"
    echo "Example: wola.t TEMPERATURE 0.7"
    echo "         wola.t CHAT_PORT 8000"
    echo "=================================================="
    exit 0
fi

KEY=$(echo "$1" | tr '[:lower:]' '[:upper:]')
shift
VAL="$*"

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
print(f"[{'+' if found else '+'}] {'Updated' if found else 'Added'}: {key}=\"{val}\"")
' "$KEY" "$VAL" "$CONFIG"

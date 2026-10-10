#!/data/data/com.termux/files/usr/bin/bash
set -e

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PREFIX="${PREFIX:-/data/data/com.termux/files/usr}"

echo "[*] Installing Wola CLI suite into $PREFIX/bin..."
mkdir -p "$HOME/models"

# Copy python daemons
cp "$DIR/agent.py" "$HOME/agent.py"
[ -f "$DIR/indexer.py" ] && cp "$DIR/indexer.py" "$HOME/indexer.py"

# Copy binary commands
for f in "$DIR/bin"/*; do
    fname=$(basename "$f")
    cp "$f" "$PREFIX/bin/$fname"
    chmod +x "$PREFIX/bin/$fname"
done

# Create command symlinks
ln -sf "$PREFIX/bin/wolarc" "$PREFIX/bin/wola.r"
ln -sf "$PREFIX/bin/wolarc" "$PREFIX/bin/wola.k"
ln -sf "$PREFIX/bin/wolarc" "$PREFIX/bin/wola.rs"
ln -sf "$PREFIX/bin/wola-h" "$PREFIX/bin/wola.h"

# Initialize config if not existing
if [ ! -f "$HOME/.wolarc" ]; then
    cp "$DIR/config/wolarc.example" "$HOME/.wolarc"
    echo "[+] Created ~/.wolarc template"
fi

echo "[+] Installation complete. Run 'wola-h' to view available commands."

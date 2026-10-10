#!/data/data/com.termux/files/usr/bin/bash

CONFIG="$HOME/.wolarc"
[ -f "$CONFIG" ] && source "$CONFIG"

DB_FILE="${DB_FILE:-$HOME/knowledge_db.json}"
KNOWLEDGE_DIR="${KNOWLEDGE_DIR:-$HOME/storage/shared/Documents/agent_knowledge}"
INDEXER="$HOME/indexer.py"

case "$1" in
    "sync")
        echo "[*] Running knowledge base sync..."
        python "$INDEXER"
        echo "[+] Sync complete."
        ;;
    "rebuild"|"clean")
        echo "[*] Removing vector database ($DB_FILE)..."
        rm -f "$DB_FILE"
        echo "[*] Re-indexing notes from scratch..."
        python "$INDEXER"
        echo "[+] Rebuild complete."
        ;;
    "status")
        python3 -c '
import json, os, sys
p = os.path.expanduser(sys.argv[1])
if not os.path.exists(p):
    print(f"[-] Database file not found: {p}")
else:
    try:
        with open(p, "r") as f:
            db = json.load(f)
        print("==================================================")
        print("           WOLA KNOWLEDGE BASE STATUS")
        print("==================================================")
        print(f"  Database File : {p}")
        print(f"  Total Docs    : {len(db)}")
        print(f"  Total Chunks  : {sum(len(v.get(\"chunks\", [])) for v in db.values())}")
        print("--------------------------------------------------")
        for f_path, data in db.items():
            name = data.get("filename", os.path.basename(f_path))
            chunks = len(data.get("chunks", []))
            print(f"  • {name} ({chunks} chunks)")
        print("==================================================")
    except Exception as e:
        print(f"[-] Error reading database: {e}")
' "$DB_FILE"
        ;;
    "list")
        echo "[*] Source Documents in $KNOWLEDGE_DIR:"
        ls -lh "$KNOWLEDGE_DIR" 2>/dev/null || echo "[-] Directory empty or not accessible."
        ;;
    *)
        echo "Usage: wola.i [status | sync | rebuild | list]"
        ;;
esac

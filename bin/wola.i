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

DB_FILE="${DB_FILE:-$HOME/knowledge_db.json}"
KNOWLEDGE_DIR="${KNOWLEDGE_DIR:-$HOME/storage/shared/Documents/agent_knowledge}"
CHATS_DIR="${CHATS_DIR:-$HOME/storage/shared/Documents/agent_chats}"
INDEXER="$HOME/indexer.py"
AGENT_PORT="${AGENT_PORT:-5000}"

case "$1" in
    "sync")
        echo -e "${C_CYAN}[*] Running knowledge base synchronization...${C_RESET}"
        python3 "$INDEXER" --once
        echo -e "${C_GREEN}[+] Synchronization complete.${C_RESET}"
        ;;
    "add")
        shift
        TARGET="$1"
        if [ -z "$TARGET" ] || [ ! -e "$TARGET" ]; then
            echo -e "${C_RED}[-] Error: Please specify a valid file or directory to import.${C_RESET}"
            echo "    Example: wola.i add ~/Downloads/notes.md"
            exit 1
        fi
        mkdir -p "$KNOWLEDGE_DIR"
        echo -e "${C_CYAN}[*] Importing into knowledge base (${KNOWLEDGE_DIR})...${C_RESET}"
        cp -r "$TARGET" "$KNOWLEDGE_DIR/"
        fname=$(basename "$TARGET")
        echo -e "${C_GREEN}[+] Imported '${fname}' into ${KNOWLEDGE_DIR}.${C_RESET}"
        echo -e "${C_CYAN}[*] Indexing new content...${C_RESET}"
        python3 "$INDEXER" --once
        echo -e "${C_GREEN}[+] Ready! You can test it with: wola.i search \"<query>\"${C_RESET}"
        ;;
    "search"|"query")
        shift
        QUERY="$*"
        if [ -z "$QUERY" ]; then
            echo -e "${C_RED}[-] Error: Please provide a search query.${C_RESET}"
            echo "    Example: wola.i search \"What is my project roadmap?\""
            exit 1
        fi
        echo -e "${C_CYAN}=================================================================${C_RESET}"
        echo -e "                   ${C_BOLD}WOLA VECTOR RAG SEARCH${C_RESET}"
        echo -e "${C_CYAN}=================================================================${C_RESET}"
        echo -e "  Query: ${C_BOLD}\"${QUERY}\"${C_RESET}\n"

        python3 -c "
import sys, os, requests, json
import numpy as np

query = sys.argv[1]
db_file = os.path.expanduser(sys.argv[2])
embed_port = sys.argv[3]

if not os.path.exists(db_file):
    print('[-] Database file not found: ' + db_file)
    sys.exit(0)

q_raw = []
for url, payload in [
    (f'http://127.0.0.1:{embed_port}/v1/embeddings', {'input': query}),
    (f'http://127.0.0.1:{embed_port}/embedding', {'content': query}),
    (f'http://127.0.0.1:{embed_port}/embedding', {'input': query})
]:
    try:
        res = requests.post(url, json=payload, timeout=10)
        if res.status_code == 200:
            data = res.json()
            if isinstance(data, list) and len(data) > 0:
                first = data[0]
                q_raw = first.get('embedding', []) if isinstance(first, dict) else first
            elif isinstance(data, dict):
                if 'embedding' in data:
                    q_raw = data['embedding']
                elif 'data' in data and len(data['data']) > 0:
                    q_raw = data['data'][0].get('embedding', [])
            while isinstance(q_raw, list) and len(q_raw) > 0 and isinstance(q_raw[0], list):
                q_raw = q_raw[0]
            if q_raw:
                break
    except Exception:
        continue

if not q_raw:
    print(f'[-] Could not contact embedding engine (Port {embed_port}) or received empty embedding.')
    sys.exit(1)

q_vec = np.array(q_raw, dtype=np.float32).flatten()
q_norm = q_vec / np.linalg.norm(q_vec)

with open(db_file, 'r') as f:
    db = json.load(f)

vectors = []
metadata = []
dim = q_norm.size

def get_icon(fn, ftype=''):
    if ftype == 'conversation' or os.path.basename(fn).startswith('chat_'):
        return '💬'
    e = os.path.splitext(fn)[1].lower()
    if e in ['.mp3', '.wav', '.m4a', '.ogg', '.flac', '.opus', '.aac']: return '🎵'
    if e in ['.mp4', '.mkv', '.mov', '.avi', '.webm', '.flv']: return '🎬'
    if e in ['.jpg', '.jpeg', '.png', '.webp', '.gif', '.svg']: return '🖼️'
    if e in ['.py', '.go', '.rs', '.js', '.ts', '.sh', '.json', '.yaml']: return '💻'
    if e in ['.pdf', '.docx', '.epub', '.rtf', '.odt']: return '📕'
    if e in ['.pptx', '.xlsx', '.csv', '.tsv']: return '📊'
    return '📄'

for fpath, fdata in db.items():
    fname = fdata.get('filename', os.path.basename(fpath))
    ftype = fdata.get('file_type', '')
    for c in fdata.get('chunks', []):
        v = np.array(c.get('vector', []), dtype=np.float32).flatten()
        if v.size == dim and np.linalg.norm(v) > 0:
            vectors.append(v / np.linalg.norm(v))
            metadata.append({'name': fname, 'text': c.get('text', ''), 'type': ftype})

if not vectors:
    print('[-] No matching vectors in database.')
    sys.exit(0)

mat = np.vstack(vectors)
sims = np.dot(mat, q_norm)
top_idx = np.argsort(-sims)[:3]

for rank, idx in enumerate(top_idx, 1):
    score = float(sims[idx])
    item = metadata[idx]
    pct = score * 100
    color = '\033[1;32m' if score >= 0.65 else '\033[1;33m'
    m_name = item.get('name', '')
    m_type = item.get('type', '')
    icon = get_icon(m_name, m_type)
    print(f'  {rank}. {icon} [{m_name}] Match: {color}{pct:.1f}%\033[0m')
    snippet = item.get('text', '')[:220].replace('\n', ' ')
    print(f'     \033[2m\"{snippet}...\"\033[0m\n')
print('=================================================================')
" "$QUERY" "$DB_FILE" "${EMBED_PORT:-8081}"
        ;;
    "rebuild"|"clean")
        echo -e "${C_YELLOW}[*] Removing vector database ($DB_FILE)...${C_RESET}"
        rm -f "$DB_FILE"
        echo -e "${C_CYAN}[*] Re-indexing notes from scratch...${C_RESET}"
        python3 "$INDEXER" --once
        echo -e "${C_GREEN}[+] Rebuild complete.${C_RESET}"
        ;;
    "status")
        python3 -c '
import json, os, sys
p = os.path.expanduser(sys.argv[1])
kdir = os.path.expanduser(sys.argv[2])
cdir = os.path.expanduser(sys.argv[3])

if not os.path.exists(p):
    print(f"[-] Database file not found: {p}")
else:
    try:
        with open(p, "r") as f:
            db = json.load(f)
        total_chunks = sum(len(v.get("chunks", [])) for v in db.values())
        conv_docs = sum(1 for k, v in db.items() if v.get("file_type") == "conversation" or os.path.basename(k).startswith("chat_"))
        know_docs = len(db) - conv_docs
        print("==================================================")
        print("           WOLA KNOWLEDGE BASE STATUS")
        print("==================================================")
        print(f"  Database File  : {p}")
        print(f"  Knowledge Docs : {know_docs} ({kdir})")
        print(f"  Chat Memories  : {conv_docs} ({cdir})")
        print(f"  Total Chunks   : {total_chunks}")
        print("--------------------------------------------------")
        def get_icon(fn, ftype=""):
            if ftype == "conversation" or os.path.basename(fn).startswith("chat_"):
                return "💬"
            e = os.path.splitext(fn)[1].lower()
            if e in [".mp3", ".wav", ".m4a", ".ogg", ".flac", ".opus", ".aac"]: return "🎵"
            if e in [".mp4", ".mkv", ".mov", ".avi", ".webm", ".flv"]: return "🎬"
            if e in [".jpg", ".jpeg", ".png", ".webp", ".gif", ".svg"]: return "🖼️"
            if e in [".py", ".go", ".rs", ".js", ".ts", ".sh", ".json", ".yaml"]: return "💻"
            if e in [".pdf", ".docx", ".epub", ".rtf", ".odt"]: return "📕"
            if e in [".pptx", ".xlsx", ".csv", ".tsv"]: return "📊"
            return "📄"
        for f_path, data in db.items():
            name = data.get("filename", os.path.basename(f_path))
            ftype = data.get("file_type", "")
            chunks = len(data.get("chunks", []))
            print(f"  • {get_icon(name, ftype)} {name} ({chunks} chunks)")
        print("==================================================")
    except Exception as e:
        print(f"[-] Error reading database: {e}")
' "$DB_FILE" "$KNOWLEDGE_DIR" "$CHATS_DIR"
        ;;
    "list")
        echo -e "${C_CYAN}[*] Source Documents in ${KNOWLEDGE_DIR}:${C_RESET}"
        ls -lh "$KNOWLEDGE_DIR" 2>/dev/null || echo "[-] Directory empty or not accessible."
        if [ -d "$CHATS_DIR" ]; then
            echo -e "\n${C_CYAN}[*] Saved Dialogue Sessions in ${CHATS_DIR}:${C_RESET}"
            ls -lh "$CHATS_DIR" 2>/dev/null || echo "[-] No chat sessions saved yet."
        fi
        ;;
    "-h"|"--help"|"help")
        echo -e "${C_BOLD}Wola Vector Indexer & Knowledge Manager${C_RESET}"
        echo "Usage:"
        echo "  wola.knowledge status     (Casual) Inspect documents and chunk statistics"
        echo "  wola.i status             (Technical)"
        echo "  wola.i add <file|dir>     Import notes, media, audio, images or code into knowledge base"
        echo "  wola.i search \"<query>\"   Test vector similarity retrieval"
        echo "  wola.i sync               Trigger incremental scan of knowledge directory"
        echo "  wola.i rebuild            Purge database and re-vectorize from scratch"
        echo "  wola.i list               List raw source files in knowledge & chats directories"
        exit 0
        ;;
    *)
        echo -e "${C_BOLD}Wola Vector Indexer & Knowledge Manager${C_RESET}"
        echo "Usage:"
        echo "  wola.knowledge status     (Casual) Inspect documents and chunk statistics"
        echo "  wola.i status             (Technical)"
        echo "  wola.i add <file|dir>     Import notes, media, audio, images or code into knowledge base"
        echo "  wola.i search \"<query>\"   Test vector similarity retrieval"
        echo "  wola.i sync               Trigger incremental scan of knowledge directory"
        echo "  wola.i rebuild            Purge database and re-vectorize from scratch"
        echo "  wola.i list               List raw source files in knowledge & chats directories"
        ;;
esac

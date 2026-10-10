import os, time, json, requests, subprocess, shutil
from watchdog.observers import Observer
from watchdog.events import FileSystemEventHandler

CONFIG_PATH = os.path.expanduser("~/.wolarc")

def load_cfg():
    cfg = {
        "KNOWLEDGE_DIR": os.path.expanduser("~/storage/shared/Documents/agent_knowledge"),
        "DB_FILE": os.path.expanduser("~/knowledge_db.json"),
        "EMBED_PORT": "8081",
        "CHUNK_SIZE": 1200
    }
    if os.path.exists(CONFIG_PATH):
        try:
            with open(CONFIG_PATH, "r") as f:
                for line in f:
                    line = line.strip()
                    if "=" in line and not line.startswith("#"):
                        k, v = line.split("=", 1)
                        k, v = k.strip(), v.strip("\"' ")
                        if k in ["KNOWLEDGE_DIR", "DB_FILE"]:
                            cfg[k] = os.path.expanduser(os.path.expandvars(v))
                        elif k in ["EMBED_PORT"]:
                            cfg[k] = v
        except Exception:
            pass
    return cfg

CFG = load_cfg()
WATCH_DIRS = [CFG["KNOWLEDGE_DIR"]]
EMBED_URL = f"http://127.0.0.1:{CFG['EMBED_PORT']}/embedding"
DB_FILE = CFG["DB_FILE"]
CHUNK_SIZE = CFG["CHUNK_SIZE"]

def load_db():
    if os.path.exists(DB_FILE):
        try:
            with open(DB_FILE, "r") as f:
                return json.load(f)
        except Exception:
            return {}
    return {}

def save_db(data):
    tmp_file = f"{DB_FILE}.tmp"
    with open(tmp_file, "w") as f:
        json.dump(data, f, indent=2)
    os.replace(tmp_file, DB_FILE)

def extract_text(filepath):
    if filepath.endswith((".txt", ".md")):
        with open(filepath, "r", errors="ignore") as f:
            return f.read().strip()
    elif filepath.endswith(".pdf"):
        if not shutil.which("pdftotext"):
            print(f"[Warning] 'pdftotext' not found. Skipping PDF {os.path.basename(filepath)}. Run 'pkg install poppler' to enable.")
            return ""
        res = subprocess.run(["pdftotext", filepath, "-"], capture_output=True, text=True)
        return res.stdout.strip()
    return ""

def get_embedding(text):
    try:
        res = requests.post(EMBED_URL, json={"content": text}, timeout=15)
        if res.status_code == 200:
            data = res.json()
            if isinstance(data, list) and len(data) > 0:
                first = data[0]
                return first.get("embedding", []) if isinstance(first, dict) else first
            elif isinstance(data, dict):
                if "embedding" in data:
                    return data["embedding"]
                data_list = data.get("data", [])
                if data_list and isinstance(data_list[0], dict):
                    return data_list[0].get("embedding", [])
    except Exception:
        pass
    return []

def chunk_text(text, max_chars=CHUNK_SIZE):
    paragraphs = text.split("\n\n")
    chunks = []
    current = ""
    for para in paragraphs:
        if len(current) + len(para) <= max_chars:
            current += para + "\n\n"
        else:
            if current.strip():
                chunks.append(current.strip())
            current = para + "\n\n"
    if current.strip():
        chunks.append(current.strip())
    return chunks if chunks else [text]

import hashlib

def index_file(filepath):
    if not filepath.endswith((".txt", ".md", ".pdf")):
        return
    try:
        content = extract_text(filepath)
        if not content:
            return

        db = load_db()
        existing_map = {}
        if filepath in db:
            for old_c in db[filepath].get("chunks", []):
                h = old_c.get("hash") or hashlib.sha256(old_c.get("text", "").encode("utf-8")).hexdigest()
                if old_c.get("vector"):
                    existing_map[h] = old_c["vector"]

        chunks = chunk_text(content)
        records = []
        new_count = 0
        reused_count = 0

        for idx, chunk in enumerate(chunks):
            c_hash = hashlib.sha256(chunk.encode("utf-8")).hexdigest()
            if c_hash in existing_map:
                records.append({
                    "chunk_id": idx,
                    "text": chunk,
                    "hash": c_hash,
                    "vector": existing_map[c_hash]
                })
                reused_count += 1
            else:
                vector = get_embedding(chunk)
                if vector:
                    records.append({
                        "chunk_id": idx,
                        "text": chunk,
                        "hash": c_hash,
                        "vector": vector
                    })
                    new_count += 1

        if records:
            if new_count == 0 and filepath in db and len(records) == len(db[filepath].get("chunks", [])):
                return
            db[filepath] = {
                "filename": os.path.basename(filepath),
                "chunks": records,
                "updated_at": time.time()
            }
            save_db(db)
            print(f"[Indexed] {os.path.basename(filepath)} ({new_count} new, {reused_count} cached chunks)")
    except Exception as e:
        print(f"[Error] Processing {filepath}: {e}")

class ChangeHandler(FileSystemEventHandler):
    def on_modified(self, event):
        if not event.is_directory:
            time.sleep(0.5)
            index_file(event.src_path)

    def on_created(self, event):
        if not event.is_directory:
            time.sleep(0.5)
            index_file(event.src_path)

if __name__ == "__main__":
    observer = Observer()
    for d in WATCH_DIRS:
        os.makedirs(d, exist_ok=True)
        for root, _, files in os.walk(d):
            for file in files:
                index_file(os.path.join(root, file))
        observer.schedule(ChangeHandler(), d, recursive=True)

    observer.start()
    print("Watching for .txt, .md, and .pdf files...")
    try:
        while True:
            time.sleep(2)
    except KeyboardInterrupt:
        observer.stop()

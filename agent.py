import json, os, requests, time, threading, signal, subprocess, re, socket, hashlib
import numpy as np
from flask import Flask, request, Response, jsonify
from flask_cors import CORS

import sys
_HERE = os.path.dirname(os.path.abspath(__file__))
if _HERE not in sys.path:
    sys.path.insert(0, _HERE)

try:
    from web_ui import WEB_UI_HTML
except Exception:
    WEB_UI_HTML = "<h1>Wola Local AI Online</h1>"

try:
    signal.signal(signal.SIGHUP, signal.SIG_IGN)
except Exception:
    pass

app = Flask(__name__)
CORS(app, resources={r"/*": {"origins": "*"}})

CONFIG_PATH = os.path.expanduser("~/.wolarc")
THINKING_LOG = os.path.expanduser("~/.wola_thinking.log")

_THINK_LOCK = threading.Lock()
_THINK_BUFFER = []
_LAST_FLUSH = time.time()

_CACHE_LOCK = threading.Lock()
_STORE_LOCK = threading.Lock()
_CHAT_SAVE_LOCK = threading.Lock()
_SESSION_LOCK = threading.Lock()
_ACTIVE_SESSION = {"id": None, "last_activity": 0}

def get_or_create_session_id(provided_id=None):
    global _ACTIVE_SESSION
    with _SESSION_LOCK:
        now = time.time()
        if provided_id and str(provided_id).strip():
            clean_id = re.sub(r'[^a-zA-Z0-9_\-\.]', '_', str(provided_id).strip())
            _ACTIVE_SESSION["id"] = clean_id
            _ACTIVE_SESSION["last_activity"] = now
            return clean_id
        if _ACTIVE_SESSION["id"] and (now - _ACTIVE_SESSION["last_activity"] < 1800):
            _ACTIVE_SESSION["last_activity"] = now
            return _ACTIVE_SESSION["id"]
        new_id = f"session_{time.strftime('%Y%m%d_%H%M%S')}"
        _ACTIVE_SESSION["id"] = new_id
        _ACTIVE_SESSION["last_activity"] = now
        return new_id

def load_cfg():
    default_doc_dir = os.path.expanduser("~/storage/shared/Documents") if os.path.exists(os.path.expanduser("~/storage/shared/Documents")) else os.path.expanduser("~/Documents")
    cfg = {
        "EMBED_PORT": "8081",
        "CHAT_PORT": "8080",
        "AGENT_PORT": "5000",
        "CHAT_MODEL": "chat-model",
        "EMBED_MODEL": "embed-model",
        "KNOWLEDGE_DIR": os.path.join(default_doc_dir, "agent_knowledge"),
        "CHATS_DIR": os.path.join(default_doc_dir, "agent_chats"),
        "DB_FILE": os.path.expanduser("~/knowledge_db.json"),
        "METRICS_FILE": os.path.expanduser("~/.wola_metrics.json"),
        "TEMPERATURE": 0.6,
        "SIMILARITY_THRESHOLD": 0.50,
        "TOP_K": 2,
        "SYSTEM_PROMPT": "You are an intelligent, perceptive assistant. Answer using your own rich knowledge base, reasoning, and creativity. When personal background notes or memories are available, weave them naturally into your response to personalize it without robotic quotes or disclaimers.",
        "ENABLE_THINKING": True
    }
    if os.path.exists(CONFIG_PATH):
        try:
            with open(CONFIG_PATH, "r") as f:
                for line in f:
                    line = line.strip()
                    if "=" in line and not line.startswith("#"):
                        k, v = line.split("=", 1)
                        k, v = k.strip(), v.strip("\"' ")
                        if k in ["TEMPERATURE"]:
                            cfg[k] = float(v)
                        elif k in ["SIMILARITY_THRESHOLD"]:
                            cfg[k] = float(v)
                        elif k in ["TOP_K"]:
                            cfg[k] = int(v)
                        elif k in ["ENABLE_THINKING"]:
                            cfg[k] = v.lower() not in ["false", "0", "off", "no"]
                        elif k in ["DB_FILE", "METRICS_FILE", "KNOWLEDGE_DIR", "CHATS_DIR"]:
                            cfg[k] = os.path.expanduser(os.path.expandvars(v))
                        else:
                            cfg[k] = v
        except Exception:
            pass
    try:
        os.makedirs(cfg["CHATS_DIR"], exist_ok=True)
    except Exception:
        pass
    return cfg

def flush_think_tokens():
    with _THINK_LOCK:
        flush_think_tokens_locked()

def flush_think_tokens_locked():
    global _THINK_BUFFER, _LAST_FLUSH
    if not _THINK_BUFFER:
        return
    text = "".join(_THINK_BUFFER)
    _THINK_BUFFER = []
    _LAST_FLUSH = time.time()
    try:
        with open(THINKING_LOG, "a", encoding="utf-8") as f:
            f.write(text)
    except Exception:
        pass

def log_think_event(tag, content):
    flush_think_tokens()
    try:
        with open(THINKING_LOG, "a", encoding="utf-8") as f:
            f.write(f"[{tag}] {content}\n")
    except Exception:
        pass

def log_think_token(token):
    global _THINK_BUFFER, _LAST_FLUSH
    with _THINK_LOCK:
        _THINK_BUFFER.append(token)
        now = time.time()
        if now - _LAST_FLUSH > 0.08 or len(_THINK_BUFFER) >= 8:
            flush_think_tokens_locked()

def record_metrics(timings, cfg):
    try:
        metrics_file = cfg["METRICS_FILE"]
        current = {}
        if os.path.exists(metrics_file):
            with open(metrics_file, "r") as f:
                current = json.load(f)
        pred_n = timings.get("predicted_n", 0)
        prompt_n = timings.get("prompt_n", 0)
        metrics = {
            "last_updated": time.time(),
            "prompt_tokens": prompt_n,
            "predicted_tokens": pred_n,
            "prompt_per_second": round(timings.get("prompt_per_second", 0.0), 2),
            "predicted_per_second": round(timings.get("predicted_per_second", 0.0), 2),
            "total_predicted_tokens": current.get("total_predicted_tokens", 0) + pred_n,
            "total_prompt_tokens": current.get("total_prompt_tokens", 0) + prompt_n,
            "requests_completed": current.get("requests_completed", 0) + 1
        }
        with open(metrics_file, "w") as f:
            json.dump(metrics, f)
    except Exception:
        pass

_QUERY_EMBED_CACHE = {}
_VECTOR_STORE = {"mtime": 0, "matrix": None, "metadata": []}

class StreamingThinkFilter:
    """Real-time streaming filter that discards <think>...</think> blocks with zero delay on clean content."""
    def __init__(self):
        self.in_think = False
        self.pending = ""

    def process(self, token):
        if not token:
            return ""
        self.pending += token
        out = ""

        while self.pending:
            if not self.in_think:
                if "<think>" in self.pending:
                    before, after = self.pending.split("<think>", 1)
                    out += before
                    self.pending = after
                    self.in_think = True
                elif "<" in self.pending:
                    idx = self.pending.rfind("<")
                    if "<think>".startswith(self.pending[idx:]):
                        out += self.pending[:idx]
                        self.pending = self.pending[idx:]
                        break
                    else:
                        out += self.pending
                        self.pending = ""
                else:
                    out += self.pending
                    self.pending = ""
            else:
                if "</think>" in self.pending:
                    _, after = self.pending.split("</think>", 1)
                    self.pending = after
                    self.in_think = False
                elif "<" in self.pending:
                    idx = self.pending.rfind("<")
                    if "</think>".startswith(self.pending[idx:]):
                        self.pending = self.pending[idx:]
                        break
                    else:
                        self.pending = ""
                else:
                    self.pending = ""

        return out

    def flush(self):
        if not self.in_think and self.pending:
            res = self.pending
            self.pending = ""
            return res
        return ""

def extract_embedding_vector(data):
    """Safely extracts and flattens an embedding vector to a 1D list of floats."""
    if not data:
        return []
    vec = []
    if isinstance(data, list) and len(data) > 0:
        first = data[0]
        if isinstance(first, dict):
            vec = first.get("embedding", [])
        elif isinstance(first, list):
            vec = first
        elif isinstance(first, (int, float)):
            vec = data
    elif isinstance(data, dict):
        if "embedding" in data:
            vec = data["embedding"]
        elif "data" in data and isinstance(data["data"], list) and len(data["data"]) > 0:
            first_item = data["data"][0]
            if isinstance(first_item, dict):
                vec = first_item.get("embedding", [])
            elif isinstance(first_item, list):
                vec = first_item

    while isinstance(vec, list) and len(vec) > 0 and isinstance(vec[0], list):
        vec = vec[0]

    if isinstance(vec, list) and len(vec) > 0:
        try:
            return [float(x) for x in vec]
        except Exception:
            return []
    return []

def get_query_vector(text, embed_port):
    text_key = text.strip()
    if not text_key:
        return []
    with _CACHE_LOCK:
        if text_key in _QUERY_EMBED_CACHE:
            return _QUERY_EMBED_CACHE[text_key]

    clean_text = "".join(ch for ch in text_key if ch in ('\n', '\t', '\r') or ord(ch) >= 32)
    if len(clean_text) > 1500:
        clean_text = clean_text[:1500]

    candidates = [
        (f"http://127.0.0.1:{embed_port}/v1/embeddings", {"input": clean_text}),
        (f"http://127.0.0.1:{embed_port}/embedding", {"content": clean_text}),
        (f"http://127.0.0.1:{embed_port}/embedding", {"input": clean_text}),
        (f"http://127.0.0.1:{embed_port}/v1/embeddings", {"content": clean_text}),
    ]
    for url, payload in candidates:
        try:
            res = requests.post(url, json=payload, timeout=12)
            if res.status_code == 200:
                vec = extract_embedding_vector(res.json())
                if vec and len(vec) > 0 and not any(np.isnan(x) or np.isinf(x) for x in vec):
                    with _CACHE_LOCK:
                        if len(_QUERY_EMBED_CACHE) > 512:
                            _QUERY_EMBED_CACHE.pop(next(iter(_QUERY_EMBED_CACHE)))
                        _QUERY_EMBED_CACHE[text_key] = vec
                    return vec
        except Exception:
            continue
    return []

def load_vector_store(db_file):
    global _VECTOR_STORE
    if not os.path.exists(db_file):
        return None, []
    try:
        mtime = os.path.getmtime(db_file)
    except Exception:
        mtime = 0

    with _STORE_LOCK:
        if _VECTOR_STORE["matrix"] is not None and _VECTOR_STORE["mtime"] == mtime and mtime > 0:
            return _VECTOR_STORE["matrix"], _VECTOR_STORE["metadata"]

        db = None
        for _ in range(3):
            try:
                with open(db_file, "r") as f:
                    db = json.load(f)
                break
            except Exception:
                time.sleep(0.05)

        if not db or not isinstance(db, dict):
            return None, []

        vectors = []
        metadata = []
        dim = None

        for file_path, file_data in db.items():
            fname = file_data.get("filename", os.path.basename(file_path))
            ftype = file_data.get("file_type", "")
            for chunk in file_data.get("chunks", []):
                raw_v = chunk.get("vector", [])
                if not raw_v:
                    continue
                v = np.array(raw_v, dtype=np.float32).flatten()
                if v.size == 0 or np.isnan(v).any() or np.isinf(v).any():
                    continue
                if dim is None:
                    dim = v.size
                elif v.size != dim:
                    continue
                norm = np.linalg.norm(v)
                if norm > 0:
                    vectors.append(v / norm)
                    metadata.append({
                        "text": chunk.get("text", ""),
                        "filename": fname,
                        "file_path": file_path,
                        "file_type": ftype
                    })

        if vectors:
            matrix = np.vstack(vectors)
        else:
            matrix = None

        _VECTOR_STORE = {"mtime": mtime, "matrix": matrix, "metadata": metadata}
        return matrix, metadata

def _background_index_chat_turn(user_msg, assistant_msg, sess_id, chat_file, cfg):
    try:
        embed_port = cfg.get("EMBED_PORT", "8081")
        chunk_text = f"[Past Conversation Dialogue - {sess_id}]\nUser: {user_msg}\nAssistant: {assistant_msg}"
        vec = get_query_vector(chunk_text, embed_port)
        if not vec:
            return

        db_file = cfg.get("DB_FILE")
        if not db_file:
            return

        with _STORE_LOCK:
            db = {}
            if os.path.exists(db_file):
                try:
                    with open(db_file, "r") as f:
                        db = json.load(f)
                except Exception:
                    db = {}

            c_hash = hashlib.sha256(chunk_text.encode("utf-8")).hexdigest()
            fname = os.path.basename(chat_file)

            if chat_file not in db:
                db[chat_file] = {
                    "filename": fname,
                    "file_type": "conversation",
                    "file_size": os.path.getsize(chat_file) if os.path.exists(chat_file) else 0,
                    "chunks": [],
                    "updated_at": time.time()
                }

            chunks = db[chat_file].get("chunks", [])
            if not any(c.get("hash") == c_hash for c in chunks):
                chunks.append({
                    "chunk_id": len(chunks),
                    "text": chunk_text,
                    "hash": c_hash,
                    "vector": vec
                })
                db[chat_file]["chunks"] = chunks
                db[chat_file]["updated_at"] = time.time()

                tmp_db = f"{db_file}.tmp"
                with open(tmp_db, "w") as f:
                    json.dump(db, f, indent=2)
                os.replace(tmp_db, db_file)

                global _VECTOR_STORE
                _VECTOR_STORE["mtime"] = 0
    except Exception:
        pass

def save_chat_interaction(messages, assistant_reply, session_id=None, cfg=None):
    if not messages or not assistant_reply:
        return
    if cfg is None:
        cfg = load_cfg()
    chats_dir = cfg.get("CHATS_DIR")
    if not chats_dir:
        return
    try:
        os.makedirs(chats_dir, exist_ok=True)
    except Exception:
        pass

    last_user_msg = ""
    for msg in reversed(messages):
        if msg.get("role") == "user":
            c = msg.get("content", "")
            if isinstance(c, list):
                parts = [item.get("text", "") for item in c if isinstance(item, dict) and item.get("type") == "text"]
                last_user_msg = " ".join(parts).strip()
            elif isinstance(c, str):
                last_user_msg = c.strip()
            break

    if not last_user_msg:
        return

    clean_reply = re.sub(r'<think>[\s\S]*?</think>', '', assistant_reply).strip()
    if not clean_reply:
        return

    sess_id = get_or_create_session_id(session_id)
    chat_file = os.path.join(chats_dir, f"chat_{sess_id}.md")
    now_str = time.strftime("%Y-%m-%d %H:%M:%S")

    is_new = not os.path.exists(chat_file) or os.path.getsize(chat_file) == 0

    with _CHAT_SAVE_LOCK:
        try:
            with open(chat_file, "a", encoding="utf-8") as f:
                if is_new:
                    f.write(f"# Chat Session: {sess_id}\nStarted: {now_str}\n\n──────────────────────────────────────────────────────────────────\n\n")
                f.write(f"### User [{now_str}]\n{last_user_msg}\n\n### Assistant [{now_str}]\n{clean_reply}\n\n")
        except Exception:
            return

    threading.Thread(
        target=_background_index_chat_turn,
        args=(last_user_msg, clean_reply, sess_id, chat_file, cfg),
        daemon=True
    ).start()

def is_conversational_or_generic(query: str) -> bool:
    clean = re.sub(r'[^\w\s]', '', query.strip().lower())
    words = clean.split()
    if not words:
        return True
    common_short = {
        "hi", "hello", "hey", "hola", "yo", "sup", "greetings", "good morning",
        "good afternoon", "good evening", "how are you", "how are ya", "who are you",
        "what can you do", "help", "thanks", "thank you", "thx", "bye", "goodbye",
        "see you", "ok", "okay", "cool", "great", "nice", "yes", "no", "yep", "nope",
        "test", "ping"
    }
    if clean in common_short or (len(words) <= 3 and any(w in ["hi", "hello", "hey", "thanks", "bye"] for w in words)):
        return True
    return False

def search_knowledge(query, cfg):
    if is_conversational_or_generic(query):
        return []

    db_file = cfg["DB_FILE"]
    matrix, metadata = load_vector_store(db_file)
    if matrix is None or len(metadata) == 0:
        return []

    raw_q = get_query_vector(query, cfg["EMBED_PORT"])
    if not raw_q:
        return []
    q_vec = np.array(raw_q, dtype=np.float32).flatten()
    if q_vec.size == 0 or np.isnan(q_vec).any() or np.isinf(q_vec).any():
        return []
    q_norm_val = np.linalg.norm(q_vec)
    if q_norm_val == 0:
        return []
    q_norm = q_vec / q_norm_val

    if matrix.shape[1] != q_norm.size:
        log_think_event("EMBED_WARN", f"Dimension mismatch: Vector DB={matrix.shape[1]}D, Query={q_norm.size}D")
        return []

    sims = np.dot(matrix, q_norm)
    threshold = float(cfg.get("SIMILARITY_THRESHOLD", 0.50))
    effective_threshold = max(0.46, threshold)
    top_k = int(cfg.get("TOP_K", 3))

    matched_indices = np.where(sims >= effective_threshold)[0]
    if len(matched_indices) == 0:
        return []

    sorted_sub = matched_indices[np.argsort(-sims[matched_indices])]
    best_score = float(sims[sorted_sub[0]])
    if best_score < 0.46:
        return []

    results = []
    seen_texts = set()
    for idx in sorted_sub:
        score = float(sims[idx])
        if score < best_score * 0.82:
            break
        meta = metadata[idx]
        text = meta["text"].strip()
        if text in seen_texts:
            continue
        seen_texts.add(text)
        results.append({
            "score": score,
            "text": text,
            "filename": meta.get("filename", ""),
            "file_type": meta.get("file_type", "")
        })
        if len(results) >= top_k:
            break

    return results

def build_rag_prompt(messages, cfg, enable_thinking=True):
    last_user_msg = ""
    for msg in reversed(messages):
        if msg.get("role") == "user":
            content = msg.get("content", "")
            if isinstance(content, list):
                text_parts = [item.get("text", "") for item in content if isinstance(item, dict) and item.get("type") == "text"]
                last_user_msg = " ".join(text_parts)
            elif isinstance(content, str):
                last_user_msg = content
            break

    if enable_thinking:
        log_think_event("PROMPT_START", last_user_msg)
    valid_chunks = search_knowledge(last_user_msg, cfg)

    # 1. Preserve any custom system instructions provided by user/client
    user_system_prompts = []
    for msg in messages:
        if msg.get("role") == "system":
            c = msg.get("content", "")
            if isinstance(c, str) and c.strip():
                user_system_prompts.append(c.strip())

    base_system = cfg["SYSTEM_PROMPT"]
    if user_system_prompts:
        base_system = "\n\n".join(user_system_prompts)

    system_sections = [base_system]

    if not enable_thinking:
        system_sections.append(
            "\n[Thinking Directive: Direct Response Mode]\n"
            "CRITICAL INSTRUCTION: Do NOT include internal thoughts, reasoning steps, or <think>...</think> tags. "
            "Output your direct answer immediately and concisely."
        )

    if valid_chunks:
        conv_chunks = []
        doc_chunks = []
        for ch in valid_chunks:
            fname = ch["filename"]
            ftype = ch.get("file_type", "")
            if fname.startswith("chat_") or ftype == "conversation" or "[Past Conversation" in ch["text"]:
                conv_chunks.append(ch)
            else:
                doc_chunks.append(ch)

        if conv_chunks:
            if enable_thinking:
                log_think_event("RAG_MATCH", f"Found {len(conv_chunks)} past dialogue memory chunks")
            chat_lines = ["\n[Past Conversation Memory & User Context]"]
            chat_lines.append("The user had previous conversations with you. Remember these established preferences, facts, and past discussions:")
            chat_lines.append("──────────────────────────────────────────────────────────────────")
            for ch in conv_chunks:
                score, text, name = ch["score"], ch["text"], ch["filename"]
                if enable_thinking:
                    log_think_event("RAG_MATCH", f"[Chat: {name}] Relevance: {score:.3f}")
                chat_lines.append(f"• Past Exchange ({name}, Relevance: {score * 100:.1f}%):\n{text}\n")
            chat_lines.append("──────────────────────────────────────────────────────────────────")
            chat_lines.append("Guidelines for past conversation memory:")
            chat_lines.append("1. Adopt user-stated preferences, conventions, and style seamlessly.")
            chat_lines.append("2. Ground your response in previously established decisions without stating 'As we discussed before' unless relevant.")
            system_sections.append("\n".join(chat_lines))

        if doc_chunks:
            if enable_thinking:
                log_think_event("RAG_MATCH", f"Found {len(doc_chunks)} verified knowledge chunks")
            memory_lines = ["\n[Personal Knowledge & Memory Context]"]
            memory_lines.append("The user has stored the following personal background files, notes, and media. Use this information seamlessly to personalize and ground your answer:")
            memory_lines.append("──────────────────────────────────────────────────────────────────")
            for ch in doc_chunks:
                score, text, name = ch["score"], ch["text"], ch["filename"]
                if enable_thinking:
                    log_think_event("RAG_MATCH", f"[{name}] Relevance: {score:.3f}")
                memory_lines.append(f"• Source: {name} (Relevance: {score * 100:.1f}%)\n{text}\n")
            memory_lines.append("──────────────────────────────────────────────────────────────────")
            memory_lines.append("Guidelines for responding with this context:")
            memory_lines.append("1. Answer using your own foundational intelligence, logic, and reasoning as the primary driver.")
            memory_lines.append("2. Seamlessly blend the personal facts, preferences, and context above to give a tailored, personalized answer.")
            memory_lines.append("3. NEVER recite 'According to your notes' or 'Based on the provided documents' unless explicitly asked.")
            memory_lines.append("4. If the question is outside this context, answer thoroughly and creatively using your full intelligence without mentioning any missing data.")
            system_sections.append("\n".join(memory_lines))
    else:
        if enable_thinking:
            log_think_event("RAG_MATCH", "None (Operating with autonomous general intelligence)")
        system_sections.append(
            "\n[Operating Mode: Autonomous Intelligence]\n"
            "Respond using your own comprehensive knowledge, analytical reasoning, and creativity. "
            "Be articulate, direct, insightful, and helpful."
        )

    augmented_system_prompt = "\n\n".join(system_sections)

    modified_messages = [{"role": "system", "content": augmented_system_prompt}]
    for msg in messages:
        if msg.get("role") != "system":
            modified_messages.append(msg)

    return modified_messages

def get_active_tunnel_url():
    url_file = os.path.expanduser("~/.wola_url")
    if os.path.exists(url_file):
        try:
            with open(url_file, "r") as f:
                content = f.read().strip()
                if content.startswith("http"):
                    return content
        except Exception:
            pass
    # Fallback to tunnel.log
    t_log = os.path.expanduser("~/tunnel.log")
    if os.path.exists(t_log):
        try:
            import re
            with open(t_log, "r") as f:
                urls = re.findall(r"https://[a-zA-Z0-9.-]+(?:\.lhr\.life|\.serveo\.net|\.pinggy\.link|\.free\.pinggy\.link)", f.read())
                if urls:
                    return urls[-1]
        except Exception:
            pass
    return "Not active"

def get_local_interfaces():
    interfaces = []
    seen_ips = set()
    for cmd in [['ifconfig'], ['ip', '-4', 'addr']]:
        try:
            out = subprocess.check_output(cmd, stderr=subprocess.DEVNULL).decode()
            cur_iface = None
            for line in out.splitlines():
                if line and not line.startswith(' '):
                    cur_iface = line.split(':')[0].split()[0].strip()
                m = re.search(r'inet\s+(\d+\.\d+\.\d+\.\d+)', line)
                if m:
                    ip = m.group(1)
                    if ip != '127.0.0.1' and not ip.startswith('169.254.'):
                        if ip not in seen_ips:
                            seen_ips.add(ip)
                            label = 'Local Wi-Fi / LAN'
                            if cur_iface and ('ap' in cur_iface or 'swlan' in cur_iface):
                                label = 'Mobile Hotspot'
                            elif cur_iface and ('rndis' in cur_iface or 'usb' in cur_iface):
                                label = 'USB Tethering'
                            elif ip.startswith('192.168.43.'):
                                label = 'Mobile Hotspot'
                            elif ip.startswith('192.168.42.'):
                                label = 'USB Tethering'
                            interfaces.append({'iface': cur_iface or 'net', 'ip': ip, 'label': label})
            if interfaces:
                break
        except Exception:
            pass

    for probe in ['8.8.8.8', '192.168.43.254', '192.168.1.1', '10.255.255.254']:
        try:
            s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
            s.connect((probe, 80))
            ip = s.getsockname()[0]
            s.close()
            if ip and ip != '127.0.0.1' and ip not in seen_ips:
                seen_ips.add(ip)
                label = 'Local Wi-Fi / LAN'
                if ip.startswith('192.168.43.'):
                    label = 'Mobile Hotspot'
                elif ip.startswith('192.168.42.'):
                    label = 'USB Tethering'
                interfaces.append({'iface': 'sock', 'ip': ip, 'label': label})
        except Exception:
            pass

    hotspot_ip = '192.168.43.1'
    if hotspot_ip not in seen_ips:
        interfaces.append({'iface': 'hotspot', 'ip': hotspot_ip, 'label': 'Android Hotspot (Default Gateway)'})

    return interfaces

@app.route("/", methods=["GET"])
def index():
    return Response(WEB_UI_HTML, mimetype="text/html")

@app.route("/local_info", methods=["GET"])
@app.route("/v1/local_info", methods=["GET"])
def local_info():
    cfg = load_cfg()
    interfaces = get_local_interfaces()
    port = int(cfg.get("AGENT_PORT", 5000))
    hotspot_url = f"http://192.168.43.1:{port}"
    primary_url = f"http://{interfaces[0]['ip']}:{port}" if interfaces else hotspot_url
    lan_url = ""
    usb_url = ""
    for iface in interfaces:
        ip = iface["ip"]
        if ip.startswith("192.168.42.") or "rndis" in iface["iface"] or "usb" in iface["iface"]:
            usb_url = f"http://{ip}:{port}"
        elif ip != "192.168.43.1" and not lan_url:
            lan_url = f"http://{ip}:{port}"

    return jsonify({
        "status": "online",
        "port": port,
        "interfaces": interfaces,
        "primary_url": primary_url,
        "hotspot_url": hotspot_url,
        "lan_url": lan_url,
        "usb_url": usb_url,
        "api_endpoint": f"{primary_url}/v1",
        "tunnel_url": get_active_tunnel_url()
    })

@app.route("/health", methods=["GET"])
@app.route("/v1/health", methods=["GET"])
def health():
    return jsonify({"status": "ok"})

@app.route("/tunnel", methods=["GET"])
def tunnel():
    return jsonify({"tunnel_url": get_active_tunnel_url()})

@app.route("/status", methods=["GET"])
@app.route("/v1/status", methods=["GET"])
def server_status():
    cfg = load_cfg()
    db_file = cfg["DB_FILE"]
    docs = 0
    chunks = 0
    conv_docs = 0
    if os.path.exists(db_file):
        try:
            with open(db_file, "r") as f:
                db = json.load(f)
                docs = len(db)
                chunks = sum(len(v.get("chunks", [])) for v in db.values())
                conv_docs = sum(1 for k, v in db.items() if v.get("file_type") == "conversation" or os.path.basename(k).startswith("chat_"))
        except Exception:
            pass
    chats_dir = cfg.get("CHATS_DIR", "")
    saved_chats = 0
    if os.path.exists(chats_dir):
        try:
            saved_chats = len([f for f in os.listdir(chats_dir) if f.endswith(".md")])
        except Exception:
            pass

    return jsonify({
        "status": "online",
        "tunnel_url": get_active_tunnel_url(),
        "knowledge_base": {
            "documents": docs,
            "chunks": chunks,
            "conversation_memories": conv_docs
        },
        "chats": {
            "directory": chats_dir,
            "saved_sessions": saved_chats
        },
        "enable_thinking": bool(cfg.get("ENABLE_THINKING", True)),
        "temperature": cfg["TEMPERATURE"],
        "top_k": cfg["TOP_K"],
        "similarity_threshold": cfg["SIMILARITY_THRESHOLD"]
    })

@app.route("/thinking", methods=["GET", "POST"])
@app.route("/v1/thinking", methods=["GET", "POST"])
def manage_thinking():
    cfg = load_cfg()
    if request.method == "POST":
        data = request.get_json(force=True, silent=True) or {}
        val = data.get("enable_thinking", data.get("enabled", True))
        if isinstance(val, str):
            val = val.lower() not in ["false", "0", "off", "no"]
        val_str = "true" if val else "false"
        try:
            lines = []
            if os.path.exists(CONFIG_PATH):
                with open(CONFIG_PATH, "r") as f:
                    lines = f.readlines()
            pat = re.compile(r'^\s*ENABLE_THINKING=.*$')
            found = False
            new_lines = []
            for l in lines:
                if pat.match(l):
                    new_lines.append(f'ENABLE_THINKING="{val_str}"\n')
                    found = True
                else:
                    new_lines.append(l)
            if not found:
                new_lines.append(f'ENABLE_THINKING="{val_str}"\n')
            with open(CONFIG_PATH, "w") as f:
                f.writelines(new_lines)
        except Exception:
            pass
        return jsonify({"status": "ok", "enable_thinking": bool(val)})
    return jsonify({"status": "ok", "enable_thinking": bool(cfg.get("ENABLE_THINKING", True))})

@app.route("/models", methods=["GET"])
@app.route("/v1/models", methods=["GET"])
def list_models():
    cfg = load_cfg()
    models = []
    try:
        res = requests.get(f"http://127.0.0.1:{cfg['CHAT_PORT']}/v1/models", timeout=3)
        if res.status_code == 200:
            chat_data = res.json()
            if isinstance(chat_data, dict) and "data" in chat_data:
                models.extend(chat_data["data"])
    except Exception:
        pass

    chat_model = cfg.get("CHAT_MODEL", "active-chat-model") or "active-chat-model"
    if not any(m.get("id") == chat_model for m in models):
        models.append({"id": chat_model, "object": "model", "owned_by": "wola"})

    embed_model = cfg.get("EMBED_MODEL", "active-embed-model") or "active-embed-model"
    if not any(m.get("id") == embed_model for m in models):
        models.append({"id": embed_model, "object": "model", "owned_by": "wola"})

    return jsonify({"object": "list", "data": models})

@app.route("/embeddings", methods=["OPTIONS"])
@app.route("/v1/embeddings", methods=["OPTIONS"])
def embeddings_options():
    return Response("", status=204)

@app.route("/embeddings", methods=["POST"])
@app.route("/v1/embeddings", methods=["POST"])
def create_embeddings():
    cfg = load_cfg()
    body = request.get_json(force=True, silent=True) or {}
    raw_input = body.get("input", "")
    if not raw_input and "content" in body:
        raw_input = body.get("content", "")

    if isinstance(raw_input, str):
        items = [raw_input]
    elif isinstance(raw_input, list):
        items = [str(x) for x in raw_input]
    else:
        items = [str(raw_input)]

    embed_port = cfg.get("EMBED_PORT", "8081")
    data_list = []
    total_tokens = 0

    for idx, text in enumerate(items):
        vec = get_query_vector(text, embed_port)
        tok_est = max(1, len(text.split()))
        total_tokens += tok_est
        data_list.append({
            "object": "embedding",
            "index": idx,
            "embedding": vec
        })

    model_id = body.get("model") or cfg.get("EMBED_MODEL", "local-embed")
    return jsonify({
        "object": "list",
        "data": data_list,
        "model": model_id,
        "usage": {
            "prompt_tokens": total_tokens,
            "total_tokens": total_tokens
        }
    })

@app.after_request
def after_request(response):
    response.headers["Access-Control-Allow-Origin"] = "*"
    response.headers["Access-Control-Allow-Headers"] = "*"
    response.headers["Access-Control-Allow-Methods"] = "GET, POST, OPTIONS, PUT, DELETE"
    response.headers["Access-Control-Max-Age"] = "86400"
    return response

@app.route("/chat/completions", methods=["OPTIONS"])
@app.route("/v1/chat/completions", methods=["OPTIONS"])
def chat_options():
    return Response("", status=204)

@app.route("/chat/completions", methods=["POST"])
@app.route("/v1/chat/completions", methods=["POST"])
def chat_completions():
    cfg = load_cfg()
    body = request.get_json(force=True, silent=True) or {}
    messages = body.get("messages", [])
    client_stream = body.get("stream", False)
    session_id = body.get("session_id") or body.get("chat_id") or body.get("conversation_id")

    req_think = body.get("enable_thinking")
    if req_think is not None:
        if isinstance(req_think, str):
            enable_thinking = req_think.lower() not in ["false", "0", "off", "no"]
        else:
            enable_thinking = bool(req_think)
    else:
        enable_thinking = bool(cfg.get("ENABLE_THINKING", True))

    rag_messages = build_rag_prompt(messages, cfg, enable_thinking=enable_thinking)
    payload = {
        "messages": rag_messages,
        "temperature": body.get("temperature", cfg["TEMPERATURE"]),
        "max_tokens": body.get("max_tokens", 2048),
        "stream": True
    }
    for opt in ["top_p", "presence_penalty", "frequency_penalty", "stop", "tools", "tool_choice", "response_format", "functions", "function_call", "seed", "user"]:
        if opt in body:
            payload[opt] = body[opt]

    chat_url = f"http://127.0.0.1:{cfg['CHAT_PORT']}/v1/chat/completions"
    try:
        upstream = requests.post(chat_url, json=payload, stream=True, timeout=180)
    except Exception as e:
        return jsonify({"error": {"message": f"Upstream chat engine unavailable: {e}", "type": "upstream_error"}}), 502

    if enable_thinking:
        log_think_event("STREAM_START", "")

    if client_stream:
        def stream_generator():
            line_buf = b""
            last_timings = None
            stream_accum = []
            think_filter = StreamingThinkFilter() if not enable_thinking else None
            try:
                for chunk in upstream.iter_content(chunk_size=None):
                    if not chunk:
                        continue
                    line_buf += chunk
                    while b"\n" in line_buf:
                        raw_line, line_buf = line_buf.split(b"\n", 1)
                        line_str = raw_line.decode("utf-8", errors="replace").strip()
                        if line_str.startswith("data: "):
                            data_str = line_str[6:].strip()
                            if data_str == "[DONE]":
                                if not enable_thinking and think_filter:
                                    flushed = think_filter.flush()
                                    if flushed:
                                        stream_accum.append(flushed)
                                        out_pkg = json.dumps({"choices": [{"index": 0, "delta": {"content": flushed}}]})
                                        yield f"data: {out_pkg}\n\n".encode("utf-8")
                                yield b"data: [DONE]\n\n"
                                continue
                            try:
                                parsed = json.loads(data_str)
                                choices = parsed.get("choices", [])
                                if choices:
                                    delta = choices[0].get("delta", {})
                                    content_tok = delta.get("content", "")
                                    reasoning_tok = delta.get("reasoning_content", "")

                                    if enable_thinking:
                                        tok = content_tok or reasoning_tok
                                        if tok:
                                            log_think_token(tok)
                                            stream_accum.append(tok)
                                        yield raw_line + b"\n\n"
                                    else:
                                        clean_tok = think_filter.process(content_tok)
                                        if clean_tok:
                                            stream_accum.append(clean_tok)
                                            out_chunk = json.dumps({
                                                "choices": [{
                                                    "index": 0,
                                                    "delta": {"content": clean_tok}
                                                }]
                                            })
                                            yield f"data: {out_chunk}\n\n".encode("utf-8")

                                if "timings" in parsed:
                                    last_timings = parsed["timings"]
                            except Exception:
                                pass
                if enable_thinking:
                    log_think_event("STREAM_END", "")
                    flush_think_tokens()
                if last_timings:
                    record_metrics(last_timings, cfg)
                full_reply = "".join(stream_accum)
                save_chat_interaction(messages, full_reply, session_id, cfg)
            except GeneratorExit:
                try:
                    upstream.close()
                except Exception:
                    pass
                if enable_thinking:
                    log_think_event("STREAM_ABORT", "Client disconnected")
                    flush_think_tokens()
                full_reply = "".join(stream_accum)
                if full_reply:
                    save_chat_interaction(messages, full_reply, session_id, cfg)
            except Exception as e:
                try:
                    upstream.close()
                except Exception:
                    pass
                if enable_thinking:
                    log_think_event("STREAM_ERR", str(e))
                    flush_think_tokens()
            finally:
                try:
                    upstream.close()
                except Exception:
                    pass

        resp = Response(stream_generator(), content_type="text/event-stream; charset=utf-8")
        resp.headers["Cache-Control"] = "no-cache, no-transform"
        resp.headers["X-Accel-Buffering"] = "no"
        resp.headers["Connection"] = "keep-alive"
        resp.headers["Access-Control-Allow-Origin"] = "*"
        return resp

    # Aggregated fallback when client does not stream
    full_content = ""
    last_timings = None
    finish_reason = "stop"
    tool_calls_map = {}
    try:
        for raw_line in upstream.iter_lines():
            if not raw_line:
                continue
            line = raw_line.decode("utf-8", errors="replace")
            if line.startswith("data: "):
                data_str = line[6:].strip()
                if data_str != "[DONE]":
                    try:
                        parsed = json.loads(data_str)
                        choices = parsed.get("choices", [])
                        if choices:
                            c = choices[0]
                            delta = c.get("delta", {})
                            token = delta.get("content", "") or delta.get("reasoning_content", "")
                            if token:
                                full_content += token
                                if enable_thinking:
                                    log_think_token(token)
                            if c.get("finish_reason"):
                                finish_reason = c["finish_reason"]
                            tc_list = delta.get("tool_calls", [])
                            for tc in tc_list:
                                idx = tc.get("index", 0)
                                if idx not in tool_calls_map:
                                    tool_calls_map[idx] = {
                                        "id": tc.get("id", f"call_{idx}"),
                                        "type": tc.get("type", "function"),
                                        "function": {
                                            "name": tc.get("function", {}).get("name", ""),
                                            "arguments": tc.get("function", {}).get("arguments", "")
                                        }
                                    }
                                else:
                                    if "id" in tc and tc["id"]:
                                        tool_calls_map[idx]["id"] = tc["id"]
                                    fn = tc.get("function", {})
                                    if "name" in fn and fn["name"]:
                                        tool_calls_map[idx]["function"]["name"] += fn["name"]
                                    if "arguments" in fn and fn["arguments"]:
                                        tool_calls_map[idx]["function"]["arguments"] += fn["arguments"]
                        if "timings" in parsed:
                            last_timings = parsed["timings"]
                    except Exception:
                        pass
    finally:
        try:
            upstream.close()
        except Exception:
            pass

    if enable_thinking:
        log_think_event("STREAM_END", "")
        flush_think_tokens()
    if last_timings:
        record_metrics(last_timings, cfg)

    if not enable_thinking:
        full_content = re.sub(r'<think>[\s\S]*?</think>', '', full_content).strip()

    save_chat_interaction(messages, full_content, session_id, cfg)

    msg_obj = {"role": "assistant", "content": full_content}
    if tool_calls_map:
        sorted_indices = sorted(tool_calls_map.keys())
        msg_obj["tool_calls"] = [tool_calls_map[i] for i in sorted_indices]
        if finish_reason == "stop":
            finish_reason = "tool_calls"

    return jsonify({
        "choices": [{
            "index": 0,
            "message": msg_obj,
            "finish_reason": finish_reason
        }],
        "model": body.get("model") or "local-agent"
    })

if __name__ == "__main__":
    init_cfg = load_cfg()
    port = int(init_cfg["AGENT_PORT"])
    try:
        import tempfile
        from gunicorn.app.base import BaseApplication

        tmp_dir = os.environ.get("TMPDIR", tempfile.gettempdir())
        if not os.path.exists(tmp_dir) or not os.access(tmp_dir, os.W_OK):
            tmp_dir = tempfile.gettempdir()
        if not os.path.exists(tmp_dir) or not os.access(tmp_dir, os.W_OK):
            tmp_dir = os.path.expanduser("~")

        class StandaloneGunicorn(BaseApplication):
            def __init__(self, app, options=None):
                self.options = options or {}
                self.application = app
                super().__init__()
            def load_config(self):
                for k, v in self.options.items():
                    if k in self.cfg.settings and v is not None:
                        self.cfg.set(self.cfg.settings[k].name, v)
            def load(self):
                return self.application
            def init_signals(self):
                super().init_signals()
                try:
                    signal.signal(signal.SIGHUP, signal.SIG_IGN)
                except Exception:
                    pass

        opts = {
            "bind": f"0.0.0.0:{port}",
            "workers": 1,
            "threads": 8,
            "worker_class": "gthread",
            "worker_tmp_dir": tmp_dir,
            "timeout": 300,
            "graceful_timeout": 5,
            "accesslog": None,
            "errorlog": None
        }
        StandaloneGunicorn(app, opts).run()
    except Exception:
        app.run(host="0.0.0.0", port=port, threaded=True)

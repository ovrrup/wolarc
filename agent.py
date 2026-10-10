import json, os, requests, time, threading, signal
import numpy as np
from flask import Flask, request, Response, jsonify
from flask_cors import CORS

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

def load_cfg():
    cfg = {
        "EMBED_PORT": "8081",
        "CHAT_PORT": "8080",
        "AGENT_PORT": "5000",
        "DB_FILE": os.path.expanduser("~/knowledge_db.json"),
        "METRICS_FILE": os.path.expanduser("~/.wola_metrics.json"),
        "TEMPERATURE": 0.3,
        "SIMILARITY_THRESHOLD": 0.65,
        "TOP_K": 2,
        "SYSTEM_PROMPT": "Answer directly and accurately."
    }
    if os.path.exists(CONFIG_PATH):
        try:
            with open(CONFIG_PATH, "r") as f:
                for line in f:
                    line = line.strip()
                    if "=" in line and not line.startswith("#"):
                        k, v = line.split("=", 1)
                        k, v = k.strip(), v.strip("\"' ")
                        if k in ["TEMPERATURE", "SIMILARITY_THRESHOLD"]:
                            cfg[k] = float(v)
                        elif k in ["TOP_K"]:
                            cfg[k] = int(v)
                        elif k in ["DB_FILE", "METRICS_FILE"]:
                            cfg[k] = os.path.expanduser(os.path.expandvars(v))
                        else:
                            cfg[k] = v
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

def get_query_vector(text, embed_port):
    text_key = text.strip()
    if not text_key:
        return []
    if text_key in _QUERY_EMBED_CACHE:
        return _QUERY_EMBED_CACHE[text_key]

    try:
        url = f"http://127.0.0.1:{embed_port}/embedding"
        res = requests.post(url, json={"content": text}, timeout=10)
        data = res.json()
        vec = []
        if isinstance(data, list) and len(data) > 0:
            first = data[0]
            vec = first.get("embedding", []) if isinstance(first, dict) else first
        elif isinstance(data, dict):
            if "embedding" in data:
                vec = data["embedding"]
            else:
                data_list = data.get("data", [])
                if data_list and isinstance(data_list[0], dict):
                    vec = data_list[0].get("embedding", [])
        if vec:
            if len(_QUERY_EMBED_CACHE) > 256:
                _QUERY_EMBED_CACHE.pop(next(iter(_QUERY_EMBED_CACHE)))
            _QUERY_EMBED_CACHE[text_key] = vec
            return vec
    except Exception:
        pass
    return []

def load_vector_store(db_file):
    global _VECTOR_STORE
    if not os.path.exists(db_file):
        return None, []
    mtime = os.path.getmtime(db_file)
    if _VECTOR_STORE["matrix"] is not None and _VECTOR_STORE["mtime"] == mtime:
        return _VECTOR_STORE["matrix"], _VECTOR_STORE["metadata"]

    try:
        with open(db_file, "r") as f:
            db = json.load(f)
    except Exception:
        return None, []

    vectors = []
    metadata = []
    dim = None

    for file_path, file_data in db.items():
        fname = file_data.get("filename", os.path.basename(file_path))
        for chunk in file_data.get("chunks", []):
            raw_v = chunk.get("vector", [])
            if not raw_v:
                continue
            v = np.array(raw_v, dtype=np.float32).flatten()
            if dim is None:
                dim = v.size
            elif v.size != dim:
                continue
            norm = np.linalg.norm(v)
            if norm > 0:
                vectors.append(v / norm)
                metadata.append({
                    "text": chunk.get("text", ""),
                    "filename": fname
                })

    if vectors:
        matrix = np.vstack(vectors)
    else:
        matrix = None

    _VECTOR_STORE = {"mtime": mtime, "matrix": matrix, "metadata": metadata}
    return matrix, metadata

def search_knowledge(query, cfg):
    db_file = cfg["DB_FILE"]
    matrix, metadata = load_vector_store(db_file)
    if matrix is None or len(metadata) == 0:
        return []

    raw_q = get_query_vector(query, cfg["EMBED_PORT"])
    if not raw_q:
        return []
    q_vec = np.array(raw_q, dtype=np.float32).flatten()
    q_norm_val = np.linalg.norm(q_vec)
    if q_vec.size == 0 or q_norm_val == 0:
        return []
    q_norm = q_vec / q_norm_val

    if matrix.shape[1] != q_norm.size:
        return []

    # Instant vectorized similarity calculation across all chunks
    sims = np.dot(matrix, q_norm)
    threshold = cfg["SIMILARITY_THRESHOLD"]
    top_k = cfg["TOP_K"]

    matched_indices = np.where(sims >= threshold)[0]
    if len(matched_indices) == 0:
        return []

    sorted_sub = matched_indices[np.argsort(-sims[matched_indices])]
    results = []
    for idx in sorted_sub[:top_k]:
        score = float(sims[idx])
        meta = metadata[idx]
        results.append((score, meta["text"], meta["filename"]))

    return results

def build_rag_prompt(messages, cfg):
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

    log_think_event("PROMPT_START", last_user_msg)
    valid_chunks = search_knowledge(last_user_msg, cfg)
    context_text = ""
    if valid_chunks:
        context_text = "\n\n=== RELEVANT CONTEXT FROM USER NOTES ===\n"
        for score, text, name in valid_chunks:
            log_think_event("RAG_MATCH", f"[{name}] Similarity: {score:.3f}")
            context_text += f"[{name} | Match: {score:.2f}]:\n{text}\n\n"
        context_text += "=== END OF NOTES ===\n\n"
    else:
        log_think_event("RAG_MATCH", "None (Using base model intelligence)")

    modified_messages = [{"role": "system", "content": cfg["SYSTEM_PROMPT"]}]
    for msg in messages:
        if msg.get("role") != "system":
            if msg.get("content") == last_user_msg and context_text:
                modified_messages.append({
                    "role": "user",
                    "content": f"{context_text}Question: {last_user_msg}"
                })
            else:
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

@app.route("/", methods=["GET"])
def index():
    return "<h1>Local RAG Agent Online</h1>"

@app.route("/health", methods=["GET"])
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
    if os.path.exists(db_file):
        try:
            with open(db_file, "r") as f:
                db = json.load(f)
                docs = len(db)
                chunks = sum(len(v.get("chunks", [])) for v in db.values())
        except Exception:
            pass
    return jsonify({
        "status": "online",
        "tunnel_url": get_active_tunnel_url(),
        "knowledge_base": {
            "documents": docs,
            "chunks": chunks
        },
        "temperature": cfg["TEMPERATURE"],
        "top_k": cfg["TOP_K"],
        "similarity_threshold": cfg["SIMILARITY_THRESHOLD"]
    })

@app.route("/models", methods=["GET"])
@app.route("/v1/models", methods=["GET"])
def list_models():
    cfg = load_cfg()
    try:
        res = requests.get(f"http://127.0.0.1:{cfg['CHAT_PORT']}/v1/models", timeout=3)
        return (res.content, res.status_code, res.headers.items())
    except Exception:
        return jsonify({"object": "list", "data": [{"id": "active-model", "object": "model"}]})

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

    rag_messages = build_rag_prompt(messages, cfg)
    payload = {
        "messages": rag_messages,
        "temperature": body.get("temperature", cfg["TEMPERATURE"]),
        "max_tokens": body.get("max_tokens", 2048),
        "stream": True
    }
    for opt in ["top_p", "presence_penalty", "frequency_penalty", "stop"]:
        if opt in body:
            payload[opt] = body[opt]

    chat_url = f"http://127.0.0.1:{cfg['CHAT_PORT']}/v1/chat/completions"
    try:
        upstream = requests.post(chat_url, json=payload, stream=True, timeout=180)
    except Exception as e:
        return jsonify({"error": {"message": f"Upstream chat engine unavailable: {e}", "type": "upstream_error"}}), 502

    log_think_event("STREAM_START", "")

    if client_stream:
        def stream_generator():
            line_buf = b""
            last_timings = None
            try:
                for chunk in upstream.iter_content(chunk_size=None):
                    if not chunk:
                        continue
                    # Immediately yield raw chunk to client with zero buffering
                    yield chunk
                    line_buf += chunk
                    while b"\n" in line_buf:
                        raw_line, line_buf = line_buf.split(b"\n", 1)
                        line_str = raw_line.decode("utf-8", errors="replace").strip()
                        if line_str.startswith("data: "):
                            data_str = line_str[6:].strip()
                            if data_str != "[DONE]":
                                try:
                                    parsed = json.loads(data_str)
                                    choices = parsed.get("choices", [])
                                    if choices:
                                        delta = choices[0].get("delta", {})
                                        token = delta.get("content", "") or delta.get("reasoning_content", "")
                                        if token:
                                            log_think_token(token)
                                    if "timings" in parsed:
                                        last_timings = parsed["timings"]
                                except Exception:
                                    pass
                log_think_event("STREAM_END", "")
                flush_think_tokens()
                if last_timings:
                    record_metrics(last_timings, cfg)
            except GeneratorExit:
                try:
                    upstream.close()
                except Exception:
                    pass
                log_think_event("STREAM_ABORT", "Client disconnected")
                flush_think_tokens()
            except Exception as e:
                try:
                    upstream.close()
                except Exception:
                    pass
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
                            delta = choices[0].get("delta", {})
                            token = delta.get("content", "") or delta.get("reasoning_content", "")
                            if token:
                                full_content += token
                                log_think_token(token)
                        if "timings" in parsed:
                            last_timings = parsed["timings"]
                    except Exception:
                        pass
    finally:
        try:
            upstream.close()
        except Exception:
            pass

    log_think_event("STREAM_END", "")
    flush_think_tokens()
    if last_timings:
        record_metrics(last_timings, cfg)

    return jsonify({
        "choices": [{
            "index": 0,
            "message": {"role": "assistant", "content": full_content},
            "finish_reason": "stop"
        }],
        "model": "local-agent"
    })

if __name__ == "__main__":
    init_cfg = load_cfg()
    port = int(init_cfg["AGENT_PORT"])
    try:
        from gunicorn.app.base import BaseApplication
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
            "timeout": 300,
            "graceful_timeout": 5,
            "accesslog": None,
            "errorlog": None
        }
        StandaloneGunicorn(app, opts).run()
    except Exception:
        app.run(host="0.0.0.0", port=port, threaded=True)

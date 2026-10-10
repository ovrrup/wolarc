#!/data/data/com.termux/files/usr/bin/python3
import os, sys, time, json, subprocess, requests

def get_process_status(name):
    try:
        out = subprocess.check_output(["pgrep", "-f", name]).decode().strip().split("\n")
        return True, len(out), out[0]
    except Exception:
        return False, 0, "-"

def get_model_name(port):
    try:
        res = requests.get(f"http://127.0.0.1:{port}/props", timeout=1)
        data = res.json()
        model_path = data.get("default_generation_settings", {}).get("model", "")
        if not model_path:
            model_path = data.get("model_path", "")
        if model_path:
            return os.path.basename(model_path)
    except Exception:
        pass
    
    # Fallback to inspecting process command-line
    try:
        out = subprocess.check_output(["pgrep", "-a", "-f", f"port {port}"]).decode().strip()
        parts = out.split("-m ")
        if len(parts) > 1:
            return os.path.basename(parts[1].split()[0])
    except Exception:
        pass
    return "Offline / Unreachable"

def get_tunnel_url():
    tunnel_log = os.path.expanduser("~/tunnel.log")
    if os.path.exists(tunnel_log):
        try:
            with open(tunnel_log, "r") as f:
                content = f.read()
            import re
            urls = re.findall(r"https://[a-zA-Z0-9.-]+(?:\.lhr\.life|\.serveo\.net|\.localhost\.run)", content)
            if urls:
                return urls[-1]
        except Exception:
            pass
    return "No active tunnel detected"

def get_embedding_stats():
    db_file = os.path.expanduser("~/knowledge_db.json")
    if not os.path.exists(db_file):
        return 0, 0, "No database file"
    try:
        with open(db_file, "r") as f:
            db = json.load(f)
        docs_count = len(db)
        chunks_count = sum(len(v.get("chunks", [])) for v in db.values())
        return docs_count, chunks_count, "Synchronized"
    except Exception as e:
        return 0, 0, f"Error: {e}"

def get_metrics():
    metrics_file = os.path.expanduser("~/.wola_metrics.json")
    if os.path.exists(metrics_file):
        try:
            with open(metrics_file, "r") as f:
                return json.load(f)
        except Exception:
            pass
    return {
        "prompt_tokens": 0,
        "predicted_tokens": 0,
        "prompt_per_second": 0.0,
        "predicted_per_second": 0.0,
        "total_predicted_tokens": 0,
        "total_prompt_tokens": 0,
        "requests_completed": 0
    }

def get_last_indexer_log():
    log_file = os.path.expanduser("~/indexer.log")
    if os.path.exists(log_file):
        try:
            with open(log_file, "r") as f:
                lines = [line.strip() for line in f.readlines() if line.strip()]
                if lines:
                    return lines[-1][:55]
        except Exception:
            pass
    return "Idle / Waiting for note changes"

def render():
    sys.stdout.write("\033[2J\033[H")
    
    lock_file = os.path.expanduser("~/.wola.lock")
    is_on = os.path.exists(lock_file)
    state_str = "\033[1;32m[ON]\033[0m" if is_on else "\033[1;31m[OFF]\033[0m"

    chat_model = get_model_name(8080)
    embed_model = get_model_name(8081)
    tunnel_url = get_tunnel_url()
    docs, chunks, db_status = get_embedding_stats()
    metrics = get_metrics()
    indexer_activity = get_last_indexer_log()

    # Process health
    chat_ok, _, _ = get_process_status("llama-server.*8080")
    embed_ok, _, _ = get_process_status("llama-server.*8081")
    agent_ok, _, _ = get_process_status("agent.py")
    idx_ok, _, _ = get_process_status("indexer.py")

    print("=================================================================")
    print(f"               WOLA AI NODE MONITOR {state_str}")
    print("=================================================================")
    print(f"Public Base URL : \033[1;36m{tunnel_url}\033[0m")
    print(f"Chat Endpoint   : http://127.0.0.1:5000/v1/chat/completions")
    print("-----------------------------------------------------------------")
    print("  ACTIVE MODELS")
    print(f"  • Chat LLM    : \033[1;33m{chat_model}\033[0m (Port 8080)")
    print(f"  • Embedder    : \033[1;33m{embed_model}\033[0m (Port 8081)")
    print("-----------------------------------------------------------------")
    print("  LIVE INFERENCE METRICS")
    print(f"  • Generation Speed : \033[1;32m{metrics['predicted_per_second']} tokens/sec\033[0m")
    print(f"  • Prompt Eval Speed: {metrics['prompt_per_second']} tokens/sec")
    print(f"  • Last Completion  : {metrics['predicted_tokens']} tokens generated (Prompt: {metrics['prompt_tokens']})")
    print(f"  • Total Generated  : {metrics['total_predicted_tokens']} tokens ({metrics['requests_completed']} calls)")
    print("-----------------------------------------------------------------")
    print("  VECTOR DATABASE & EMBEDDING ENGINE")
    print(f"  • Watcher Daemon   : {'\033[32mActive\033[0m' if idx_ok else '\033[31mStopped\033[0m'}")
    print(f"  • Knowledge Base   : {docs} Documents | {chunks} Chunks ({db_status})")
    print(f"  • Live Event       : {indexer_activity}")
    print("-----------------------------------------------------------------")
    print("  SERVICES STATUS")
    p_chat = "\033[32mOK\033[0m" if chat_ok else "\033[31mDOWN\033[0m"
    p_emb = "\033[32mOK\033[0m" if embed_ok else "\033[31mDOWN\033[0m"
    p_agent = "\033[32mOK\033[0m" if agent_ok else "\033[31mDOWN\033[0m"
    print(f"  llama-chat: {p_chat}  |  llama-embed: {p_emb}  |  agent-api: {p_agent}")
    print("=================================================================")
    print("  Refreshing every 1s | Press Ctrl+C to exit")

def main():
    try:
        while True:
            render()
            time.sleep(1)
    except KeyboardInterrupt:
        sys.stdout.write("\nExiting Wola Dashboard.\n")
        sys.exit(0)

if __name__ == "__main__":
    main()

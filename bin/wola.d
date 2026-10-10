#!/data/data/com.termux/files/usr/bin/python3
import os, sys, time, json, subprocess, requests, select, termios, tty, atexit, re

# ANSI Color Codes & Styles
C_RESET   = "\033[0m"
C_BOLD    = "\033[1m"
C_DIM     = "\033[2m"
C_GREEN   = "\033[1;32m"
C_RED     = "\033[1;31m"
C_YELLOW  = "\033[1;33m"
C_BLUE    = "\033[1;34m"
C_MAGENTA = "\033[1;35m"
C_CYAN    = "\033[1;36m"
C_WHITE   = "\033[1;37m"
C_BG_DARK = "\033[48;5;235m"

HIDE_CURSOR = "\033[?25l"
SHOW_CURSOR = "\033[?25h"

ORIG_TERMIOS = None

def cleanup():
    global ORIG_TERMIOS
    sys.stdout.write(SHOW_CURSOR + C_RESET + "\n")
    sys.stdout.flush()
    if ORIG_TERMIOS is not None and sys.stdin.isatty():
        try:
            termios.tcsetattr(sys.stdin, termios.TCSADRAIN, ORIG_TERMIOS)
        except Exception:
            pass

atexit.register(cleanup)

def get_process_status(name):
    try:
        out = subprocess.check_output(["pgrep", "-f", name]).decode().strip().split("\n")
        return True, len(out), out[0]
    except Exception:
        return False, 0, "-"

def get_ram_info():
    total_mb, avail_mb = 0, 0
    try:
        with open("/proc/meminfo", "r") as f:
            for line in f:
                if line.startswith("MemTotal:"):
                    total_mb = int(line.split()[1]) // 1024
                elif line.startswith("MemAvailable:"):
                    avail_mb = int(line.split()[1]) // 1024
    except Exception:
        pass
    used_mb = max(0, total_mb - avail_mb)
    pct = int((used_mb / total_mb * 100)) if total_mb > 0 else 0
    return total_mb, used_mb, avail_mb, pct

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
    
    try:
        out = subprocess.check_output(["pgrep", "-a", "-f", f"port {port}"]).decode().strip()
        parts = out.split("-m ")
        if len(parts) > 1:
            return os.path.basename(parts[1].split()[0])
    except Exception:
        pass
    return "Offline / Inactive"

def get_tunnel_url():
    tunnel_log = os.path.expanduser("~/tunnel.log")
    if os.path.exists(tunnel_log):
        try:
            with open(tunnel_log, "r") as f:
                content = f.read()
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
                    return lines[-1][:46]
        except Exception:
            pass
    return "Idle • Monitoring notes"

def get_update_status():
    cache_file = os.path.expanduser("~/.wola_update_cache.json")
    if os.path.exists(cache_file):
        try:
            with open(cache_file, "r") as f:
                data = json.load(f)
            if data.get("update_available"):
                return True, data.get("remote_sha", "new")
        except Exception:
            pass
    return False, ""

def make_bar(percent, width=12, fill_char="█", empty_char="░"):
    clamped = max(0, min(100, int(percent)))
    fill_len = int((clamped / 100) * width)
    empty_len = width - fill_len
    return f"{fill_char * fill_len}{empty_char * empty_len}"

def get_uptime_str():
    lock_file = os.path.expanduser("~/.wola.lock")
    if os.path.exists(lock_file):
        try:
            mtime = os.path.getmtime(lock_file)
            diff = int(time.time() - mtime)
            hours, remainder = divmod(diff, 3600)
            mins, secs = divmod(remainder, 60)
            if hours > 0:
                return f"{hours}h {mins}m {secs}s"
            return f"{mins}m {secs}s"
        except Exception:
            pass
    return "Inactive"

def render():
    lock_file = os.path.expanduser("~/.wola.lock")
    is_on = os.path.exists(lock_file)

    chat_model = get_model_name(8080)
    embed_model = get_model_name(8081)
    tunnel_url = get_tunnel_url()
    docs, chunks, db_status = get_embedding_stats()
    metrics = get_metrics()
    indexer_activity = get_last_indexer_log()
    has_update, update_sha = get_update_status()

    tot_ram, used_ram, avail_ram, ram_pct = get_ram_info()
    ram_bar = make_bar(ram_pct, width=10)

    # Process health
    chat_ok, _, chat_pid = get_process_status("llama-server.*8080")
    embed_ok, _, embed_pid = get_process_status("llama-server.*8081")
    agent_ok, _, agent_pid = get_process_status("agent.py")
    idx_ok, _, idx_pid = get_process_status("indexer.py")
    tunnel_ok = "https://" in tunnel_url

    speed_pct = min(100, int(metrics['predicted_per_second'] * 3.5))
    speed_bar = make_bar(speed_pct, width=10)

    sys.stdout.write("\033[H")

    out = []
    out.append(f"{C_CYAN}╭────────────────────────────────────────────────────────────────╮{C_RESET}")
    
    # Title & State
    state_badge = f"{C_GREEN}● ONLINE{C_RESET}" if is_on else f"{C_RED}○ OFFLINE{C_RESET}"
    uptime_badge = f"{C_DIM}Uptime: {get_uptime_str()}{C_RESET}" if is_on else f"{C_DIM}Switch: OFF{C_RESET}"
    out.append(f"{C_CYAN}│{C_RESET}  {C_BOLD}WOLARC TELEMETRY CONSOLE{C_RESET}       {state_badge}   {uptime_badge:<18} {C_CYAN}│{C_RESET}")
    
    if has_update:
        up_text = f"⚡ UPDATE AVAILABLE ({update_sha}) • Press 'u' to upgrade"
        out.append(f"{C_CYAN}│{C_RESET}  {C_YELLOW}{C_BOLD}{up_text:<62}{C_RESET} {C_CYAN}│{C_RESET}")

    out.append(f"{C_CYAN}├────────────────────────────────────────────────────────────────┤{C_RESET}")
    
    # Remote Gateway & Tunnel
    out.append(f"{C_CYAN}│{C_RESET} {C_BOLD}PUBLIC TUNNEL & GATEWAYS{C_RESET}                                       {C_CYAN}│{C_RESET}")
    out.append(f"{C_CYAN}│{C_RESET}  Base URL  : {C_GREEN}{tunnel_url:<50}{C_RESET} {C_CYAN}│{C_RESET}")
    out.append(f"{C_CYAN}│{C_RESET}  Local API : {C_DIM}http://127.0.0.1:5000/v1/chat/completions{C_RESET}           {C_CYAN}│{C_RESET}")
    
    out.append(f"{C_CYAN}├────────────────────────────────────────────────────────────────┤{C_RESET}")

    # Active Models
    c_color = C_YELLOW if chat_ok else C_DIM
    e_color = C_YELLOW if embed_ok else C_DIM
    c_status = f"{C_GREEN}Port 8080{C_RESET}" if chat_ok else f"{C_RED}Stopped{C_RESET}"
    e_status = f"{C_GREEN}Port 8081{C_RESET}" if embed_ok else f"{C_RED}Stopped{C_RESET}"

    out.append(f"{C_CYAN}│{C_RESET} {C_BOLD}ACTIVE LOCAL MODELS{C_RESET}                                            {C_CYAN}│{C_RESET}")
    out.append(f"{C_CYAN}│{C_RESET}  • Chat LLM  : {c_color}{chat_model[:34]:<34}{C_RESET} [{c_status:<9}]   {C_CYAN}│{C_RESET}")
    out.append(f"{C_CYAN}│{C_RESET}  • Embedder  : {e_color}{embed_model[:34]:<34}{C_RESET} [{e_status:<9}]   {C_CYAN}│{C_RESET}")

    out.append(f"{C_CYAN}├────────────────────────────────────────────────────────────────┤{C_RESET}")

    # Live Inference Telemetry
    gen_spd = f"{metrics['predicted_per_second']:.1f} t/s"
    pr_spd = f"{metrics['prompt_per_second']:.1f} t/s"
    calls_tot = metrics['requests_completed']
    toks_tot = metrics['total_predicted_tokens']

    out.append(f"{C_CYAN}│{C_RESET} {C_BOLD}REAL-TIME INFERENCE METRICS{C_RESET}                                    {C_CYAN}│{C_RESET}")
    out.append(f"{C_CYAN}│{C_RESET}  • Velocity  : {C_GREEN}[{speed_bar}] {gen_spd:<10}{C_RESET} (Eval: {pr_spd})        {C_CYAN}│{C_RESET}")
    out.append(f"{C_CYAN}│{C_RESET}  • Last Call : {metrics['predicted_tokens']:<4} gen tokens | {metrics['prompt_tokens']:<4} prompt tokens             {C_CYAN}│{C_RESET}")
    out.append(f"{C_CYAN}│{C_RESET}  • Session   : {toks_tot:<6} total tokens across {calls_tot} request(s)          {C_CYAN}│{C_RESET}")

    out.append(f"{C_CYAN}├────────────────────────────────────────────────────────────────┤{C_RESET}")

    # Vector RAG Knowledge Base
    idx_state = f"{C_GREEN}Watching{C_RESET}" if idx_ok else f"{C_RED}Offline{C_RESET}"
    out.append(f"{C_CYAN}│{C_RESET} {C_BOLD}VECTOR RAG & KNOWLEDGE BASE{C_RESET}                                    {C_CYAN}│{C_RESET}")
    out.append(f"{C_CYAN}│{C_RESET}  • Documents : {C_BOLD}{docs}{C_RESET} docs | {C_BOLD}{chunks}{C_RESET} vector chunks ({db_status})              {C_CYAN}│{C_RESET}")
    out.append(f"{C_CYAN}│{C_RESET}  • Watcher   : [{idx_state}] {C_DIM}{indexer_activity:<37}{C_RESET} {C_CYAN}│{C_RESET}")

    out.append(f"{C_CYAN}├────────────────────────────────────────────────────────────────┤{C_RESET}")

    # System Health & Hardware
    out.append(f"{C_CYAN}│{C_RESET} {C_BOLD}SYSTEM HARDWARE & SUBSYSTEMS{C_RESET}                                   {C_CYAN}│{C_RESET}")
    out.append(f"{C_CYAN}│{C_RESET}  • Memory    : [{C_CYAN}{ram_bar}{C_RESET}] {ram_pct}% ({used_ram}M / {tot_ram}M)               {C_CYAN}│{C_RESET}")

    p_chat = f"{C_GREEN}OK{C_RESET}" if chat_ok else f"{C_RED}OFF{C_RESET}"
    p_emb = f"{C_GREEN}OK{C_RESET}" if embed_ok else f"{C_RED}OFF{C_RESET}"
    p_agent = f"{C_GREEN}OK{C_RESET}" if agent_ok else f"{C_RED}OFF{C_RESET}"
    p_tun = f"{C_GREEN}OK{C_RESET}" if tunnel_ok else f"{C_YELLOW}WAIT{C_RESET}"
    out.append(f"{C_CYAN}│{C_RESET}  • Nodes     : chat:{p_chat} | emb:{p_emb} | api:{p_agent} | tun:{p_tun}                  {C_CYAN}│{C_RESET}")

    out.append(f"{C_CYAN}╰────────────────────────────────────────────────────────────────╯{C_RESET}")
    out.append(f"  {C_BOLD}[q]{C_RESET} Quit   {C_BOLD}[r]{C_RESET} Redraw   {C_BOLD}[s]{C_RESET} Start/Stop   {C_BOLD}[u]{C_RESET} Update   {C_BOLD}[h]{C_RESET} Manual")

    sys.stdout.write("\n".join(out) + "\n")
    sys.stdout.flush()

def main():
    global ORIG_TERMIOS
    is_tty = sys.stdin.isatty()
    if is_tty:
        try:
            ORIG_TERMIOS = termios.tcgetattr(sys.stdin)
            tty.setcbreak(sys.stdin.fileno())
        except Exception:
            pass

    sys.stdout.write("\033[2J" + HIDE_CURSOR)
    sys.stdout.flush()

    try:
        while True:
            render()

            # Non-blocking wait for 1 second or keypress
            if is_tty:
                rlist, _, _ = select.select([sys.stdin], [], [], 1.0)
                if rlist:
                    ch = sys.stdin.read(1).lower()
                    if ch in ['q', '\x03']: # q or Ctrl+C
                        break
                    elif ch == 'r':
                        sys.stdout.write("\033[2J")
                        continue
                    elif ch == 's':
                        cleanup()
                        lock_file = os.path.expanduser("~/.wola.lock")
                        if os.path.exists(lock_file):
                            os.system("wola.k")
                        else:
                            os.system("wola.r")
                        time.sleep(1.5)
                        if is_tty and ORIG_TERMIOS:
                            tty.setcbreak(sys.stdin.fileno())
                        sys.stdout.write("\033[2J" + HIDE_CURSOR)
                    elif ch == 'u':
                        cleanup()
                        os.system("wola.u")
                        input("\nPress Enter to return to Dashboard...")
                        if is_tty and ORIG_TERMIOS:
                            tty.setcbreak(sys.stdin.fileno())
                        sys.stdout.write("\033[2J" + HIDE_CURSOR)
                    elif ch == 'h':
                        cleanup()
                        os.system("wola-h")
                        input("\nPress Enter to return to Dashboard...")
                        if is_tty and ORIG_TERMIOS:
                            tty.setcbreak(sys.stdin.fileno())
                        sys.stdout.write("\033[2J" + HIDE_CURSOR)
            else:
                time.sleep(1.0)
    except KeyboardInterrupt:
        pass
    finally:
        cleanup()
        sys.stdout.write("Dashboard detached.\n")

if __name__ == "__main__":
    main()

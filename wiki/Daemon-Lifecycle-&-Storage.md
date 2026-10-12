# Daemon Lifecycle & Storage Management ⚙️

Understand how Wolarc daemons maintain 24/7 background persistence, handle OS signals, and manage filesystem hygiene.

---

## 🔄 Daemon Startup Sequence (`wola.start` / `wola.r`)

When you run `wola.start` or `wola.r`, the system executes an atomic boot sequence:

```
wola.start
   │
   ├─► 1. Acquire Android CPU Wake-Lock (termux-wake-lock)
   ├─► 2. Ensure consolidated ~/.wola layout exists (lib, logs, run, data)
   ├─► 3. Source runtime configuration (~/.wolarc)
   ├─► 4. Acquire Global Lifecycle Mutex (~/.wola/run/wola.lock)
   ├─► 5. Launch Embedding Server (llama-server --embedding on Port 8081)
   ├─► 6. Launch Primary Chat LLM (llama-server on Port 8080)
   ├─► 7. Launch Document Vector Indexer (indexer.py)
   ├─► 8. Launch Agent API Gateway (agent.py WSGI on Port 5000)
   ├─► 9. Launch Autonomous Tunnel Watchdog Daemon (wola-tunnel-daemon)
   └─► 10. Record child process PIDs in ~/.wola/run/*.pid
```

### Process Detachment
To guarantee daemons continue running after you close Termux or your terminal emulator:
- Processes are launched via `setsid -f` or background job detachment.
- Standard input is disconnected (`< /dev/null`).
- Standard output and standard error are redirected to centralized logs in `~/.wola/logs/`.

---

## 🛑 Clean Shutdown Sequence (`wola.stop` / `wola.k`)

Shutting down services safely prevents locked databases or orphan sockets:

```bash
wola.stop         # Technical: wola.k
```

1. **Signal Trapping**: Sends `SIGTERM` to each recorded PID (`wola_chat.pid`, `wola_embed.pid`, `wola_agent.pid`, `wola_indexer.pid`, `wola_tunnel.pid`).
2. **Cooldown & Escalation**: Waits up to 3 seconds for clean exit. If a process does not exit, escalates to `SIGKILL` (`kill -9`).
3. **Lock & State Teardown**: Deletes stale PID files and unlocks `~/.wola/run/wola.lock`.
4. **Wake-Lock Release**: Calls `termux-wake-unlock` so Android OS can return to optimal power management.

---

## 🧹 Consolidated Storage Layout (`~/.wola/`)

Older versions of Wolarc placed runtime dotfiles directly in `$HOME`. Wolarc now automatically consolidates all internal files into `~/.wola/`:

| Directory | Purpose | Managed Files |
|:---|:---|:---|
| **`~/.wola/lib/`** | Script execution directory | `agent.py`, `indexer.py`, `web_ui.py` (prevents `__pycache__` in `$HOME`) |
| **`~/.wola/logs/`** | Centralized rotating logs | `server.log`, `embed.log`, `agent.log`, `indexer.log`, `tunnel.log`, `thinking.log` |
| **`~/.wola/run/`** | Process & runtime locks | `*.pid`, `wola.lock`, `wola_metrics.json`, `wola_tunnel_url` |
| **`~/.wola/data/`** | Vector database & locks | `knowledge_db.json`, `knowledge_db.json.lock` |

### Automatic Migration (`ensure_wola_layout`)
Every invocation of `wolarc` automatically checks for legacy files in `$HOME` (e.g. `~/knowledge_db.json`, `~/.wola_url`, loose logs) and migrates them safely into `~/.wola/` while reconciling paths in `~/.wolarc`.

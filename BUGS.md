# WOLARC BUG AUDIT & RESOLUTION REPORT (`BUGS.md`)

This document provides a comprehensive, centralized record of all critical, high-risk, and operational bugs identified, diagnosed, and resolved across the **Wolarc Autonomous Local AI Node** architecture. It also catalogs known system constraints and edge cases to maintain a clear roadmap for stability and performance.

---

## 1. Executive Summary

During system audits across the core services (`agent.py`, `indexer.py`, `bin/wola-tunnel-daemon`, `bin/wolarc`, `bin/wola.url`, `bin/wola.chat`, and `install.sh`), several crucial vulnerabilities and concurrency hazards were identified:
- **Vector Database Corruption & Reset Vectors**: Multi-process race conditions between the continuous file indexer (`indexer.py`) and the dialogue logging subsystem (`agent.py`), coupled with destructive fallback exception handlers that wiped the database to `{}`.
- **Stale In-Memory Data Clobbering**: Stale dictionary snapshots in `sync_knowledge_base()` overwrote newly indexed document embeddings upon deleting dead file entries.
- **Hardware Contention & Battery Thrashing**: Redundant double-embedding of markdown chat turns by both processes simultaneously on mobile CPUs.
- **Infinite Offline Cascades**: Tunnel watchdog continuously spawning SSH forwarding sessions with 56-second timeout loops when disconnected from the internet.
- **Stale Lock Deadlocks**: Crash or reboot lock persistence in `~/.wola.lock` blocking daemon startup without manual intervention.
- **Streaming Parser ANSI Leaks**: Sub-chunk token boundaries splitting `<think>` tags and corrupting CLI terminal rendering.

All identified vulnerabilities have been patched with verified unit and operational tests.

---

## 2. Solved Bugs & Mitigations

### Bug 1: Multi-Process Vector DB Race Condition & Catastrophic Reset [CRITICAL]
- **Component**: [`indexer.py`](file:///data/data/com.termux/files/home/wola/indexer.py), [`agent.py`](file:///data/data/com.termux/files/home/wola/agent.py)
- **Root Cause**: Both `indexer.py` (daemon) and `agent.py` (`_background_index_chat_turn`) concurrently performed read-modify-write operations on `~/knowledge_db.json` using a hardcoded static temporary file (`knowledge_db.json.tmp`) without cross-process advisory file locking. When both processes wrote simultaneously, the file was truncated or corrupted. The parse exception handler in `indexer.py` moved the file to `.corrupted.<timestamp>` and returned an empty dictionary `{}`, wiping out all existing document embeddings.
- **Solution**:
  - Implemented atomic cross-process file locking via `fcntl.flock` with shared reader locks (`fcntl.LOCK_SH`) and exclusive writer locks (`fcntl.LOCK_EX`) on `~/knowledge_db.json.lock`.
  - Replaced static `.tmp` filenames with collision-free temporary files (`tempfile.mkstemp` in the same filesystem directory) followed by atomic `os.replace`.
  - Added a multi-attempt retry loop (4 attempts with exponential backoff) before declaring database corruption.
- **Verification**: Verified concurrent readers and writers acquire locks without JSON parse collisions; corrupt backup triggers only on persistent read errors of non-empty files.

---

### Bug 2: Stale In-Memory DB Overwrite in `sync_knowledge_base()` [HIGH RISK]
- **Component**: [`indexer.py`](file:///data/data/com.termux/files/home/wola/indexer.py)
- **Root Cause**: `sync_knowledge_base()` loaded an in-memory snapshot `db = load_db()` at the beginning of the directory walk. During the walk, `index_file()` indexed new files and committed them to disk. At the end of the walk, `sync_knowledge_base()` evaluated deleted files against the initial `db` snapshot and called `save_db(db)`, overwriting the on-disk database with the stale snapshot and permanently erasing all newly indexed document vectors.
- **Solution**:
  - Refactored `sync_knowledge_base()` to walk and index all directories first.
  - At the pruning stage, reloaded a fresh database snapshot (`fresh_db = load_db()`) under an exclusive lock (`db_lock(exclusive=True)`) before checking and pruning non-existent files.
  - Refactored `index_file()` to compute embeddings outside the lock and acquire the lock only to merge records into the fresh database state.
- **Verification**: Verified newly added files are retained when pruning stale paths simultaneously.

---

### Bug 3: Double Indexing & Resource Contention on Chat Turns [HIGH RISK]
- **Component**: [`agent.py`](file:///data/data/com.termux/files/home/wola/agent.py), [`indexer.py`](file:///data/data/com.termux/files/home/wola/indexer.py)
- **Root Cause**: `indexer.py` watched `CHATS_DIR` using filesystem events, automatically chunking dialogue turns and computing embeddings. Concurrently, `agent.py` spawned `_background_index_chat_turn` upon every completed chat completion. This caused double embedding computation on CPU/battery, duplicated vectors, and frequent file write collisions.
- **Solution**:
  - Added `is_indexer_active()` check in `agent.py`.
  - If `indexer.py` daemon is active, `agent.py` invalidates its in-memory vector cache (`_VECTOR_STORE["mtime"] = 0`) and delegates all indexing to `indexer.py`, which deduplicates turns by sha256.
  - If `agent.py` runs in standalone mode (no indexer active), it uses `_background_index_chat_turn` with `db_lock` as a safe fallback.
- **Verification**: Verified single embedding invocation per dialogue turn when Wola node is running.

---

### Bug 4: Infinite Reconnect Thrashing & Battery Drain When Offline [HIGH RISK]
- **Component**: [`bin/wola-tunnel-daemon`](file:///data/data/com.termux/files/home/wola/bin/wola-tunnel-daemon)
- **Root Cause**: In offline environments (airplane mode, travel, local air-gapped hotspot), tunnel health checks failed. After 2 failures, `reconnect_cascade` cycled through all 4 providers (Cloudflare, localhost.run, Serveo, Pinggy) with 14-second timeouts each (~56s total), marked the provider as `"offline"`, and exited. On the next tick 20s later, `pgrep -f "ssh.*offline"` evaluated false, immediately re-triggering the entire 56s cascade. This looped indefinitely, burning mobile CPU and depleting battery.
- **Solution**:
  - Implemented `is_internet_available()` using lightweight, low-timeout WAN probes (`generate_204`, `1.1.1.1`).
  - Added power-saving exponential backoff (`30s` $\to$ `60s` $\to$ `120s` $\to$ `300s` max) when offline.
  - Guarded `reconnect_cascade` so heavy multi-provider SSH cascades only execute when WAN reachability is confirmed.
  - Updated `wola-tunnel-daemon reconnect` command to automatically ensure background watchdog worker is active.
- **Verification**: Verified watchdog enters sleep mode when disconnected and resumes automatically upon network restoration.

---

### Bug 5: Stale Lock Deadlock on Node Startup [HIGH RISK]
- **Component**: [`bin/wolarc`](file:///data/data/com.termux/files/home/wola/bin/wolarc)
- **Root Cause**: `start_node()` checked `if is_running || [ -f "$LOCK" ]`. If Android OOM killer terminated the node or the phone rebooted, `~/.wola.lock` remained on disk while processes were dead (`is_running` returned false). Subsequent `wola.start` or `wola.r` invocations aborted with `Error: Wola AI Node is ALREADY RUNNING. Switch is ON. Run 'wola.k' to turn it off first.`
- **Solution**:
  - Separated the running process check from lockfile presence.
  - If processes are not running but `~/.wola.lock` exists, log a warning, automatically remove the stale lockfile, release stale wake-locks, and proceed with startup.
- **Verification**: Verified node recovers and starts cleanly after simulated crash or ungraceful shutdown.

---

### Bug 6: Duplicate Conflicting `@app.after_request` Hooks [MEDIUM RISK]
- **Component**: [`agent.py`](file:///data/data/com.termux/files/home/wola/agent.py)
- **Root Cause**: `@app.after_request` was registered at line 24 (`add_cors_headers`) and duplicated at line 981 (`after_request`). In Flask, multiple `after_request` handlers execute in reverse registration order, causing the second hook to overwrite `Access-Control-Allow-Headers` with `*` rather than the explicit headers required by strict fetch clients and browsers.
- **Solution**:
  - Removed redundant definition at lines 981-987, consolidating all CORS header configuration in the primary top-level hook with comprehensive HTTP methods and headers.
- **Verification**: Verified single consolidated CORS response header set across all REST endpoints.

---

### Bug 7: Split-Chunk `<think>` Tag Leaks and ANSI Corruption [MEDIUM RISK]
- **Component**: [`bin/wola.chat`](file:///data/data/com.termux/files/home/wola/bin/wola.chat)
- **Root Cause**: In streaming mode, `stream_chat` matched `"<think>" in token` and `"</think>" in token` on raw received chunks. If upstream LLM token boundaries split the XML tag (e.g. `<thi` in chunk 1 and `nk>` in chunk 2), string matching failed. The raw tag leaked to stdout, the thinking state was never triggered or closed, and the terminal remained locked in magenta/dim ANSI color mode.
- **Solution**:
  - Built a sliding token buffer (`think_buf`) that inspects potential partial tag prefixes (`<`, `</`, `<thi`, `</thi`).
  - Flushed safe content while holding potential tag fragments until fully resolved or discarded.
  - Ensured ANSI state is cleanly restored in `finally` blocks even upon aborted streams.
- **Verification**: Verified split tokens across arbitrary character boundaries are parsed without tag leakage.

---

### Bug 8: Unsafe Static Temp Filename Collisions Across Threads [HIGH RISK]
- **Component**: [`agent.py`](file:///data/data/com.termux/files/home/wola/agent.py), [`indexer.py`](file:///data/data/com.termux/files/home/wola/indexer.py)
- **Root Cause**: Hardcoded temp file `f"{db_file}.tmp"` was shared across Gunicorn worker threads and background daemon tasks. Concurrent serialization attempts collided on the same file descriptor, corrupting JSON data before atomic replacement.
- **Solution**:
  - Replaced all static temp paths with `tempfile.mkstemp(prefix="...", suffix=".tmp", dir=...)` in the same directory, guaranteeing thread-safe and process-safe unique file descriptors followed by atomic `os.replace`.
- **Verification**: Verified zero file collisions under concurrent multi-threaded execution.

---

### Bug 9: Broken Direct Invocations for `wola.ts` and `wola.tailscale` [MEDIUM RISK]
- **Component**: [`bin/wola.url`](file:///data/data/com.termux/files/home/wola/bin/wola.url)
- **Root Cause**: `bin/wola.url` inspected `SELF_NAME` only for `wola.local` and `wola.hotspot`. When invoked directly via symlink as `wola.ts` or `wola.tailscale`, it fell through to default tunnel display rather than showing Tailscale peer configuration and MagicDNS URLs.
- **Solution**:
  - Added direct inspection of `SELF_NAME` for `wola.ts` and `wola.tailscale`, routing directly to `show_tailscale_gateway` and `manage_tailscale_cli`.
- **Verification**: Verified `wola.ts` and `wola.tailscale` display Tailscale mesh details directly.

---

## 3. Catalog of Bugs Found During Audit

| Bug ID | Severity | Affected File | Symptoms / Failure Mode | Status |
|---|---|---|---|---|
| **BUG-01** | CRITICAL | `indexer.py`, `agent.py` | Vector database reset to `{}` during concurrent file indexing & chat logging | **SOLVED** |
| **BUG-02** | HIGH | `indexer.py` | New document embeddings erased by stale snapshot in `sync_knowledge_base()` | **SOLVED** |
| **BUG-03** | HIGH | `agent.py`, `indexer.py` | Double embedding compute and lock collisions per dialogue turn | **SOLVED** |
| **BUG-04** | HIGH | `bin/wola-tunnel-daemon` | Infinite 56s SSH timeout loops burning CPU and battery when offline | **SOLVED** |
| **BUG-05** | HIGH | `bin/wolarc` | Deadlock on startup if `~/.wola.lock` exists after device reboot or crash | **SOLVED** |
| **BUG-06** | MEDIUM | `agent.py` | Duplicate `@app.after_request` causing inconsistent CORS headers | **SOLVED** |
| **BUG-07** | MEDIUM | `bin/wola.chat` | Split-chunk `<think>` tags leaking and corrupting terminal ANSI formatting | **SOLVED** |
| **BUG-08** | HIGH | `agent.py`, `indexer.py` | Thread collision on hardcoded `knowledge_db.json.tmp` filename | **SOLVED** |
| **BUG-09** | MEDIUM | `bin/wola.url` | Direct invocation of `wola.ts` / `wola.tailscale` falling through to public URL | **SOLVED** |
| **BUG-10** | HIGH | Entire Suite | System root pollution: 17+ loose scripts, logs, dotfiles and `__pycache__` in `$HOME` | **SOLVED** |
| **BUG-11** | HIGH | `bin/wola.m`, `bin/wola.i` | Lack of granular model, document, and chat session deletion commands causing disk bloat | **SOLVED** |
| **BUG-12** | MEDIUM | `bin/wola.log`, `bin/wola.purge` | Inability to truncate bloated logs or selectively purge caches/chats without full wipe | **SOLVED** |

---

## 4. Known Bugs, Edge Cases & Operational Constraints

The following items are documented operational constraints and edge cases tracked for ongoing maintenance:

### 1. Android OEM Aggressive Battery Optimization & Phantom Process Killer
- **Description**: Android 12+ features a Phantom Process Killer that caps child processes per application (typically 32) and kills processes consuming excessive CPU in the background when Termux is not the foreground app.
- **Mitigation**:
  - `wolarc` automatically requests `termux-wake-lock` during startup.
  - Users should disable "Battery Optimization" for Termux in Android Settings $\to$ Apps $\to$ Termux $\to$ Battery $\to$ Unrestricted.

### 2. Android Shared Storage Permission Disconnection
- **Description**: If `KNOWLEDGE_DIR` points to `/sdcard/Documents/agent_knowledge` and Android revokes storage permission, file indexing fails with `PermissionDenied`.
- **Mitigation**: Run `termux-setup-storage` if using shared internal storage, or keep knowledge and chats within `$HOME` (e.g. `~/knowledge` and `~/chats`).

### 3. High Context OOM on Low-Memory Devices (RAM $\le$ 4 GB)
- **Description**: Running chat models with large context windows (`CHAT_CTX=4096` or higher) alongside an embedding model (`EMBED_CTX=2048`) can exceed available RAM on 3GB/4GB Android devices, triggering kernel OOM.
- **Mitigation**: Use `wola.config` or edit `~/.wolarc` to set `CHAT_CTX=2048` and use Q4_K_M quantizations.

### 4. Cloudflare Quick Tunnel Rate Limiting & Captive Portals
- **Description**: Quick Tunnels (`trycloudflare.com`) occasionally rate-limit connections or fail on hotel/airport captive portals where DNS is intercepted.
- **Mitigation**: The tunnel watchdog automatically cascades to `localhost.run`, `serveo.net`, and `pinggy.io`. For 100% reliable air-gapped connectivity, use `wola.hotspot` or `wola.tailscale`.

### 5. Termux Inotify Limit on Deeply Nested Trees
- **Description**: The native Linux kernel inotify watcher used by `watchdog` has a limit (`max_user_watches`). Indexing gigantic directory structures (e.g. containing `node_modules` or `.git`) can exhaust watches.
- **Mitigation**: `should_ignore_file()` in `indexer.py` explicitly skips `.git`, `node_modules`, `__pycache__`, and temporary/backup files.

---

## 5. Verification & Testing Matrix

| Test Case | Target Script | Test Command | Result |
|---|---|---|---|
| Python Syntax Validation | `agent.py`, `indexer.py`, `bin/wola.chat` | `python3 -m py_compile <file>` | **PASS (0 errors)** |
| Bash Syntax Validation | `bin/wolarc`, `bin/wola-tunnel-daemon`, `bin/wola.url` | `bash -n <file>` | **PASS (0 errors)** |
| Multi-Process DB Locking | `indexer.py`, `agent.py` | Concurrent `fcntl.flock` read/write tests | **PASS (Zero collisions)** |
| Stale Lock Recovery | `bin/wolarc` | Startup with dead PID and active `~/.wola.lock` | **PASS (Auto-cleared)** |
| Offline Tunnel Backoff | `bin/wola-tunnel-daemon` | Simulation with unreachable WAN gateway | **PASS (Power-save backoff)** |
| Split `<think>` Streaming | `bin/wola.chat` | Fragmented token boundary unit test | **PASS (Clean ANSI transition)** |
| Unified Command Pairs | `bin/wolarc`, `install.sh` | Technical $\leftrightarrow$ Casual symlink verification | **PASS (1-to-1 parity)** |

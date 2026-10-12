# Troubleshooting & FAQ 🩺

Answers to common operational questions, error codes, and step-by-step diagnostic fixes.

---

## ❓ Frequently Asked Questions

### Q1: Why did Termux kill the model process during generation?
**Answer**: Android's Out-Of-Memory (OOM) killer terminates processes that exceed the memory limit allocated to Termux.
- **Fix 1**: Use a lightweight model optimized for mobile RAM (e.g. `Llama 3.2 1B Q4_K_M` or `Qwen 2.5 1.5B Q4_K_M`).
- **Fix 2**: Reduce context size via `wola.t CHAT_CTX 2048`.
- **Fix 3**: Close memory-heavy background apps on Android before starting inference.

---

### Q2: How do I fix "Port 8080 / 8081 / 5000 already in use"?
**Answer**: A previous instance did not shut down cleanly.
1. Run clean stop:
   ```bash
   wola.stop        # or wola.k
   ```
2. If stubborn processes remain:
   ```bash
   pkill -f llama-server
   pkill -f agent.py
   ```
3. Restart cleanly:
   ```bash
   wola.start       # or wola.r
   ```

---

### Q3: How do I prevent Android from sleeping when the screen turns off?
**Answer**: Android battery managers suspend background CPU execution.
1. In Android Settings -> Apps -> Termux -> **Battery** -> Select **Unrestricted**.
2. Run `termux-wake-lock` (Wolarc acquires this automatically on `wola.start`).
3. Swipe down from Android notifications and verify the "Termux Wakelock Acquired" persistent notification is visible.

---

### Q4: Why is my remote tunnel URL showing "Connection Refused"?
**Answer**:
1. Check if the local API Gateway is running:
   ```bash
   curl -I http://127.0.0.1:5000/health
   ```
2. If the gateway is running, trigger a tunnel provider switch:
   ```bash
   wola.url switch cloudflare
   # Or switch to an SSH provider:
   wola.url switch localhost.run
   ```

---

### Q5: How do I clear log files that are taking up storage?
**Answer**: Run the log truncation command:
```bash
wola.log clear       # or wola.log clean
```
This zeroes out all files in `~/.wola/logs/` and reports the reclaimed disk space.

---

### Q6: How do I recover from a stale lock error?
**Answer**: Run cache purging:
```bash
wola.purge --cache
```
This safely removes stale PID files, temporary socket locks, and Python bytecaches without touching your downloaded models or vector databases.

---

### Q7: Why are reasoning tokens (`<think>`) appearing or not appearing?
**Answer**: You can toggle thinking output dynamically:
```bash
wola.think on       # Enable reasoning output
wola.think off      # Disable reasoning output for maximum generation velocity
wola.think status   # Check current status
```
In the Go TUI (`wola`), press `Ctrl+T` in the Chat view to collapse or expand thinking blocks.

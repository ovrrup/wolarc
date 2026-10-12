# Hardware Tuning & Benchmarks ⚡

Maximize generation throughput, minimize thermal throttling, and optimize battery efficiency on mobile ARM64 and edge Linux hardware.

---

## 📱 Mobile CPU Architecture: Big vs. LITTLE Clusters

Most modern Android processors (Qualcomm Snapdragon, MediaTek Dimensity, Google Tensor) employ heterogeneous multi-processing:
- **High-Performance "Big" Cores** (Cortex-X / Cortex-A7xx): Designed for sustained high-IPC compute.
- **Power-Efficient "LITTLE" Cores** (Cortex-A5xx): Designed for low-power background tasks.

> [!IMPORTANT]
> **The Thread Trapping Rule**:
> Setting CPU threads (`CHAT_THREADS`) to all 8 cores on an 8-core mobile CPU will **slow down** inference! Threads scheduled on LITTLE cores will bottleneck synchronized matrix multiplications across big cores.
> **Optimal setting**: Use exactly the number of big cores on your chip (typically **`4`**).

---

## 🏎️ Running the Automated Benchmark (`wola.bench`)

Wolarc includes an automated hardware benchmark suite:

```bash
wola.bench        # Technical: wola.b
```

### What it does:
1. Performs prompt evaluation (prefill velocity) and token generation passes across 1, 2, 4, 6, and 8 thread allocations.
2. Measures tokens per second (`tok/s`), RAM pressure, and thermals.
3. Automatically identifies the peak efficiency sweet spot.
4. Offers to save the optimal thread count directly into `~/.wolarc`.

---

## 📊 Recommended Mobile Models & RAM Footprint

| Model Name | Parameter Count | Quantization | RAM Usage | Speed (Snapdragon 8 Gen 2/3) |
|:---|:---|:---|:---|:---|
| **Llama 3.2 1B** | 1.23 Billion | Q4_K_M | ~850 MB | 25–35 tok/s |
| **Qwen 2.5 1.5B**| 1.54 Billion | Q4_K_M | ~1.1 GB | 20–28 tok/s |
| **Gemma 2 2B** | 2.61 Billion | Q4_K_M | ~1.6 GB | 18–24 tok/s |
| **Phi-4 Mini** | 3.82 Billion | Q4_K_M | ~2.5 GB | 10–16 tok/s |
| **BGE-Small v1.5**| 33 Million (Embed)| Q4_K_M | ~150 MB | Instant (<10ms/chunk) |

---

## 🎛️ Optimization Knobs

Tune these in `~/.wolarc` via `wola.t`:

### 1. Flash Attention (`FLASH_ATTENTION`)
```bash
wola.t FLASH_ATTENTION auto
```
- Drastically reduces memory bandwidth consumption during prompt evaluation and KV cache updates.

### 2. Context Window (`CHAT_CTX`)
```bash
wola.t CHAT_CTX 4096      # 4K context (balanced)
wola.t CHAT_CTX 2048      # 2K context (conserves ~400MB RAM on 4GB devices)
```

### 3. Parallel Execution Slots (`CHAT_PARALLEL`)
```bash
wola.t CHAT_PARALLEL 2
```
- Configures `llama-server` with 2 parallel slots. Prevents remote client queries from locking out local indexing or interactive terminal queries.

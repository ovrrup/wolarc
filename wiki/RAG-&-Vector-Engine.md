# Universal RAG & Vector Memory Engine 🧠

Wolarc features a continuous, multi-format Retrieval-Augmented Generation (RAG) and semantic memory engine designed to operate autonomously with zero external cloud API dependencies.

---

## 📥 1. Multi-Format Ingestion Pipeline

The background indexer (`indexer.py`) watches `~/Documents/wola/knowledge/` and `~/Documents/wola/chats/` for changes, automatically parsing and vectorizing files into mathematical embedding spaces.

| File Category | Supported Formats | Extraction Mechanism |
|:---|:---|:---|
| **Plain Text & Code** | `.md`, `.txt`, `.py`, `.go`, `.sh`, `.json`, `.csv`, `.yml` | Direct UTF-8 stream reader with line boundary chunking |
| **Documents** | `.pdf` | `pypdf` / `pdfminer` page text extraction |
| **Word Processing** | `.docx` | `python-docx` XML paragraph & table parser |
| **E-books & Publications**| `.epub` | `ebooklib` HTML chapter and metadata parser |
| **Audio Tracks & Notes**| `.mp3`, `.wav`, `.m4a`, `.ogg`, `.flac` | ID3 metadata, audio duration, embedded lyrics & transcript text |
| **Video Recordings** | `.mp4`, `.mkv`, `.mov` | Container metadata, chapters, track titles, and dialogue tracks |
| **Images & Diagrams** | `.jpg`, `.jpeg`, `.png`, `.webp` | EXIF camera metadata, timestamps, location, and OCR text fallback |

---

## ✂️ 2. Chunking, Overlap & Deduplication

### Chunking Strategy
To optimize retrieval accuracy without fragmenting conceptual context:
- **Chunk Size**: 500 characters (~100–130 tokens).
- **Chunk Overlap**: 100 characters (~25 tokens) sliding window to preserve continuity across boundaries.
- **Natural Boundary Snapping**: Splits on paragraph breaks (`\n\n`), sentence terminators (`. `, `? `, `! `), and newline boundaries before hard character truncation.

### SHA-256 Deduplication
Before generating computationally heavy embeddings:
1. Indexer computes a SHA-256 hash of the entire file content.
2. If the hash matches the database record, the file is skipped instantly.
3. If the hash differs, previous chunks are purged, and the updated content is re-indexed atomically.

---

## 💾 3. Vector Database Architecture (`knowledge_db.json`)

The vector database is stored as human-auditable JSON at `~/.wola/data/knowledge_db.json`.

### Schema Structure
```json
{
  "docs": {
    "/home/user/Documents/wola/knowledge/roadmap.md": {
      "hash": "a1b2c3d4e5f6...",
      "mtime": 1728800000.0,
      "chunks": [
        {
          "id": "a1b2c3d4_0",
          "text": "Milestone 1 focuses on multi-transport connection fabric...",
          "vector": [0.0213, -0.0542, 0.1104, "... 384 dimensions ..."],
          "source": "roadmap.md",
          "chunk_index": 0
        }
      ]
    }
  }
}
```

### Multi-Process Concurrency & Advisory Locking
To prevent database corruption between concurrent processes (`indexer.py` updating files and `agent.py` logging chat turns):
- **Advisory File Lock (`fcntl.flock`)**:
  - `fcntl.LOCK_SH` (Shared Lock): Acquired during vector searches and reads.
  - `fcntl.LOCK_EX` (Exclusive Lock): Acquired during chunk updates, inserts, and pruning.
- **Atomic File Swapping**: Updates write to a temporary file in `~/.wola/data/` via `tempfile.mkstemp` and replace the target atomically via `os.replace`.

---

## 📐 4. Cosine Similarity & Dynamic Thresholding

When a user submits a query to `/v1/chat/completions`:
1. The query text is vectorized by the embedding server (`Port 8081`).
2. The query vector $\vec{Q}$ is scored against every stored chunk vector $\vec{D}_i$ via **Cosine Similarity**:

$$\text{Similarity}(\vec{Q}, \vec{D}_i) = \frac{\vec{Q} \cdot \vec{D}_i}{\|\vec{Q}\| \|\vec{D}_i\|}$$

3. Chunks with scores below `SIMILARITY_THRESHOLD` (default: `0.50`) are rejected.
4. Top matching chunks up to `TOP_K` (default: `2`) are ranked and formatted for context injection.

---

## 🧬 5. Cognitive Memory Blending

Traditional RAG setups produce robotic answers ("According to document X, ..."). Wolarc takes a cognitive blending approach:

### System Prompt Structure
```markdown
[SYSTEM INSTRUCTION]
You are Wola, a private autonomous local AI companion.
You possess sharp intelligence, reasoning independently.

[LONG-TERM MEMORY CONTEXT]
The following background context reflects verified knowledge and personal memory:
---
[Chunk 1: roadmap.md] (Score: 0.82)
Milestone 1 focuses on multi-transport connection fabric across mDNS, Tailscale, and Cloudflare.
---

[CONVERSATION TURNS]
User: What are we building next for host connectivity?
Assistant: <think>
User is asking about upcoming host connection improvements. Retrieved memory references Milestone 1: multi-transport connection fabric with mDNS, Tailscale, and Cloudflare.
</think>We are working on Milestone 1: a resilient connection fabric featuring local mDNS auto-discovery, Tailscale WireGuard mesh networking, and named Cloudflare tunnels.
```

The model treats retrieved context as **its own ingrained memories**, delivering direct, intelligent responses without repetitive citation disclaimers.

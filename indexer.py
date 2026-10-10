import os, sys, time, json, subprocess, shutil, hashlib, struct, re, zlib, zipfile
import urllib.request

try:
    import requests
except ImportError:
    requests = None

try:
    from watchdog.observers import Observer
    from watchdog.events import FileSystemEventHandler
    HAS_WATCHDOG = True
except ImportError:
    Observer = None
    FileSystemEventHandler = object
    HAS_WATCHDOG = False

CONFIG_PATH = os.path.expanduser("~/.wolarc")

def load_cfg():
    default_doc_dir = os.path.expanduser("~/storage/shared/Documents") if os.path.exists(os.path.expanduser("~/storage/shared/Documents")) else os.path.expanduser("~/Documents")
    cfg = {
        "KNOWLEDGE_DIR": os.path.join(default_doc_dir, "agent_knowledge"),
        "CHATS_DIR": os.path.join(default_doc_dir, "agent_chats"),
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
                        if k in ["KNOWLEDGE_DIR", "CHATS_DIR", "DB_FILE"]:
                            cfg[k] = os.path.expanduser(os.path.expandvars(v))
                        elif k in ["EMBED_PORT"]:
                            cfg[k] = v
                        elif k in ["CHUNK_SIZE"]:
                            cfg[k] = int(v)
        except Exception:
            pass
    for d in [cfg["KNOWLEDGE_DIR"], cfg["CHATS_DIR"]]:
        try:
            os.makedirs(d, exist_ok=True)
        except Exception:
            pass
    return cfg

CFG = load_cfg()
WATCH_DIRS = [CFG["KNOWLEDGE_DIR"], CFG["CHATS_DIR"]]
EMBED_URL = f"http://127.0.0.1:{CFG['EMBED_PORT']}/embedding"
DB_FILE = CFG["DB_FILE"]
CHUNK_SIZE = CFG["CHUNK_SIZE"]

def load_db():
    if os.path.exists(DB_FILE):
        try:
            with open(DB_FILE, "r") as f:
                data = json.load(f)
            dirty = False
            for k, doc in data.items():
                for ch in doc.get("chunks", []):
                    v = ch.get("vector", [])
                    while isinstance(v, list) and len(v) > 0 and isinstance(v[0], list):
                        v = v[0]
                        ch["vector"] = v
                        dirty = True
            if dirty:
                save_db(data)
            return data
        except Exception:
            try:
                corrupt_bak = f"{DB_FILE}.corrupted.{int(time.time())}"
                shutil.copy2(DB_FILE, corrupt_bak)
                print(f"[Warn] Corrupted database backed up to {corrupt_bak}")
            except Exception:
                pass
            return {}
    return {}

def save_db(data):
    tmp_file = f"{DB_FILE}.tmp"
    with open(tmp_file, "w") as f:
        json.dump(data, f, indent=2)
        f.flush()
        try:
            os.fsync(f.fileno())
        except Exception:
            pass
    os.replace(tmp_file, DB_FILE)

def extract_embedding_vector(data):
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

def http_post_json(url, payload, timeout=20):
    if requests is not None:
        try:
            return requests.post(url, json=payload, timeout=timeout)
        except Exception:
            return None
    try:
        data = json.dumps(payload).encode('utf-8')
        req = urllib.request.Request(url, data=data, headers={'Content-Type': 'application/json'}, method='POST')
        with urllib.request.urlopen(req, timeout=timeout) as resp:
            class MockResponse:
                def __init__(self, code, body):
                    self.status_code = code
                    self.body = body
                def json(self):
                    return json.loads(self.body.decode('utf-8'))
            return MockResponse(resp.status, resp.read())
    except Exception:
        return None

def get_embedding(text):
    text_clean = text.strip()
    if not text_clean:
        return []
    clean_text = "".join(ch for ch in text_clean if ch in ('\n', '\t', '\r') or ord(ch) >= 32)
    if len(clean_text) > 1500:
        clean_text = clean_text[:1500]

    port = CFG.get("EMBED_PORT", "8081")
    candidates = [
        (f"http://127.0.0.1:{port}/v1/embeddings", {"input": clean_text}),
        (f"http://127.0.0.1:{port}/embedding", {"content": clean_text}),
        (f"http://127.0.0.1:{port}/embedding", {"input": clean_text}),
    ]
    for attempt in range(3):
        for url, payload in candidates:
            try:
                res = http_post_json(url, payload, timeout=15)
                if res and res.status_code == 200:
                    vec = extract_embedding_vector(res.json())
                    if vec and len(vec) > 0 and not any(isinstance(x, (int, float)) and (x != x or abs(x) == float('inf')) for x in vec):
                        return vec
            except Exception:
                continue
        if attempt < 2:
            time.sleep(0.3 * (2 ** attempt))
    return []

# =========================================================================
# MULTI-MODAL & MULTI-FORMAT EXTRACTION ENGINE
# =========================================================================

AUDIO_EXTS = {".mp3", ".wav", ".m4a", ".ogg", ".flac", ".aac", ".opus", ".wma", ".webm", ".mid", ".midi"}
VIDEO_EXTS = {".mp4", ".mkv", ".avi", ".mov", ".webm", ".flv", ".wmv", ".m4v", ".3gp", ".ts", ".ogv"}
IMAGE_EXTS = {".jpg", ".jpeg", ".png", ".webp", ".gif", ".bmp", ".tiff", ".tif", ".heic", ".svg", ".ico"}
DOC_EXTS   = {".txt", ".md", ".markdown", ".pdf", ".docx", ".pptx", ".xlsx", ".epub", ".odt", ".ods", ".odp", ".rtf", ".csv", ".tsv", ".html", ".htm", ".tex", ".rst", ".org", ".log", ".adoc"}
CODE_EXTS  = {
    ".py", ".js", ".jsx", ".ts", ".tsx", ".css", ".scss", ".sass", ".json", ".jsonl", ".yaml", ".yml",
    ".toml", ".xml", ".sql", ".sh", ".bash", ".zsh", ".c", ".cpp", ".cc", ".cxx", ".h", ".hpp",
    ".rs", ".go", ".java", ".kt", ".kts", ".php", ".rb", ".lua", ".dart", ".swift", ".r", ".m",
    ".scala", ".pl", ".asm", ".dockerfile", ".ini", ".conf", ".cfg", ".env", ".diff", ".patch"
}

def categorize_file(filepath):
    chats_dir = CFG.get("CHATS_DIR", "")
    base = os.path.basename(filepath).lower()
    ext = os.path.splitext(filepath)[1].lower()
    if (chats_dir and os.path.abspath(filepath).startswith(os.path.abspath(chats_dir))) or base.startswith("chat_"):
        return "conversation", ext
    if base in ["dockerfile", "makefile", "license", "readme"]:
        return "code", ext
    if ext in AUDIO_EXTS:
        return "audio", ext
    if ext in VIDEO_EXTS:
        return "video", ext
    if ext in IMAGE_EXTS:
        return "image", ext
    if ext in CODE_EXTS:
        return "code", ext
    if ext in DOC_EXTS:
        return "document", ext
    return "other", ext

def find_companion_text(filepath):
    """Find sidecar transcript, caption, or notes file matching basename."""
    base_no_ext, _ = os.path.splitext(filepath)
    parent_dir = os.path.dirname(filepath)
    candidates = [
        f"{base_no_ext}.txt",
        f"{base_no_ext}.caption",
        f"{base_no_ext}.srt",
        f"{base_no_ext}.vtt",
        f"{base_no_ext}.lrc",
        f"{base_no_ext}.sub",
        f"{base_no_ext}.json"
    ]
    for c in candidates:
        if os.path.exists(c) and c != filepath:
            try:
                if c.endswith(".json"):
                    with open(c, "r", errors="ignore") as f:
                        data = json.load(f)
                    if isinstance(data, dict):
                        return " ".join(f"{k}: {v}" for k, v in data.items() if isinstance(v, (str, int, float)))
                with open(c, "r", errors="ignore") as f:
                    content = f.read().strip()
                if content:
                    return content
            except Exception:
                pass
    return ""

# ----------------- Audio Extraction -----------------
def extract_audio_content(filepath):
    fname = os.path.basename(filepath)
    ext = os.path.splitext(filepath)[1].lower()
    meta = {}

    # Tool extraction: ffprobe if present
    if shutil.which("ffprobe"):
        try:
            cmd = ["ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", filepath]
            out = subprocess.run(cmd, capture_output=True, text=True, timeout=10)
            if out.returncode == 0 and out.stdout.strip():
                pdata = json.loads(out.stdout)
                fmt = pdata.get("format", {})
                dur = float(fmt.get("duration", 0))
                if dur > 0:
                    meta["Duration"] = f"{int(dur // 60)}m {int(dur % 60)}s"
                tags = fmt.get("tags", {})
                for k in ["title", "artist", "album", "genre", "date", "comment", "lyrics"]:
                    for tk, tv in tags.items():
                        if tk.lower() == k and tv:
                            meta[k.capitalize()] = tv
        except Exception:
            pass

    # Pure Python ID3v1 / ID3v2 extraction
    if not meta and ext == ".mp3" and os.path.exists(filepath):
        try:
            with open(filepath, "rb") as f:
                header = f.read(10)
                # ID3v2
                if header[:3] == b'ID3':
                    tag_size = ((header[6] & 0x7F) << 21) | ((header[7] & 0x7F) << 14) | ((header[8] & 0x7F) << 7) | (header[9] & 0x7F)
                    frames_data = f.read(min(tag_size, 65536))
                    frame_map = {b'TIT2': 'Title', b'TPE1': 'Artist', b'TALB': 'Album', b'TCON': 'Genre', b'COMM': 'Comment', b'USLT': 'Lyrics'}
                    pos = 0
                    while pos + 10 < len(frames_data):
                        fid = frames_data[pos:pos+4]
                        if fid[0] == 0:
                            break
                        fsize = struct.unpack(">I", frames_data[pos+4:pos+8])[0]
                        pos += 10
                        if pos + fsize <= len(frames_data) and fid in frame_map:
                            raw = frames_data[pos:pos+fsize]
                            if len(raw) > 1:
                                enc = raw[0]
                                body = raw[1:]
                                try:
                                    if enc == 0: val = body.decode('latin-1', errors='ignore')
                                    elif enc == 1: val = body.decode('utf-16', errors='ignore')
                                    elif enc == 3: val = body.decode('utf-8', errors='ignore')
                                    else: val = body.decode('utf-8', errors='ignore')
                                    meta[frame_map[fid]] = val.strip('\x00 ')
                                except Exception:
                                    pass
                        pos += fsize
                # ID3v1 fallback
                f.seek(max(0, os.path.getsize(filepath) - 128))
                v1 = f.read(128)
                if v1[:3] == b'TAG':
                    if "Title" not in meta: meta["Title"] = v1[3:33].decode('latin-1', errors='ignore').strip('\x00 ')
                    if "Artist" not in meta: meta["Artist"] = v1[33:63].decode('latin-1', errors='ignore').strip('\x00 ')
                    if "Album" not in meta: meta["Album"] = v1[63:93].decode('latin-1', errors='ignore').strip('\x00 ')
        except Exception:
            pass

    # Pure Python WAV metadata
    if ext == ".wav" and os.path.exists(filepath):
        try:
            with open(filepath, "rb") as f:
                hdr = f.read(44)
                if hdr[:4] == b'RIFF' and hdr[8:12] == b'WAVE':
                    channels = struct.unpack('<H', hdr[22:24])[0]
                    s_rate = struct.unpack('<I', hdr[24:28])[0]
                    b_rate = struct.unpack('<I', hdr[28:32])[0]
                    if b_rate > 0:
                        dur = os.path.getsize(filepath) / b_rate
                        meta["Duration"] = f"{int(dur // 60)}m {int(dur % 60)}s"
                    meta["Format"] = f"WAV Audio ({s_rate} Hz, {channels} ch)"
        except Exception:
            pass

    companion = find_companion_text(filepath)
    parts = [f"[Audio File: {fname}]", f"Type: {ext.upper().strip('.')} Audio"]
    for k, v in meta.items():
        if v: parts.append(f"{k}: {v}")
    if companion:
        parts.append(f"Transcript / Associated Text:\n{companion}")
    return "\n".join(parts)

# ----------------- Video Extraction -----------------
def extract_video_content(filepath):
    fname = os.path.basename(filepath)
    ext = os.path.splitext(filepath)[1].lower()
    meta = {}

    if shutil.which("ffprobe"):
        try:
            cmd = ["ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", "-show_streams", filepath]
            out = subprocess.run(cmd, capture_output=True, text=True, timeout=12)
            if out.returncode == 0 and out.stdout.strip():
                pdata = json.loads(out.stdout)
                fmt = pdata.get("format", {})
                dur = float(fmt.get("duration", 0))
                if dur > 0:
                    meta["Duration"] = f"{int(dur // 60)}m {int(dur % 60)}s"
                tags = fmt.get("tags", {})
                for k in ["title", "description", "artist", "date"]:
                    for tk, tv in tags.items():
                        if tk.lower() == k and tv:
                            meta[k.capitalize()] = tv
                for s in pdata.get("streams", []):
                    if s.get("codec_type") == "video":
                        w = s.get("width")
                        h = s.get("height")
                        if w and h:
                            meta["Resolution"] = f"{w}x{h}"
                        meta["Video Codec"] = s.get("codec_name", "")
                        break
        except Exception:
            pass

    companion = find_companion_text(filepath)
    parts = [f"[Video File: {fname}]", f"Format: {ext.upper().strip('.')} Video"]
    for k, v in meta.items():
        if v: parts.append(f"{k}: {v}")
    if companion:
        parts.append(f"Subtitles / Speech Transcript:\n{companion}")
    return "\n".join(parts)

# ----------------- Image / Photo Extraction -----------------
def extract_image_content(filepath):
    fname = os.path.basename(filepath)
    ext = os.path.splitext(filepath)[1].lower()
    meta = {}

    # SVG parser
    if ext == ".svg":
        try:
            with open(filepath, "r", errors="ignore") as f:
                raw_svg = f.read()
            text_nodes = re.findall(r'<text[^>]*>(.*?)</text>', raw_svg, re.DOTALL)
            title = re.search(r'<title>(.*?)</title>', raw_svg)
            desc = re.search(r'<desc>(.*?)</desc>', raw_svg)
            body = []
            if title: body.append(f"Title: {title.group(1).strip()}")
            if desc: body.append(f"Description: {desc.group(1).strip()}")
            if text_nodes:
                clean_texts = [re.sub(r'<[^>]+>', '', t).strip() for t in text_nodes if t.strip()]
                body.append("Text in diagram:\n" + "\n".join(clean_texts))
            return f"[SVG Vector Graphic: {fname}]\n" + "\n".join(body)
        except Exception:
            pass

    # Pure Python JPEG EXIF reader
    if ext in [".jpg", ".jpeg"] and os.path.exists(filepath):
        try:
            with open(filepath, "rb") as f:
                data = f.read(min(os.path.getsize(filepath), 131072))
            if len(data) > 4 and data[:2] == b'\xff\xd8':
                pos = 2
                while pos < len(data) - 4:
                    if data[pos] != 0xff: break
                    marker = data[pos+1]
                    pos += 2
                    if marker in (0xd9, 0xda): break
                    if pos + 2 > len(data): break
                    seg_len = struct.unpack('>H', data[pos:pos+2])[0]
                    seg = data[pos+2:pos+seg_len]
                    if marker == 0xe1 and seg.startswith(b'Exif\x00\x00'):
                        exif = seg[6:]
                        if len(exif) > 8:
                            endian = '<' if exif[:2] == b'II' else '>'
                            ifd_offset = struct.unpack(endian + 'I', exif[4:8])[0]
                            if ifd_offset < len(exif) - 2:
                                num_entries = struct.unpack(endian + 'H', exif[ifd_offset:ifd_offset+2])[0]
                                curr = ifd_offset + 2
                                tag_names = {
                                    0x010F: "Camera Make", 0x0110: "Camera Model", 0x0132: "DateTime",
                                    0x9003: "Date Taken", 0x010E: "Image Description", 0x9286: "User Comment",
                                    0x0131: "Software"
                                }
                                for _ in range(min(num_entries, 60)):
                                    if curr + 12 > len(exif): break
                                    tag, dtype, count = struct.unpack(endian + 'HHI', exif[curr:curr+8])
                                    val_bytes = exif[curr+8:curr+12]
                                    if tag in tag_names and dtype == 2: # ASCII
                                        if count <= 4:
                                            meta[tag_names[tag]] = val_bytes[:count].decode('latin-1', errors='ignore').strip('\x00 ')
                                        else:
                                            off = struct.unpack(endian + 'I', val_bytes)[0]
                                            if off + count <= len(exif):
                                                meta[tag_names[tag]] = exif[off:off+count].decode('latin-1', errors='ignore').strip('\x00 ')
                                    curr += 12
                    elif marker in (0xc0, 0xc2): # SOF0 / SOF2 dimensions
                        if len(seg) >= 5:
                            h, w = struct.unpack('>HH', seg[1:5])
                            meta["Dimensions"] = f"{w}x{h}"
                    pos += seg_len
        except Exception:
            pass

    # Pure Python PNG text chunks
    if ext == ".png" and os.path.exists(filepath):
        try:
            with open(filepath, "rb") as f:
                sig = f.read(8)
                if sig == b'\x89PNG\r\n\x1a\n':
                    while True:
                        c_hdr = f.read(8)
                        if len(c_hdr) < 8: break
                        c_len, c_type = struct.unpack('>I4s', c_hdr)
                        c_data = f.read(c_len)
                        f.read(4) # CRC
                        if c_type == b'IHDR' and len(c_data) >= 8:
                            w, h = struct.unpack('>II', c_data[:8])
                            meta["Dimensions"] = f"{w}x{h}"
                        elif c_type == b'tEXt':
                            parts = c_data.split(b'\x00', 1)
                            if len(parts) == 2:
                                meta[parts[0].decode('latin-1', errors='ignore')] = parts[1].decode('latin-1', errors='ignore')
                        elif c_type == b'IEND':
                            break
        except Exception:
            pass

    ocr_text = ""
    if shutil.which("tesseract"):
        try:
            cmd = ["tesseract", filepath, "stdout", "--oem", "1", "-l", "eng"]
            proc = subprocess.run(cmd, capture_output=True, text=True, timeout=15)
            if proc.returncode == 0:
                ocr_text = proc.stdout.strip()
        except Exception:
            pass

    companion = find_companion_text(filepath)
    parts = [f"[Photo / Image: {fname}]", f"Type: {ext.upper().strip('.')} Image"]
    for k, v in meta.items():
        if v: parts.append(f"{k}: {v}")
    if companion:
        parts.append(f"Caption / Description:\n{companion}")
    if ocr_text:
        parts.append(f"Visual Text Content (OCR):\n{ocr_text}")
    return "\n".join(parts)

# ----------------- Office & Document Extraction -----------------
def extract_office_content(filepath):
    ext = os.path.splitext(filepath)[1].lower()
    fname = os.path.basename(filepath)

    # Word .docx
    if ext == ".docx":
        try:
            with zipfile.ZipFile(filepath) as z:
                xml_data = z.read("word/document.xml").decode("utf-8", errors="ignore")
                text = re.sub(r'<w:p[ >]', '\n', xml_data)
                text = re.sub(r'<[^>]+>', '', text)
                clean_text = "\n".join(line.strip() for line in text.splitlines() if line.strip())
                return f"[Document: {fname} (Word .docx)]\n\n{clean_text}"
        except Exception:
            pass

    # PowerPoint .pptx
    if ext == ".pptx":
        try:
            with zipfile.ZipFile(filepath) as z:
                slides_text = []
                slide_files = sorted([n for n in z.namelist() if n.startswith("ppt/slides/slide") and n.endswith(".xml")])
                for idx, sfile in enumerate(slide_files):
                    s_data = z.read(sfile).decode("utf-8", errors="ignore")
                    s_text = re.sub(r'<[^>]+>', ' ', s_data)
                    clean = " ".join(s_text.split())
                    if clean:
                        slides_text.append(f"--- Slide {idx+1} ---\n{clean}")
                return f"[Presentation: {fname} (PowerPoint .pptx)]\n\n" + "\n\n".join(slides_text)
        except Exception:
            pass

    # Excel .xlsx
    if ext == ".xlsx":
        try:
            with zipfile.ZipFile(filepath) as z:
                shared_text = []
                if "xl/sharedStrings.xml" in z.namelist():
                    s_data = z.read("xl/sharedStrings.xml").decode("utf-8", errors="ignore")
                    strings = re.findall(r'<t[^>]*>(.*?)</t>', s_data)
                    shared_text = [s.strip() for s in strings if s.strip()]
                return f"[Spreadsheet: {fname} (Excel .xlsx)]\n\nData Cells:\n" + ", ".join(shared_text[:500])
        except Exception:
            pass

    # OpenDocument (.odt, .ods, .odp)
    if ext in [".odt", ".ods", ".odp"]:
        try:
            with zipfile.ZipFile(filepath) as z:
                if "content.xml" in z.namelist():
                    c_data = z.read("content.xml").decode("utf-8", errors="ignore")
                    text = re.sub(r'<[^>]+>', ' ', c_data)
                    clean = "\n".join(line.strip() for line in text.splitlines() if line.strip())
                    return f"[Document: {fname} ({ext})]\n\n{clean}"
        except Exception:
            pass

    # EPUB E-book
    if ext == ".epub":
        try:
            with zipfile.ZipFile(filepath) as z:
                chapters = []
                for n in sorted(z.namelist()):
                    if n.endswith((".xhtml", ".html", ".htm")) and not n.endswith("toc.ncx"):
                        h_data = z.read(n).decode("utf-8", errors="ignore")
                        text = re.sub(r'<[^>]+>', ' ', h_data)
                        clean = "\n".join(l.strip() for l in text.splitlines() if l.strip())
                        if clean:
                            chapters.append(clean)
                return f"[E-Book: {fname} (.epub)]\n\n" + "\n\n".join(chapters[:20])
        except Exception:
            pass

    # Rich Text (.rtf)
    if ext == ".rtf":
        try:
            with open(filepath, "rb") as f:
                raw_bytes = f.read()
            text = raw_bytes.decode('latin-1', errors='ignore')
            text = re.sub(r'\\par[d]?', '\n', text)
            text = re.sub(r'\\tab', '\t', text)
            text = re.sub(r'\\u(-?\d+)\??', lambda m: chr(int(m.group(1)) % 65536), text)
            text = re.sub(r'\\\'([0-9a-fA-F]{2})', lambda m: chr(int(m.group(1), 16)), text)
            text = re.sub(r'\{\*?\\[^{}]+?\}', '', text)
            text = re.sub(r'\\[a-zA-Z]+-?\d* ?', '', text)
            clean = text.replace('{', '').replace('}', '').strip()
            return f"[Document: {fname} (.rtf)]\n\n{clean}"
        except Exception:
            pass

    # CSV / TSV
    if ext in [".csv", ".tsv"]:
        try:
            with open(filepath, "r", errors="ignore") as f:
                lines = f.readlines()
            head = "".join(lines[:100])
            return f"[Data Table: {fname} ({ext})]\n\n{head}"
        except Exception:
            pass

    # HTML / HTM
    if ext in [".html", ".htm"]:
        try:
            with open(filepath, "r", errors="ignore") as f:
                html_raw = f.read()
            title = re.search(r'<title>(.*?)</title>', html_raw, re.I)
            text = re.sub(r'<script.*?</script>', '', html_raw, flags=re.DOTALL | re.I)
            text = re.sub(r'<style.*?</style>', '', text, flags=re.DOTALL | re.I)
            text = re.sub(r'<[^>]+>', ' ', text)
            clean = "\n".join(l.strip() for l in text.splitlines() if l.strip())
            header = f"[Web Document: {fname}]"
            if title: header += f"\nTitle: {title.group(1).strip()}"
            return f"{header}\n\n{clean}"
        except Exception:
            pass

    # PDF Document
    if ext == ".pdf":
        if shutil.which("pdftotext"):
            res = subprocess.run(["pdftotext", filepath, "-"], capture_output=True, text=True)
            if res.stdout.strip():
                return res.stdout.strip()
        # Pure Python PDF stream fallback
        try:
            with open(filepath, "rb") as f:
                pdf_data = f.read()
            text_chunks = []
            for match in re.finditer(b'stream[\r\n]+(.*?)[\r\n]+endstream', pdf_data, re.DOTALL):
                stream_data = match.group(1)
                try:
                    decompressed = zlib.decompress(stream_data)
                except Exception:
                    try:
                        decompressed = zlib.decompress(stream_data, -15)
                    except Exception:
                        decompressed = stream_data
                if decompressed:
                    for s in re.findall(rb'\(([^\(\)\\]*(?:\\.[^\(\)\\]*)*)\)\s*T[jJ]', decompressed):
                        s_clean = re.sub(rb'\\([0-9]{3})', lambda m: bytes([int(m.group(1), 8)]), s)
                        s_clean = s_clean.replace(b'\\n', b'\n').replace(b'\\r', b'\r').replace(b'\\t', b'\t').replace(b'\\(', b'(').replace(b'\\)', b')').replace(b'\\\\', b'\\')
                        decoded = s_clean.decode('utf-8', errors='ignore').strip()
                        if decoded:
                            text_chunks.append(decoded)
            if text_chunks:
                return f"[PDF Document: {fname}]\n\n" + " ".join(text_chunks)
        except Exception:
            pass

    return ""

# ----------------- Universal Text & Fallback -----------------
def extract_universal_content(filepath):
    fname = os.path.basename(filepath)
    ext = os.path.splitext(filepath)[1].lower()

    # Plain text / code
    try:
        with open(filepath, "r", encoding="utf-8", errors="ignore") as f:
            content = f.read()
        # Verify printable content ratio
        printable = sum(1 for c in content[:4096] if c.isprintable() or c in '\n\r\t')
        if len(content) == 0 or printable / max(1, len(content[:4096])) >= 0.80:
            if ext in CODE_EXTS:
                return f"[Code File: {fname} ({ext})]\n\n{content}"
            return content.strip()
    except Exception:
        pass

    # Binary string extractor fallback (range not limited to anything)
    try:
        size = os.path.getsize(filepath)
        if size > 25 * 1024 * 1024:
            return ""
        with open(filepath, "rb") as f:
            raw = f.read(min(size, 262144))
        strings = re.findall(rb'[\x20-\x7e]{4,}', raw)
        clean_strings = [s.decode('latin-1', errors='ignore') for s in strings[:150]]
        if clean_strings:
            return f"[Binary / Data File: {fname} ({ext} | Size: {size} bytes)]\nExtracted Text Sequences:\n" + "\n".join(clean_strings)
    except Exception:
        pass

    return ""

def extract_text(filepath):
    """Main extraction dispatcher for all file types."""
    category, _ = categorize_file(filepath)

    if category == "conversation":
        try:
            with open(filepath, "r", encoding="utf-8", errors="replace") as f:
                return f.read()
        except Exception:
            return ""
    elif category == "audio":
        return extract_audio_content(filepath)
    elif category == "video":
        return extract_video_content(filepath)
    elif category == "image":
        return extract_image_content(filepath)
    elif category == "document":
        doc_txt = extract_office_content(filepath)
        if doc_txt: return doc_txt
        return extract_universal_content(filepath)
    elif category == "code":
        return extract_universal_content(filepath)
    else:
        return extract_universal_content(filepath)

def chunk_conversation(text, max_chars=CHUNK_SIZE):
    """Chunks conversation sessions by dialogue exchanges (User + Assistant pairs)."""
    pattern = re.compile(r'(?=(?:^|\n)(?:### User|## User|User:|You:))', re.IGNORECASE)
    parts = pattern.split(text)
    chunks = []
    current = ""
    for part in parts:
        part = part.strip()
        if not part:
            continue
        if len(current) + len(part) + 2 <= max_chars:
            current = (current + "\n\n" + part).strip()
        else:
            if current:
                chunks.append(current)
            if len(part) > max_chars:
                for sub in chunk_text(part, max_chars):
                    chunks.append(sub)
                current = ""
            else:
                current = part
    if current:
        chunks.append(current)
    return chunks if chunks else [text]

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

def should_ignore_file(filepath):
    fname = os.path.basename(filepath)
    if fname.startswith((".", "~")) or fname.endswith((".tmp", ".swp", ".bak", ".lock")):
        return True
    if any(part.startswith(".") for part in filepath.split(os.sep)):
        return True
    if any(part in ["node_modules", "__pycache__", ".git"] for part in filepath.split(os.sep)):
        return True
    return False

def index_file(filepath):
    if not os.path.exists(filepath) or os.path.isdir(filepath):
        return
    if should_ignore_file(filepath):
        return

    try:
        content = extract_text(filepath)
        if not content or len(content.strip()) < 10:
            return

        category, _ = categorize_file(filepath)
        db = load_db()
        existing_map = {}
        if filepath in db:
            for old_c in db[filepath].get("chunks", []):
                h = old_c.get("hash") or hashlib.sha256(old_c.get("text", "").encode("utf-8")).hexdigest()
                if old_c.get("vector"):
                    existing_map[h] = old_c["vector"]

        if category == "conversation":
            chunks = chunk_conversation(content)
        else:
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
            category, _ = categorize_file(filepath)
            db[filepath] = {
                "filename": os.path.basename(filepath),
                "file_type": category,
                "file_size": os.path.getsize(filepath),
                "chunks": records,
                "updated_at": time.time()
            }
            save_db(db)
            print(f"[Indexed: {category.upper()}] {os.path.basename(filepath)} ({new_count} new, {reused_count} cached chunks)")
    except Exception as e:
        print(f"[Error] Processing {filepath}: {e}")

def sync_knowledge_base():
    """Scan all watch directories, index files, and prune deleted entries."""
    db = load_db()
    existing_paths = set(db.keys())
    current_paths = set()

    for d in WATCH_DIRS:
        os.makedirs(d, exist_ok=True)
        for root, _, files in os.walk(d):
            for file in files:
                fpath = os.path.join(root, file)
                if not should_ignore_file(fpath):
                    current_paths.add(fpath)
                    index_file(fpath)

    # Prune files that no longer exist
    to_delete = [p for p in existing_paths if not os.path.exists(p)]
    if to_delete:
        for p in to_delete:
            del db[p]
        save_db(db)
        print(f"[Cleaned] Pruned {len(to_delete)} stale file entries from vector database.")

class ChangeHandler(FileSystemEventHandler):
    def on_modified(self, event):
        if not event.is_directory:
            time.sleep(0.5)
            index_file(event.src_path)

    def on_created(self, event):
        if not event.is_directory:
            time.sleep(0.5)
            index_file(event.src_path)

    def on_deleted(self, event):
        if not event.is_directory:
            db = load_db()
            if event.src_path in db:
                del db[event.src_path]
                save_db(db)
                print(f"[Removed] {os.path.basename(event.src_path)} from vector database")

if __name__ == "__main__":
    run_once = any(arg in sys.argv for arg in ["--once", "sync", "-1"])

    sync_knowledge_base()

    if run_once:
        print("Vector database synchronization complete.")
        sys.exit(0)

    if HAS_WATCHDOG and Observer is not None:
        observer = Observer()
        for d in WATCH_DIRS:
            observer.schedule(ChangeHandler(), d, recursive=True)
        observer.start()
        print("Watching knowledge directory for documents, audio, video, images, code & data files (native watchdog)...")
        try:
            while True:
                time.sleep(2)
        except KeyboardInterrupt:
            observer.stop()
    else:
        print("Watching knowledge directory for documents, audio, video, images, code & data files (polling mode)...")
        try:
            while True:
                time.sleep(4)
                sync_knowledge_base()
        except KeyboardInterrupt:
            pass

# -*- coding: utf-8 -*-
"""
Wola Standalone Air-Gapped Web UI
Zero external dependencies (No external CDNs, Google Fonts, or cloud scripts).
Provides full real-time SSE streaming, collapsible reasoning (<think>),
code syntax wrappers with one-click copy, and local mesh device pairing guides.
"""

WEB_UI_HTML = r"""<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no">
<title>Wola — Local AI Mesh Gateway</title>
<style>
:root {
  --bg-main: #090d16;
  --bg-card: #111827;
  --bg-panel: #162032;
  --border: #1f293d;
  --border-focus: #3b82f6;
  --text-main: #f1f5f9;
  --text-muted: #94a3b8;
  --text-dim: #64748b;
  --accent: #10b981;
  --accent-glow: rgba(16, 185, 129, 0.15);
  --blue: #3b82f6;
  --blue-user: #2563eb;
  --think-bg: #0c1424;
  --think-border: #1e3a5f;
  --code-bg: #070a12;
}
* { box-sizing: border-box; margin: 0; padding: 0; }
body {
  background: var(--bg-main);
  color: var(--text-main);
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
  height: 100vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}
header {
  background: #0d1322;
  border-bottom: 1px solid var(--border);
  padding: 12px 18px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  flex-shrink: 0;
  z-index: 10;
}
.brand-group {
  display: flex;
  align-items: center;
  gap: 10px;
}
.brand-icon {
  width: 32px;
  height: 32px;
  background: linear-gradient(135deg, #10b981, #06b6d4);
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 900;
  font-size: 18px;
  color: #030712;
  box-shadow: 0 0 12px rgba(16, 185, 129, 0.3);
}
.brand-title {
  font-weight: 700;
  font-size: 15px;
  letter-spacing: 0.5px;
  display: flex;
  align-items: center;
  gap: 8px;
}
.badge-mesh {
  font-size: 11px;
  font-weight: 600;
  padding: 3px 8px;
  border-radius: 999px;
  background: var(--accent-glow);
  color: var(--accent);
  border: 1px solid rgba(16, 185, 129, 0.3);
  display: inline-flex;
  align-items: center;
  gap: 5px;
}
.badge-mesh::before {
  content: "";
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--accent);
  display: inline-block;
  box-shadow: 0 0 6px var(--accent);
}
.header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}
.btn-action {
  background: var(--bg-card);
  border: 1px solid var(--border);
  color: var(--text-main);
  padding: 6px 12px;
  border-radius: 7px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  transition: all 0.15s ease;
}
.btn-action:hover {
  background: var(--bg-panel);
  border-color: var(--border-focus);
}
.btn-pair {
  background: rgba(59, 130, 246, 0.12);
  border-color: rgba(59, 130, 246, 0.4);
  color: #60a5fa;
}
.btn-pair:hover {
  background: rgba(59, 130, 246, 0.22);
  border-color: #60a5fa;
}
.chat-container {
  flex: 1;
  overflow-y: auto;
  padding: 20px 16px;
  display: flex;
  flex-direction: column;
  gap: 18px;
  max-width: 860px;
  width: 100%;
  margin: 0 auto;
}
.welcome-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 12px;
  padding: 24px;
  text-align: center;
  margin: auto 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 14px;
}
.welcome-icon {
  font-size: 36px;
}
.welcome-title {
  font-size: 20px;
  font-weight: 700;
  color: #f8fafc;
}
.welcome-sub {
  font-size: 14px;
  color: var(--text-muted);
  max-width: 520px;
  line-height: 1.5;
}
.starter-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 10px;
  width: 100%;
  max-width: 650px;
  margin-top: 10px;
}
.starter-pill {
  background: var(--bg-panel);
  border: 1px solid var(--border);
  padding: 10px 14px;
  border-radius: 8px;
  font-size: 13px;
  color: var(--text-main);
  cursor: pointer;
  text-align: left;
  transition: all 0.15s ease;
}
.starter-pill:hover {
  border-color: var(--blue);
  background: #1c2b42;
  transform: translateY(-1px);
}
.msg-row {
  display: flex;
  width: 100%;
  gap: 12px;
}
.msg-row.user {
  justify-content: flex-end;
}
.msg-row.assistant {
  justify-content: flex-start;
}
.msg-avatar {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: #1e293b;
  border: 1px solid var(--border);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  flex-shrink: 0;
}
.msg-bubble {
  max-width: 82%;
  padding: 12px 16px;
  border-radius: 12px;
  font-size: 14.5px;
  line-height: 1.6;
  word-break: break-word;
}
.msg-row.user .msg-bubble {
  background: var(--blue-user);
  color: #ffffff;
  border-bottom-right-radius: 3px;
}
.msg-row.assistant .msg-bubble {
  background: var(--bg-panel);
  border: 1px solid var(--border);
  color: var(--text-main);
  border-bottom-left-radius: 3px;
}
.thought-box {
  background: var(--think-bg);
  border: 1px solid var(--think-border);
  border-radius: 8px;
  margin-bottom: 12px;
  overflow: hidden;
}
.thought-summary {
  padding: 8px 12px;
  font-size: 12.5px;
  font-weight: 600;
  color: #38bdf8;
  cursor: pointer;
  user-select: none;
  display: flex;
  align-items: center;
  gap: 8px;
  background: rgba(56, 189, 248, 0.05);
}
.thought-body {
  padding: 10px 12px;
  font-size: 13px;
  color: #94a3b8;
  border-top: 1px solid rgba(30, 58, 95, 0.5);
  line-height: 1.5;
  white-space: pre-wrap;
  font-family: monospace;
}
.pulse-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #38bdf8;
  display: inline-block;
  animation: pulse 1.2s infinite ease-in-out;
}
@keyframes pulse {
  0%, 100% { opacity: 0.3; transform: scale(0.85); }
  50% { opacity: 1; transform: scale(1.15); }
}
.code-wrapper {
  background: var(--code-bg);
  border: 1px solid var(--border);
  border-radius: 8px;
  margin: 10px 0;
  overflow: hidden;
}
.code-header {
  background: #0f172a;
  padding: 6px 12px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid var(--border);
  font-size: 12px;
  color: var(--text-dim);
}
.code-copy {
  background: transparent;
  border: 1px solid var(--border);
  color: var(--text-muted);
  border-radius: 4px;
  padding: 2px 8px;
  font-size: 11px;
  cursor: pointer;
}
.code-copy:hover {
  background: #1e293b;
  color: #fff;
}
pre {
  padding: 12px;
  overflow-x: auto;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 13px;
  line-height: 1.5;
}
code.inline {
  background: #1e293b;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 13px;
  font-family: monospace;
  color: #38bdf8;
}
.input-area {
  background: #0d1322;
  border-top: 1px solid var(--border);
  padding: 12px 18px;
  flex-shrink: 0;
}
.input-box-wrapper {
  max-width: 860px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.input-row {
  display: flex;
  align-items: flex-end;
  gap: 8px;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 10px;
  padding: 8px 10px;
  transition: border-color 0.15s ease;
}
.input-row:focus-within {
  border-color: var(--border-focus);
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.2);
}
textarea {
  flex: 1;
  background: transparent;
  border: none;
  color: var(--text-main);
  font-size: 14.5px;
  resize: none;
  outline: none;
  max-height: 160px;
  min-height: 24px;
  line-height: 1.45;
  font-family: inherit;
}
.btn-send {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  background: var(--blue-user);
  border: none;
  color: #ffffff;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  flex-shrink: 0;
  transition: background 0.15s ease;
}
.btn-send:hover {
  background: #1d4ed8;
}
.btn-send.stop {
  background: #ef4444;
}
.btn-send.stop:hover {
  background: #dc2626;
}
.footer-text {
  font-size: 11.5px;
  color: var(--text-dim);
  text-align: center;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
}
.modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(3, 7, 18, 0.75);
  backdrop-filter: blur(4px);
  display: none;
  align-items: center;
  justify-content: center;
  z-index: 100;
  padding: 16px;
}
.modal-backdrop.active {
  display: flex;
}
.modal-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 14px;
  max-width: 580px;
  width: 100%;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.5);
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.modal-title {
  font-size: 17px;
  font-weight: 700;
  color: #f8fafc;
}
.modal-close {
  background: transparent;
  border: none;
  color: var(--text-muted);
  font-size: 18px;
  cursor: pointer;
}
.conn-card {
  background: var(--bg-panel);
  border: 1px solid var(--border);
  border-radius: 10px;
  padding: 14px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.conn-title {
  font-size: 13.5px;
  font-weight: 600;
  color: #38bdf8;
  display: flex;
  align-items: center;
  gap: 6px;
}
.conn-desc {
  font-size: 12.5px;
  color: var(--text-muted);
  line-height: 1.4;
}
.conn-link-row {
  display: flex;
  align-items: center;
  gap: 8px;
  background: #090e18;
  border: 1px solid var(--border);
  border-radius: 6px;
  padding: 6px 10px;
}
.conn-link-text {
  font-family: monospace;
  font-size: 12.5px;
  color: #34d399;
  flex: 1;
  word-break: break-all;
}
.conn-copy-btn {
  background: #1e293b;
  border: 1px solid var(--border);
  color: var(--text-main);
  padding: 3px 8px;
  border-radius: 4px;
  font-size: 11px;
  cursor: pointer;
  white-space: nowrap;
}
.conn-copy-btn:hover {
  background: #334155;
}
@media (max-width: 640px) {
  .msg-bubble { max-width: 90%; }
  .starter-grid { grid-template-columns: 1fr; }
}
</style>
</head>
<body>
<header>
  <div class="brand-group">
    <div class="brand-icon">⚡</div>
    <div>
      <div class="brand-title">WOLA <span class="badge-mesh">LOCAL MESH</span></div>
    </div>
  </div>
  <div class="header-actions">
    <button class="btn-action btn-pair" id="btn-pair">📱 Pair Devices</button>
    <button class="btn-action" id="btn-clear">🗑️ Clear</button>
  </div>
</header>

<main class="chat-container" id="chat-stream">
  <div class="welcome-card" id="welcome-card">
    <div class="welcome-icon">🐺</div>
    <div class="welcome-title">Wola Local Autonomous Gateway</div>
    <div class="welcome-sub">
      Chat directly with your on-device AI node, powered by autonomous reasoning and personalized with your private knowledge (documents, audio, video, photos, and code). Works 100% offline.
    </div>
    <div class="starter-grid">
      <div class="starter-pill" onclick="sendPrompt('What information is indexed in my local knowledge base?')">📚 Explore personal knowledge</div>
      <div class="starter-pill" onclick="sendPrompt('Explain how local RAG vector search works.')">🧠 Explain local RAG</div>
      <div class="starter-pill" onclick="sendPrompt('Write a clean Python script to check system telemetry.')">⚡ Python telemetry script</div>
      <div class="starter-pill" onclick="sendPrompt('How can I connect my laptop to this phone via Hotspot?')">📱 Hotspot pairing steps</div>
    </div>
  </div>
</main>

<div class="input-area">
  <div class="input-box-wrapper">
    <div class="input-row">
      <textarea id="prompt-input" rows="1" placeholder="Ask anything, discuss your documents, media, code (Shift+Enter for newline)..."></textarea>
      <button class="btn-send" id="btn-send" title="Send message">
        <svg id="icon-send" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 2L11 13M22 2l-7 20-4-9-9-4 20-7z"/></svg>
        <svg id="icon-stop" width="16" height="16" viewBox="0 0 24 24" fill="currentColor" style="display:none;"><rect x="4" y="4" width="16" height="16" rx="2"/></svg>
      </button>
    </div>
    <div class="footer-text">
      <span>🔒 100% Air-gapped Local Mesh</span>
      <span>•</span>
      <span>Direct P2P Link</span>
      <span>•</span>
      <span id="rag-badge">RAG Ready</span>
    </div>
  </div>
</div>

<div class="modal-backdrop" id="pair-modal">
  <div class="modal-card">
    <div class="modal-header">
      <div class="modal-title">📱 Connect Other Devices to Wola</div>
      <button class="modal-close" id="modal-close">✕</button>
    </div>
    <div class="conn-card">
      <div class="conn-title">🔥 Android Mobile Hotspot (No Internet or Wi-Fi Router)</div>
      <div class="conn-desc">
        1. On this Android device: Turn <b>ON</b> Mobile Hotspot / Personal Hotspot.<br>
        2. On your laptop, tablet, or secondary phone: Connect to this hotspot Wi-Fi.<br>
        3. Open your browser and navigate to:
      </div>
      <div class="conn-link-row">
        <span class="conn-link-text" id="link-hotspot">http://192.168.43.1:5000</span>
        <button class="conn-copy-btn" onclick="copyValue(document.getElementById('link-hotspot').innerText, this)">Copy</button>
      </div>
    </div>
    <div class="conn-card">
      <div class="conn-title">📶 Local Wi-Fi Network / LAN</div>
      <div class="conn-desc">
        If both devices are already on the same local Wi-Fi router:
      </div>
      <div class="conn-link-row">
        <span class="conn-link-text" id="link-lan">Detecting...</span>
        <button class="conn-copy-btn" id="btn-copy-lan" onclick="copyValue(document.getElementById('link-lan').innerText, this)">Copy</button>
      </div>
    </div>
    <div class="conn-card">
      <div class="conn-title">🔌 USB Tethering</div>
      <div class="conn-desc">
        Connect phone to laptop via USB cable and enable USB Tethering in Android settings:
      </div>
      <div class="conn-link-row">
        <span class="conn-link-text" id="link-usb">Detecting...</span>
        <button class="conn-copy-btn" onclick="copyValue(document.getElementById('link-usb').innerText, this)">Copy</button>
      </div>
    </div>
    <div class="conn-card">
      <div class="conn-title">⚡ OpenAI-Compatible API Endpoint</div>
      <div class="conn-desc">
        Connect Open-WebUI, Chatbox, Cursor, Continue, or curl:
      </div>
      <div class="conn-link-row">
        <span class="conn-link-text" id="link-api">http://192.168.43.1:5000/v1</span>
        <button class="conn-copy-btn" onclick="copyValue(document.getElementById('link-api').innerText, this)">Copy</button>
      </div>
      <div class="conn-desc" style="font-size:11.5px; color:#64748b; margin-top:2px;">
        API Key: <code>sk-test</code> &nbsp;|&nbsp; Model ID: <code>local-agent</code>
      </div>
    </div>
  </div>
</div>

<script>
let messages = [];
let activeAbort = null;

const chatStream = document.getElementById("chat-stream");
const welcomeCard = document.getElementById("welcome-card");
const promptInput = document.getElementById("prompt-input");
const btnSend = document.getElementById("btn-send");
const iconSend = document.getElementById("icon-send");
const iconStop = document.getElementById("icon-stop");
const pairModal = document.getElementById("pair-modal");
const btnPair = document.getElementById("btn-pair");
const btnClear = document.getElementById("btn-clear");
const modalClose = document.getElementById("modal-close");

btnPair.addEventListener("click", () => pairModal.classList.add("active"));
modalClose.addEventListener("click", () => pairModal.classList.remove("active"));
pairModal.addEventListener("click", (e) => {
  if (e.target === pairModal) pairModal.classList.remove("active");
});

btnClear.addEventListener("click", () => {
  messages = [];
  chatStream.innerHTML = "";
  chatStream.appendChild(welcomeCard);
  welcomeCard.style.display = "flex";
});

promptInput.addEventListener("input", function() {
  this.style.height = "auto";
  this.style.height = Math.min(this.scrollHeight, 160) + "px";
});

promptInput.addEventListener("keydown", (e) => {
  if (e.key === "Enter" && !e.shiftKey) {
    e.preventDefault();
    if (activeAbort) return;
    handleSubmit();
  }
});

btnSend.addEventListener("click", () => {
  if (activeAbort) {
    activeAbort.abort();
    setStreaming(false);
  } else {
    handleSubmit();
  }
});

function sendPrompt(text) {
  promptInput.value = text;
  handleSubmit();
}

function copyValue(val, btn) {
  navigator.clipboard.writeText(val).then(() => {
    const orig = btn.innerText;
    btn.innerText = "✓ Copied";
    setTimeout(() => { btn.innerText = orig; }, 1800);
  });
}

function escapeHtml(str) {
  return str.replace(/&/g, "&amp;")
            .replace(/</g, "&lt;")
            .replace(/>/g, "&gt;")
            .replace(/"/g, "&quot;")
            .replace(/'/g, "&#039;");
}

function formatMarkdown(text, isStreaming) {
  let thoughtHtml = "";
  let cleanText = text;

  if (cleanText.includes("<think>")) {
    if (cleanText.includes("</think>")) {
      const parts = cleanText.split("</think>");
      const rawThought = parts[0].split("<think>")[1] || "";
      cleanText = parts.slice(1).join("</think>").trim();
      thoughtHtml = `
        <details class="thought-box">
          <summary class="thought-summary">
            <span>💭 Reasoning Process</span>
          </summary>
          <div class="thought-body">${escapeHtml(rawThought.trim())}</div>
        </details>`;
    } else {
      const rawThought = cleanText.split("<think>")[1] || "";
      cleanText = "";
      thoughtHtml = `
        <details class="thought-box" open>
          <summary class="thought-summary">
            <span class="pulse-dot"></span>
            <span>Reasoning in progress...</span>
          </summary>
          <div class="thought-body">${escapeHtml(rawThought.trim())}</div>
        </details>`;
    }
  }

  const codeBlocks = [];
  cleanText = cleanText.replace(/```([a-zA-Z0-9_-]*)\n([\s\S]*?)```/g, (m, lang, code) => {
    const id = codeBlocks.length;
    codeBlocks.push({ lang: lang || "code", code: code.trim() });
    return `%%%CB_${id}%%%`;
  });

  if (isStreaming && cleanText.includes("```")) {
    cleanText = cleanText.replace(/```([a-zA-Z0-9_-]*)\n([\s\S]*)$/g, (m, lang, code) => {
      const id = codeBlocks.length;
      codeBlocks.push({ lang: lang || "code", code: code });
      return `%%%CB_${id}%%%`;
    });
  }

  let formatted = escapeHtml(cleanText);
  formatted = formatted.replace(/`([^`]+)`/g, `<code class="inline">$1</code>`);
  formatted = formatted.replace(/\*\*([^*]+)\*\*/g, `<strong>$1</strong>`);
  formatted = formatted.replace(/\*([^*]+)\*/g, `<em>$1</em>`);
  formatted = formatted.replace(/\n\n+/g, `</p><p style="margin-top:8px;">`);
  formatted = formatted.replace(/\n/g, `<br>`);
  formatted = `<p>${formatted}</p>`;

  formatted = formatted.replace(/%%%CB_(\d+)%%%/g, (m, id) => {
    const item = codeBlocks[parseInt(id, 10)];
    if (!item) return "";
    return `
      <div class="code-wrapper">
        <div class="code-header">
          <span>${escapeHtml(item.lang)}</span>
          <button class="code-copy" onclick="copyValue(this.closest('.code-wrapper').querySelector('code').innerText, this)">Copy code</button>
        </div>
        <pre><code>${escapeHtml(item.code)}</code></pre>
      </div>`;
  });

  return thoughtHtml + formatted;
}

function setStreaming(isStreaming) {
  if (isStreaming) {
    iconSend.style.display = "none";
    iconStop.style.display = "block";
    btnSend.classList.add("stop");
    btnSend.title = "Stop generating";
  } else {
    activeAbort = null;
    iconSend.style.display = "block";
    iconStop.style.display = "none";
    btnSend.classList.remove("stop");
    btnSend.title = "Send message";
  }
}

async function handleSubmit() {
  const query = promptInput.value.trim();
  if (!query) return;

  welcomeCard.style.display = "none";
  promptInput.value = "";
  promptInput.style.height = "auto";

  messages.push({ role: "user", content: query });

  const userRow = document.createElement("div");
  userRow.className = "msg-row user";
  userRow.innerHTML = `<div class="msg-bubble">${escapeHtml(query)}</div>`;
  chatStream.appendChild(userRow);

  const botRow = document.createElement("div");
  botRow.className = "msg-row assistant";
  botRow.innerHTML = `
    <div class="msg-avatar">🐺</div>
    <div class="msg-bubble bot-content"><span class="pulse-dot"></span></div>
  `;
  chatStream.appendChild(botRow);
  chatStream.scrollTop = chatStream.scrollHeight;

  const botContent = botRow.querySelector(".bot-content");
  activeAbort = new AbortController();
  setStreaming(true);

  let accumulated = "";
  let inReasoningTag = false;

  try {
    const res = await fetch("/v1/chat/completions", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ messages, stream: true }),
      signal: activeAbort.signal
    });

    if (!res.ok) {
      throw new Error(`Server returned HTTP ${res.status}`);
    }

    const reader = res.body.getReader();
    const decoder = new TextDecoder();
    let buffer = "";

    while (true) {
      const { done, value } = await reader.read();
      if (done) break;

      buffer += decoder.decode(value, { stream: true });
      const lines = buffer.split("\n");
      buffer = lines.pop();

      for (const line of lines) {
        const trimmed = line.trim();
        if (!trimmed || trimmed === "data: [DONE]") continue;
        if (trimmed.startsWith("data:")) {
          try {
            const data = JSON.parse(trimmed.slice(5).trim());
            const choice = data.choices && data.choices[0];
            const delta = choice && choice.delta;
            if (delta) {
              if (delta.reasoning_content) {
                if (!inReasoningTag && !accumulated.includes("<think>")) {
                  accumulated += "<think>";
                  inReasoningTag = true;
                }
                accumulated += delta.reasoning_content;
              } else if (delta.content) {
                if (inReasoningTag) {
                  accumulated += "</think>\n\n";
                  inReasoningTag = false;
                }
                accumulated += delta.content;
              }
              botContent.innerHTML = formatMarkdown(accumulated, true);
              chatStream.scrollTop = chatStream.scrollHeight;
            }
          } catch (e) {}
        }
      }
    }

    if (inReasoningTag) {
      accumulated += "</think>\n\n";
      inReasoningTag = false;
    }

    if (!accumulated && buffer) {
      try {
        const fullJson = JSON.parse(buffer);
        accumulated = fullJson.choices[0].message.content || "";
      } catch (e) {}
    }

    botContent.innerHTML = formatMarkdown(accumulated, false);
    messages.push({ role: "assistant", content: accumulated });
  } catch (err) {
    if (err.name === "AbortError") {
      if (inReasoningTag) {
        accumulated += "</think>\n\n";
      }
      botContent.innerHTML = formatMarkdown(accumulated + "\n\n*(Generation stopped)*", false);
      messages.push({ role: "assistant", content: accumulated });
    } else {
      botContent.innerHTML = `<span style="color:#ef4444;">⚠️ Error: ${escapeHtml(err.message)}</span>`;
    }
  } finally {
    setStreaming(false);
    chatStream.scrollTop = chatStream.scrollHeight;
  }
}

// Fetch local connectivity info
fetch("/v1/local_info")
  .then(r => r.json())
  .then(data => {
    if (data.hotspot_url) {
      document.getElementById("link-hotspot").innerText = data.hotspot_url;
    }
    if (data.lan_url) {
      document.getElementById("link-lan").innerText = data.lan_url;
    } else if (data.primary_url && data.primary_url !== data.hotspot_url) {
      document.getElementById("link-lan").innerText = data.primary_url;
    } else {
      document.getElementById("link-lan").innerText = "Connect to Wi-Fi to view LAN link";
    }
    if (data.usb_url) {
      document.getElementById("link-usb").innerText = data.usb_url;
    } else {
      document.getElementById("link-usb").innerText = "Plug in USB and enable tethering";
    }
    const apiBase = data.lan_url || data.primary_url || data.hotspot_url;
    if (apiBase) {
      document.getElementById("link-api").innerText = apiBase + "/v1";
    }
  })
  .catch(() => {});

// Fetch status info
fetch("/v1/status")
  .then(r => r.json())
  .then(data => {
    if (data.knowledge_base) {
      const docs = data.knowledge_base.documents || 0;
      const chunks = data.knowledge_base.chunks || 0;
      document.getElementById("rag-badge").innerText = `RAG: ${docs} docs (${chunks} chunks)`;
    }
  })
  .catch(() => {});
</script>
</body>
</html>
"""

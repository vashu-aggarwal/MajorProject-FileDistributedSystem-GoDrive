"""
GoDrive Pipeline Microservice
==============================
Handles: LZW Compression, AES-GCM Encryption, Base64 Encoding
And the full pipeline in both directions.

Endpoints:
  GET  /health
  POST /compress          — LZW compress plain text → base64-encoded compressed bytes
  POST /decompress        — base64-encoded compressed bytes → plain text
  POST /encrypt           — base64 data → AES-GCM encrypt → base64
  POST /decrypt           — base64 encrypted data → AES-GCM decrypt → base64
  POST /encode            — plain text → base64
  POST /decode            — base64 → plain text
  POST /pipeline/encode   — FULL: text → LZW → AES → Base64 → text
  POST /pipeline/decode   — FULL: text → Base64 → AES → LZW → text
"""

from flask import Flask, request, jsonify
from cryptography.hazmat.primitives.ciphers.aead import AESGCM
import base64
import struct
import os
import logging

# ──────────────────────────────────────────────────────────
# Setup
# ──────────────────────────────────────────────────────────

logging.basicConfig(level=logging.INFO, format="%(asctime)s [%(levelname)s] %(message)s")
log = logging.getLogger("pipeline-service")

app = Flask(__name__)

# Default AES-256 key (32 bytes).
# "GoDrivePipeline-SecretKey-2025!!" = 32 chars
DEFAULT_KEY_HEX = os.environ.get(
    "GODRIVE_AES_KEY_HEX",
    "476f44726976655069706c696e652d5365637265744b65792d323032352121"  # 31 bytes; padded below
)


# ──────────────────────────────────────────────────────────
# LZW Implementation
# ──────────────────────────────────────────────────────────

def lzw_compress(text: str) -> bytes:
    """
    LZW compress a UTF-8 string.
    Returns packed bytes (each code = 2-byte unsigned big-endian short).
    Works for codes up to 65535 which is sufficient for typical text.
    """
    dict_size = 256
    dictionary: dict[str, int] = {chr(i): i for i in range(dict_size)}

    result: list[int] = []
    w = ""
    for c in text:
        wc = w + c
        if wc in dictionary:
            w = wc
        else:
            result.append(dictionary[w])
            dictionary[wc] = dict_size
            dict_size += 1
            w = c
    if w:
        result.append(dictionary[w])

    # Pack as big-endian unsigned 2-byte shorts
    return struct.pack(f">{len(result)}H", *result)


def lzw_decompress(data: bytes) -> str:
    """
    LZW decompress packed bytes (2-byte codes) back to the original string.
    """
    if not data:
        return ""

    n = len(data) // 2
    compressed = list(struct.unpack(f">{n}H", data))
    if not compressed:
        return ""

    dict_size = 256
    dictionary: dict[int, str] = {i: chr(i) for i in range(dict_size)}

    result: list[str] = []
    w = chr(compressed[0])
    result.append(w)

    for k in compressed[1:]:
        if k in dictionary:
            entry = dictionary[k]
        elif k == dict_size:
            entry = w + w[0]
        else:
            raise ValueError(f"Bad LZW code: {k}")
        result.append(entry)
        dictionary[dict_size] = w + entry[0]
        dict_size += 1
        w = entry

    return "".join(result)


# ──────────────────────────────────────────────────────────
# AES-GCM Helpers
# ──────────────────────────────────────────────────────────

def _parse_key(hex_key: str) -> bytes:
    """Convert hex string → exactly 32 bytes (AES-256)."""
    raw = bytes.fromhex(hex_key)
    # Pad/truncate to 32 bytes
    return (raw + b"\x00" * 32)[:32]


def aes_encrypt(data: bytes, key: bytes) -> bytes:
    """AES-256-GCM encrypt. Returns nonce (12 B) + ciphertext."""
    nonce = os.urandom(12)
    aesgcm = AESGCM(key)
    ciphertext = aesgcm.encrypt(nonce, data, None)
    return nonce + ciphertext


def aes_decrypt(data: bytes, key: bytes) -> bytes:
    """AES-256-GCM decrypt. Expects nonce (12 B) prepended to ciphertext."""
    nonce = data[:12]
    ciphertext = data[12:]
    aesgcm = AESGCM(key)
    return aesgcm.decrypt(nonce, ciphertext, None)


# ──────────────────────────────────────────────────────────
# Error helpers
# ──────────────────────────────────────────────────────────

def _err(msg: str, code: int = 400):
    return jsonify({"error": msg}), code


def _ok(result, **extras):
    return jsonify({"result": result, **extras})


# ──────────────────────────────────────────────────────────
# Health
# ──────────────────────────────────────────────────────────

@app.route("/health", methods=["GET"])
def health():
    return jsonify({"status": "ok", "service": "GoDrive Pipeline Microservice"})


# ──────────────────────────────────────────────────────────
# Individual endpoints
# ──────────────────────────────────────────────────────────

@app.route("/compress", methods=["POST"])
def compress():
    """
    Input:  { "text": "plain text" }
    Output: { "result": "<base64 of lzw-compressed bytes>",
              "original_size": N, "compressed_size": M }
    """
    body = request.get_json(silent=True)
    if not body or "text" not in body:
        return _err("Missing 'text' field")
    try:
        compressed = lzw_compress(body["text"])
        b64 = base64.b64encode(compressed).decode("utf-8")
        log.info(f"[COMPRESS] {len(body['text'])} chars → {len(compressed)} bytes")
        return _ok(b64, original_size=len(body["text"]), compressed_size=len(compressed))
    except Exception as e:
        log.error(f"[COMPRESS] Error: {e}")
        return _err(str(e), 500)


@app.route("/decompress", methods=["POST"])
def decompress():
    """
    Input:  { "data": "<base64 of lzw-compressed bytes>" }
    Output: { "result": "plain text" }
    """
    body = request.get_json(silent=True)
    if not body or "data" not in body:
        return _err("Missing 'data' field")
    try:
        compressed = base64.b64decode(body["data"])
        text = lzw_decompress(compressed)
        log.info(f"[DECOMPRESS] {len(compressed)} bytes → {len(text)} chars")
        return _ok(text)
    except Exception as e:
        log.error(f"[DECOMPRESS] Error: {e}")
        return _err(str(e), 500)


@app.route("/encrypt", methods=["POST"])
def encrypt():
    """
    Input:  { "data": "<base64 raw bytes>", "key": "<hex AES key>" }
    Output: { "result": "<base64 of nonce+ciphertext>" }
    """
    body = request.get_json(silent=True)
    if not body or "data" not in body:
        return _err("Missing 'data' field")
    try:
        raw = base64.b64decode(body["data"])
        key = _parse_key(body.get("key", DEFAULT_KEY_HEX))
        encrypted = aes_encrypt(raw, key)
        log.info(f"[ENCRYPT] {len(raw)} bytes → {len(encrypted)} bytes")
        return _ok(base64.b64encode(encrypted).decode("utf-8"))
    except Exception as e:
        log.error(f"[ENCRYPT] Error: {e}")
        return _err(str(e), 500)


@app.route("/decrypt", methods=["POST"])
def decrypt():
    """
    Input:  { "data": "<base64 nonce+ciphertext>", "key": "<hex AES key>" }
    Output: { "result": "<base64 decrypted bytes>" }
    """
    body = request.get_json(silent=True)
    if not body or "data" not in body:
        return _err("Missing 'data' field")
    try:
        encrypted = base64.b64decode(body["data"])
        key = _parse_key(body.get("key", DEFAULT_KEY_HEX))
        decrypted = aes_decrypt(encrypted, key)
        log.info(f"[DECRYPT] {len(encrypted)} bytes → {len(decrypted)} bytes")
        return _ok(base64.b64encode(decrypted).decode("utf-8"))
    except Exception as e:
        log.error(f"[DECRYPT] Error: {e}")
        return _err(str(e), 500)


@app.route("/encode", methods=["POST"])
def encode_b64():
    """
    Input:  { "text": "any string" }
    Output: { "result": "<base64>" }
    """
    body = request.get_json(silent=True)
    if not body or "text" not in body:
        return _err("Missing 'text' field")
    try:
        result = base64.b64encode(body["text"].encode("utf-8")).decode("utf-8")
        return _ok(result)
    except Exception as e:
        return _err(str(e), 500)


@app.route("/decode", methods=["POST"])
def decode_b64():
    """
    Input:  { "data": "<base64>" }
    Output: { "result": "original string" }
    """
    body = request.get_json(silent=True)
    if not body or "data" not in body:
        return _err("Missing 'data' field")
    try:
        result = base64.b64decode(body["data"]).decode("utf-8")
        return _ok(result)
    except Exception as e:
        return _err(str(e), 500)


# ──────────────────────────────────────────────────────────
# Full Pipeline endpoints
# ──────────────────────────────────────────────────────────

@app.route("/pipeline/encode", methods=["POST"])
def pipeline_encode():
    """
    FULL ENCODE PIPELINE:
      plain text → LZW compress → AES-GCM encrypt → Base64 encode → text

    Input:  { "text": "plain text content", "key": "<hex AES key>" }
    Output: { "result": "<base64 string>",
              "original_size": N, "encoded_size": M }
    """
    body = request.get_json(silent=True)
    if not body or "text" not in body:
        return _err("Missing 'text' field")
    try:
        text = body["text"]
        key  = _parse_key(body.get("key", DEFAULT_KEY_HEX))

        # Step 1 — LZW Compress
        compressed = lzw_compress(text)

        # Step 2 — AES-GCM Encrypt
        encrypted = aes_encrypt(compressed, key)

        # Step 3 — Base64 Encode
        encoded = base64.b64encode(encrypted).decode("utf-8")

        # Calculate compression ratio
        reduction = 100 - (len(compressed)/len(text)*100) if len(text) > 0 else 0
        log.info(
            f"\n"
            f"┌─────────────────────────────────────┐\n"
            f"│ PIPELINE ENCODE RESULTS             │\n"
            f"├─────────────────────────────────────┤\n"
            f"│ Original Size   : {len(text):>8} chars │\n"
            f"│ Compressed Size : {len(compressed):>8} bytes │ (-{reduction:.1f}%)\n"
            f"│ Encrypted Size  : {len(encrypted):>8} bytes │\n"
            f"│ Final Base64    : {len(encoded):>8} chars │\n"
            f"└─────────────────────────────────────┘"
        )
        return _ok(encoded, original_size=len(text), encoded_size=len(encoded))
    except Exception as e:
        log.error(f"[PIPELINE/ENCODE] Error: {e}")
        return _err(str(e), 500)


@app.route("/pipeline/decode", methods=["POST"])
def pipeline_decode():
    """
    FULL DECODE PIPELINE (reverse):
      text → Base64 decode → AES-GCM decrypt → LZW decompress → plain text

    Input:  { "data": "<base64 encoded string>", "key": "<hex AES key>" }
    Output: { "result": "original plain text" }
    """
    body = request.get_json(silent=True)
    if not body or "data" not in body:
        return _err("Missing 'data' field")
    try:
        data = body["data"]
        key  = _parse_key(body.get("key", DEFAULT_KEY_HEX))

        # Step 1 — Base64 Decode
        encrypted = base64.b64decode(data)

        # Step 2 — AES-GCM Decrypt
        compressed = aes_decrypt(encrypted, key)

        # Step 3 — LZW Decompress
        text = lzw_decompress(compressed)

        log.info(
            f"\n"
            f"┌─────────────────────────────────────┐\n"
            f"│ PIPELINE DECODE RESULTS             │\n"
            f"├─────────────────────────────────────┤\n"
            f"│ Input Base64    : {len(data):>8} chars │\n"
            f"│ Decrypted Size  : {len(compressed):>8} bytes │\n"
            f"│ Decompressed    : {len(text):>8} chars │\n"
            f"└─────────────────────────────────────┘"
        )
        return _ok(text)
    except Exception as e:
        log.error(f"[PIPELINE/DECODE] Error: {e}")
        return _err(str(e), 500)


# ──────────────────────────────────────────────────────────
# Entry point
# ──────────────────────────────────────────────────────────

if __name__ == "__main__":
    port = int(os.environ.get("PORT", 8000))
    log.info(f"Starting GoDrive Pipeline Microservice on port {port}")
    log.info("Pipeline: LZW Compression → AES-256-GCM Encryption → Base64 Encoding")
    app.run(host="0.0.0.0", port=port, debug=False)

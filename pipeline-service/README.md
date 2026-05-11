# GoDrive Pipeline Microservice

A Python Flask microservice that handles the text-file processing pipeline for GoDrive:

```
ENCODE:  plain text → LZW Compress → AES-256-GCM Encrypt → Base64 Encode → text
DECODE:  text → Base64 Decode → AES-256-GCM Decrypt → LZW Decompress → plain text
```

## Setup & Run

```bat
cd pipeline-service
run.bat
```

Or manually:

```bash
pip install -r requirements.txt
python app.py
```

Runs on **http://localhost:8000** by default.  
Set `PORT` environment variable to change port.  
Set `GODRIVE_AES_KEY_HEX` to override the AES key (must be a valid hex string, ≥32 bytes).

---

## API Reference

### `GET /health`
Health check.
```json
{ "status": "ok", "service": "GoDrive Pipeline Microservice" }
```

---

### `POST /pipeline/encode` ★ Main endpoint
Full encode pipeline: text → LZW → AES → Base64 → text

**Request:**
```json
{ "text": "hello world", "key": "<hex AES key (optional)>" }
```
**Response:**
```json
{ "result": "SGVsbG9...", "original_size": 11, "encoded_size": 64 }
```

---

### `POST /pipeline/decode` ★ Main endpoint
Full decode pipeline: text → Base64 → AES → LZW → text

**Request:**
```json
{ "data": "SGVsbG9...", "key": "<hex AES key (optional)>" }
```
**Response:**
```json
{ "result": "hello world" }
```

---

### `POST /compress`
LZW compress text, returns base64-encoded compressed bytes.

**Request:** `{ "text": "..." }`  
**Response:** `{ "result": "<base64>", "original_size": N, "compressed_size": M }`

---

### `POST /decompress`
LZW decompress base64-encoded bytes back to text.

**Request:** `{ "data": "<base64>" }`  
**Response:** `{ "result": "original text" }`

---

### `POST /encrypt`
AES-256-GCM encrypt base64-encoded data.

**Request:** `{ "data": "<base64>", "key": "<hex>" }`  
**Response:** `{ "result": "<base64>" }`

---

### `POST /decrypt`
AES-256-GCM decrypt base64-encoded data.

**Request:** `{ "data": "<base64>", "key": "<hex>" }`  
**Response:** `{ "result": "<base64>" }`

---

### `POST /encode`
Simple Base64 encode plain text.

**Request:** `{ "text": "..." }`  
**Response:** `{ "result": "<base64>" }`

---

### `POST /decode`
Simple Base64 decode to plain text.

**Request:** `{ "data": "<base64>" }`  
**Response:** `{ "result": "..." }`

---

## Enabling / Disabling in GoDrive

In `config/config.yaml`:

```yaml
pipeline:
  enabled: true          # set to false to bypass pipeline entirely
  service_url: "http://localhost:8000"
  aes_key_hex: "476f44726976655069706c696e652d5365637265744b65792d323032352121"
```

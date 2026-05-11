// Package pipeline provides a Go client for the GoDrive Pipeline Microservice.
//
// The pipeline microservice runs as a separate Python process and handles:
//   - LZW Compression / Decompression
//   - AES-256-GCM Encryption / Decryption
//   - Base64 Encoding / Decoding
//
// Integration is controlled by config.yaml:
//
//	pipeline:
//	  enabled: true
//	  service_url: "http://localhost:8000"
//	  aes_key_hex: "<64-char hex string>"
package pipeline

import (
	"bytes"
	"encoding/json"
	"fmt"
	"godrive/config"
	"io"
	"log"
	"net/http"
	"time"
)

// httpClient is shared across all calls to reuse TCP connections.
var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

// ──────────────────────────────────────────────────────────
// Public helpers
// ──────────────────────────────────────────────────────────

// IsEnabled returns true when the pipeline feature flag is on in config.yaml.
func IsEnabled() bool {
	return config.ReadConfig.Pipeline.Enabled
}

// ServiceURL returns the base URL of the pipeline microservice.
func ServiceURL() string {
	return config.ReadConfig.Pipeline.ServiceURL
}

// EncodeContent runs the full encode pipeline on the given plain-text content:
//
//	plain text → LZW Compress → AES-256-GCM Encrypt → Base64 Encode → text
//
// Returns the encoded text string ready to be stored/chunked.
func EncodeContent(text string) (string, error) {
	type req struct {
		Text string `json:"text"`
		Key  string `json:"key"`
	}
	payload := req{
		Text: text,
		Key:  config.ReadConfig.Pipeline.AESKeyHex,
	}

	result, err := postJSON(ServiceURL()+"/pipeline/encode", payload)
	if err != nil {
		return "", fmt.Errorf("pipeline encode: %w", err)
	}
	log.Printf("🔐 [Pipeline] Encoded %d chars → %d chars", len(text), len(result))
	return result, nil
}

// DecodeContent runs the full decode pipeline on the given encoded text:
//
//	text → Base64 Decode → AES-256-GCM Decrypt → LZW Decompress → plain text
//
// Returns the original plain-text content.
func DecodeContent(encoded string) (string, error) {
	type req struct {
		Data string `json:"data"`
		Key  string `json:"key"`
	}
	payload := req{
		Data: encoded,
		Key:  config.ReadConfig.Pipeline.AESKeyHex,
	}

	result, err := postJSON(ServiceURL()+"/pipeline/decode", payload)
	if err != nil {
		return "", fmt.Errorf("pipeline decode: %w", err)
	}
	log.Printf("🔓 [Pipeline] Decoded %d chars → %d chars", len(encoded), len(result))
	return result, nil
}

// ──────────────────────────────────────────────────────────
// Individual-step helpers (available for testing / direct use)
// ──────────────────────────────────────────────────────────

// Compress LZW-compresses text. Returns base64-encoded compressed bytes.
func Compress(text string) (string, error) {
	return postJSON(ServiceURL()+"/compress", map[string]string{"text": text})
}

// Decompress decompresses base64-encoded LZW data back to plain text.
func Decompress(data string) (string, error) {
	return postJSON(ServiceURL()+"/decompress", map[string]string{"data": data})
}

// Encrypt AES-GCM encrypts base64-encoded data. Returns base64.
func Encrypt(dataB64 string) (string, error) {
	return postJSON(ServiceURL()+"/encrypt", map[string]string{
		"data": dataB64,
		"key":  config.ReadConfig.Pipeline.AESKeyHex,
	})
}

// Decrypt AES-GCM decrypts base64-encoded data. Returns base64.
func Decrypt(dataB64 string) (string, error) {
	return postJSON(ServiceURL()+"/decrypt", map[string]string{
		"data": dataB64,
		"key":  config.ReadConfig.Pipeline.AESKeyHex,
	})
}

// Encode base64-encodes plain text.
func Encode(text string) (string, error) {
	return postJSON(ServiceURL()+"/encode", map[string]string{"text": text})
}

// Decode base64-decodes to plain text.
func Decode(data string) (string, error) {
	return postJSON(ServiceURL()+"/decode", map[string]string{"data": data})
}

// ──────────────────────────────────────────────────────────
// Health check
// ──────────────────────────────────────────────────────────

// CheckHealth pings the pipeline service and returns true if it is reachable.
func CheckHealth() bool {
	resp, err := httpClient.Get(ServiceURL() + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// ──────────────────────────────────────────────────────────
// Internal transport
// ──────────────────────────────────────────────────────────

// postJSON marshals payload as JSON, POSTs to url, and returns the "result" field.
func postJSON(url string, payload any) (string, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	resp, err := httpClient.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("http post to %s: %w", url, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("pipeline service returned %d: %s", resp.StatusCode, string(raw))
	}

	var parsed struct {
		Result string `json:"result"`
		Error  string `json:"error"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", fmt.Errorf("parse response JSON: %w", err)
	}
	if parsed.Error != "" {
		return "", fmt.Errorf("pipeline error: %s", parsed.Error)
	}

	return parsed.Result, nil
}

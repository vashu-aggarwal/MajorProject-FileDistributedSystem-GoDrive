package master

import (
	"crypto/sha256"
	"fmt"
	"godrive/pipeline"
	"log"
	"godrive/config"
)

type FileChunk struct {
	Index int    `json:"index"`
	Data  []byte `json:"data"`
	Hash  string `json:"hash"`
}

type FileStruct struct {
	Name   string
	Chunks []FileChunk
}

// BreakFilesIntoChunks splits a file's content into fixed-size chunks.
// If the pipeline is enabled, the content is first passed through the
// full encode pipeline (LZW Compress → AES-256-GCM Encrypt → Base64 Encode)
// before chunking so every byte stored on slave nodes is compressed, encrypted,
// and text-safe.
func BreakFilesIntoChunks(incomingFile uploadedFile) FileStruct {
	name, content := incomingFile.Name, incomingFile.Content

	// ── Pipeline encode (if enabled) ──────────────────────────────────────────
	if pipeline.IsEnabled() {
		encoded, err := pipeline.EncodeContent(content)
		if err != nil {
			log.Printf("🔴 [Pipeline] EncodeContent failed for '%s': %v — uploading raw content", name, err)
		} else {
			log.Printf("🔐 [Pipeline] '%s' encoded: %d chars → %d chars", name, len(content), len(encoded))
			content = encoded
		}
	}
	// ─────────────────────────────────────────────────────────────────────────

	chunkSize := config.ReadConfig.Master.ChunkSize
	// chunkSize := 4
	var createdFile FileStruct
	contentInBytes := []byte(content)
	createdFile.Name = name

	chunkInd := 0
	for i := 0; i < len(contentInBytes); i += chunkSize {
		end := min(len(contentInBytes), i+chunkSize)
		chunkHash := sha256.Sum256(contentInBytes[i:end])
		newChunk := FileChunk{
			Index: chunkInd,
			Data:  contentInBytes[i:end],
			Hash:  fmt.Sprintf("%x", chunkHash),
		}
		chunkInd++
		createdFile.Chunks = append(createdFile.Chunks, newChunk)
	}
	return createdFile
}

// MergeChunksToFile reassembles chunk data back into a single file.
// If the pipeline is enabled, the merged content is passed through the
// full decode pipeline (Base64 Decode → AES-256-GCM Decrypt → LZW Decompress)
// to recover the original plain-text content.
func MergeChunksToFile(downloadedFile FileStruct) uploadedFile {
	var contentInBytes []byte
	for _, chunk := range downloadedFile.Chunks {
		contentInBytes = append(contentInBytes, chunk.Data...)
	}
	merged := string(contentInBytes)

	// ── Pipeline decode (if enabled) ──────────────────────────────────────────
	if pipeline.IsEnabled() {
		decoded, err := pipeline.DecodeContent(merged)
		if err != nil {
			log.Printf("🔴 [Pipeline] DecodeContent failed for '%s': %v — returning raw merged content", downloadedFile.Name, err)
		} else {
			log.Printf("🔓 [Pipeline] '%s' decoded: %d chars → %d chars", downloadedFile.Name, len(merged), len(decoded))
			merged = decoded
		}
	}
	// ─────────────────────────────────────────────────────────────────────────

	return uploadedFile{
		Name:    downloadedFile.Name,
		Content: merged,
	}
}

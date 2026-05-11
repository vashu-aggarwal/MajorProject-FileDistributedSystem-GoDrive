package master

import (
	"crypto/sha256"
	"fmt"
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

// BreakFilesIntoChunks splits the file content (already compressed+encrypted bytes)
// into fixed-size chunks using the chunk_size from config.
func BreakFilesIntoChunks(incomingFile uploadedFile) FileStruct {
	name := incomingFile.Name
	chunkSize := config.ReadConfig.Master.ChunkSize
	contentInBytes := []byte(incomingFile.Content)

	var createdFile FileStruct
	createdFile.Name = name
	chunkInd := 0

	for i := 0; i < len(contentInBytes); i += chunkSize {
		end := min(len(contentInBytes), i+chunkSize)
		chunkData := contentInBytes[i:end]
		chunkHash := sha256.Sum256(chunkData)
		newChunk := FileChunk{
			Index: chunkInd,
			Data:  chunkData,
			Hash:  fmt.Sprintf("%x", chunkHash),
		}
		chunkInd++
		createdFile.Chunks = append(createdFile.Chunks, newChunk)
	}
	return createdFile
}

// MergeChunksToFile reassembles all chunks back into a single byte payload.
// The returned uploadedFile.Content holds the raw bytes as a string
// which will be passed to the pipeline for decryption and decompression.
func MergeChunksToFile(downloadedFile FileStruct) uploadedFile {
	var contentInBytes []byte
	for _, chunk := range downloadedFile.Chunks {
		contentInBytes = append(contentInBytes, chunk.Data...)
	}
	return uploadedFile{
		Name:    downloadedFile.Name,
		Content: string(contentInBytes),
	}
}

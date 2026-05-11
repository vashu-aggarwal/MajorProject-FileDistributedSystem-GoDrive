package master

import (
	"encoding/json"
	"log"
	"os"
	"sync"
)

type Metadata struct {
	mu     sync.Mutex
	Chunks map[string]map[int]*ChunkInfo
}

type ChunkInfo struct {
	ChunkHash     string
	SlaveNodeList []string
}

var metadata Metadata

func SaveMetaDataToFile() {
	data, err := json.MarshalIndent(metadata.Chunks, "", "  ")
	if err != nil {
		log.Println("Error marshaling metadata:", err)
		return
	}

	err = os.WriteFile("master.metadata.json", data, 0644)
	if err != nil {
		log.Println("Error writing metadata to file:", err)
		return
	}

	// log.Println("Metadata successfully saved to master.metadata.json")
}

func loadMetaDataFromFile() {
	metadata.mu.Lock()
	defer metadata.mu.Unlock()

	file, err := os.ReadFile("master.metadata.json")
	if err != nil {
		if os.IsNotExist(err) {
			metadata.Chunks = make(map[string]map[int]*ChunkInfo)
			return
		}
		log.Fatalln("Couldn't read metadata", err)
		return
	}

	json.Unmarshal(file, &metadata.Chunks)
}

func addChunkInfoToMetaData(fileName string, chunkHash string, ChunkIndex int, address string) {

	metadata.mu.Lock()
	defer metadata.mu.Unlock()

	if metadata.Chunks == nil {
		metadata.Chunks = make(map[string]map[int]*ChunkInfo)
	}

	if metadata.Chunks[fileName] == nil {
		metadata.Chunks[fileName] = make(map[int]*ChunkInfo)
	}

	obtainedChunkInfo, exists := metadata.Chunks[fileName][ChunkIndex]
	if exists {
		for _, addrItr := range obtainedChunkInfo.SlaveNodeList {
			if addrItr == address {
				return
			}
		}
		obtainedChunkInfo.SlaveNodeList = append(obtainedChunkInfo.SlaveNodeList, address)

	} else {
		metadata.Chunks[fileName][ChunkIndex] = &ChunkInfo{
			ChunkHash:     chunkHash,
			SlaveNodeList: []string{address},
		}
	}
	SaveMetaDataToFile()
}

func updateChunkHashInMetaData(fileName string, chunkHash string, ChunkIndex int, address string) {

	metadata.mu.Lock()
	defer metadata.mu.Unlock()

	obtainedChunkInfo := metadata.Chunks[fileName][ChunkIndex]
	obtainedChunkInfo.ChunkHash = chunkHash

	SaveMetaDataToFile()
}
func getNodeLoad(address string) int {
	metadata.mu.Lock()
	defer metadata.mu.Unlock()

	count := 0
	for _, chunkMap := range metadata.Chunks {
		for _, chunkInfo := range chunkMap {
			for _, p := range chunkInfo.SlaveNodeList {
				if p == address {
					count++
				}
			}
		}
	}
	return count
}

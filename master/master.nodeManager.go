package master

import (
	"fmt"
	"godrive/config"
	"log"
	"net"
	"sync"
)

type NodeSelector interface {
	GiveNode() config.Node
}

func DistriButeChunksToNode(file FileStruct) bool {
	var wg sync.WaitGroup
	chunkSuccessMap := make([]int, len(file.Chunks))
	var mu sync.Mutex

	for ind, chunk := range file.Chunks {
		for replication := 0; replication < config.ReadConfig.Master.ReplicationFactor; replication++ {
			wg.Add(1)
			go func(index int, chunk FileChunk) {
				defer wg.Done()
				selectedNode := MyNodeSelector.GiveNode()
				success, _ := SendDataToSlave(selectedNode, chunk)
				if success {
					// Metrics.RecordNodeRequest(selectedNode.Port, len(chunk.Data))
					mu.Lock()
					chunkSuccessMap[index] += 1
					mu.Unlock()
					addChunkInfoToMetaData(file.Name, chunk.Hash, chunk.Index, fmt.Sprintf("%s:%s", selectedNode.Host, selectedNode.Port))
				}
			}(ind, chunk)
		}
	}
	wg.Wait()

	for index, count := range chunkSuccessMap {
		if count < config.ReadConfig.Master.WriteQuorum {
			log.Println("Write quorum not met for chunk:", index, "in file:", file.Name)
			DeleteFileIfQuoramFails(file.Name)
			return false
		}
	}
	return true
}

// func CompareChunksAndUpdate(file FileStruct) {
// 	var wg sync.WaitGroup
// 	currentFileMap := metadata.Chunks[file.Name] // ind -> fileChunk

// 	for index, chunk := range file.Chunks {
// 		if index < len(currentFileMap) {
// 			mapChunkInfo := currentFileMap[index]
// 			if chunk.Hash == mapChunkInfo.ChunkHash {
// 				continue
// 			} else {
// 				// update FileChunk On slaveNodes
// 			}
// 		} else {
// 			// save chunks to
// 		}

// 		for i := 0; i < len(mapChunkInfo.SlaveNodeList); i++ {
// 			port := mapChunkInfo.SlaveNodeList[i]
// 			wg.Add(1)
// 			go func(chunk FileChunk, port string) {
// 				defer wg.Done()
// 				selectedNode := config.Node{Host: "127.0.0.1", Port: port}

//					_, err := SendDataToSlave(selectedNode, chunk)
//					if err != nil {
//						log.Println("Error sending data to slave:", err)
//						return
//					}
//					updateChunkHashInMetaData(file.Name, chunk.Hash, chunk.Index, selectedNode.Port)
//				}(chunk, port)
//			}
//		}
//		wg.Wait()
//	}

func deleteChunkFromSlaves(chunkInfo *ChunkInfo, ackChan chan bool, wg *sync.WaitGroup) {
	defer wg.Done()
	chunkHash := chunkInfo.ChunkHash
	allDeleted := true

	for _, address := range chunkInfo.SlaveNodeList {
		host, port, _ := net.SplitHostPort(address)
		slaveNode := config.Node{Host: host, Port: port}
		err := RequestDeleteFromSlave(slaveNode, chunkHash)
		if err != nil {
			log.Printf("Failed to delete chunk %s from node %s: %v\n", chunkHash, address, err)
			allDeleted = false
		} else {
			log.Printf("Deleted chunk %s from node %s\n", chunkHash, address)
		}
	}

	ackChan <- allDeleted
}

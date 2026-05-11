package master

import (
	"fmt"
	"godrive/config"
	"log"
	"time"
)

func StartHeartBeat() {
	log.Println("Heartbeat service initiated!")
	count := 1
	for {
		time.Sleep(5 * time.Second)
		log.Printf("---------------- Heartbeat: %v ----------------", count)
		for _, slaveNode := range config.ReadConfig.SlaveNodes {
			heartbeat := SendHeartBeatToSlave(slaveNode)
			if !heartbeat {
				log.Printf("🔴 Master <=xxxxxxxx= %s", slaveNode.Port)
				go removeDeadNodeFromSlaveList(slaveNode)
			} else {
				log.Printf("🟢 Master <========== %s", slaveNode.Port)
			}
		}
		log.Println("-------------------------------------------------")
		count++
	}
}

func removeDeadNodeFromSlaveList(deadNode config.Node) {
	deleteIndex := -1
	for index, node := range config.ReadConfig.SlaveNodes {
		if node.Port == deadNode.Port && node.Host == deadNode.Host {
			deleteIndex = index
			break
		}
	}
	if deleteIndex == -1 {
		log.Printf("Node %s:%s not found in slave list", deadNode.Host, deadNode.Port)
		return
	}
	config.ReadConfig.SlaveNodes = append(config.ReadConfig.SlaveNodes[:deleteIndex], config.ReadConfig.SlaveNodes[deleteIndex+1:]...)
	for backupIndex, backupNode := range config.ReadConfig.BackupNodes {
		if heartbeat := SendHeartBeatToSlave(backupNode); heartbeat {
			handleDataTransfer(deadNode, backupNode)
			// Add dead node to BackupNodes
			config.ReadConfig.BackupNodes = append(config.ReadConfig.BackupNodes, deadNode)
			// Remove promoted backup node from BackupNodes and add to SlaveNodes
			config.ReadConfig.BackupNodes = append(config.ReadConfig.BackupNodes[:backupIndex], config.ReadConfig.BackupNodes[backupIndex+1:]...)
			config.ReadConfig.SlaveNodes = append(config.ReadConfig.SlaveNodes, backupNode)
			fmt.Println("\n-------------------------------------------------------------------------------------")
			fmt.Println("Slave Node List:\n", config.ReadConfig.SlaveNodes)
			fmt.Println("\nBackup Node List:", config.ReadConfig.BackupNodes, "")
			fmt.Print("-------------------------------------------------------------------------------------\n\n")
			break
		}
	}
}
func handleDataTransfer(fromNode config.Node, toNode config.Node) {

	log.Printf("\n\n Transferring data: [🟥]%s:%s ---> %s:%s[🟩]\n\n", fromNode.Host, fromNode.Port, toNode.Host, toNode.Port)
	metadata.mu.Lock()
	defer metadata.mu.Unlock()

	success := true
	fromNodeID := fmt.Sprintf("%s:%s", fromNode.Host, fromNode.Port)
	toNodeID := fmt.Sprintf("%s:%s", toNode.Host, toNode.Port)

	for _, chunkMap := range metadata.Chunks {
		for _, chunkInfo := range chunkMap {
			contains := false
			sourceNode := ""
			for _, node := range chunkInfo.SlaveNodeList {
				if node == fromNodeID {
					contains = true
				} else {
					sourceNode = node
				}
			}
			if contains && sourceNode != toNodeID && sourceNode != fromNodeID {
				// Extract port from sourceNode (format: "host:port")
				sourcePort := ""
				if len(sourceNode) > 0 {
					// Parse host:port format
					parts := len(sourceNode)
					// Find the last colon to get the port
					for i := parts - 1; i >= 0; i-- {
						if sourceNode[i] == ':' {
							sourcePort = sourceNode[i+1:]
							break
						}
					}
				}
				if sourcePort != "" {
					success = success && SendInterNodeTransferRequest(sourcePort, toNode.Port, chunkInfo.ChunkHash)
				}
			}
		}
	}
	if !success {
		log.Printf("🔴 InterNode chunk transfer failed from %s:%s to %s:%s\n", fromNode.Host, fromNode.Port, toNode.Host, toNode.Port)
	} else {
		log.Printf("🟢 InterNode chunk transfer successful from %s:%s to %s:%s\n", fromNode.Host, fromNode.Port, toNode.Host, toNode.Port)
	}
	for _, chunkMap := range metadata.Chunks {
		for _, chunkInfo := range chunkMap {
			newList := []string{}
			replaced := false
			for _, nodeID := range chunkInfo.SlaveNodeList {
				if nodeID == fromNodeID {
					replaced = true
					continue
				}
				newList = append(newList, nodeID)
			}
			if replaced {
				newList = append(newList, toNodeID)
				chunkInfo.SlaveNodeList = newList
			}
		}
	}
	SaveMetaDataToFile()
	log.Printf("✅ Metadata updated: replaced %s:%s with %s:%s", fromNode.Host, fromNode.Port, toNode.Host, toNode.Port)
}

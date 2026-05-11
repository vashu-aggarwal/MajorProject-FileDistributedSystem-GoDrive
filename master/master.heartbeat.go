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
		time.Sleep(10 * time.Second)
		log.Printf("---------------- Heartbeat: %v ----------------", count)
		for _, slaveNode := range config.ReadConfig.SlaveNodes {
			heartbeat := SendHeartBeatToSlave(slaveNode)
			if !heartbeat {
				log.Printf("🔴 Master <=xxxxxxxx= %s:%s", slaveNode.Host, slaveNode.Port)
				go removeDeadNodeFromSlaveList(slaveNode.Host, slaveNode.Port)
			} else {
				log.Printf("🟢 Master <========== %s:%s", slaveNode.Host, slaveNode.Port)
			}
		}
		log.Println("-------------------------------------------------")
		count++
	}
}

func removeDeadNodeFromSlaveList(host string, port string) {
	deleteIndex, deleteNode := -1, config.Node{}
	for index, node := range config.ReadConfig.SlaveNodes {
		if node.Host == host && node.Port == port {
			deleteIndex = index
			deleteNode = node
			break
		}
	}

	if deleteIndex == -1 {
		log.Printf("⚠️  Node %s:%s not found in SlaveNodes", host, port)
		return
	}

	config.ReadConfig.SlaveNodes = append(config.ReadConfig.SlaveNodes[:deleteIndex], config.ReadConfig.SlaveNodes[deleteIndex+1:]...)

	for bIndex, bNode := range config.ReadConfig.BackupNodes {
		if heartbeat := SendHeartBeatToSlave(bNode); heartbeat {
			handleDataTransfer(fmt.Sprintf("%s:%s", host, port), fmt.Sprintf("%s:%s", bNode.Host, bNode.Port))

			// Remove promoted node from BackupNodes
			config.ReadConfig.BackupNodes = append(config.ReadConfig.BackupNodes[:bIndex], config.ReadConfig.BackupNodes[bIndex+1:]...)

			// Add dead node to BackupNodes
			config.ReadConfig.BackupNodes = append(config.ReadConfig.BackupNodes, deleteNode)

			// Add promoted node to SlaveNodes
			config.ReadConfig.SlaveNodes = append(config.ReadConfig.SlaveNodes, bNode)

			// Update the active node selector with the new slave list
			algoConfigMu.Lock()
			MyNodeSelector = InitNodeSelector(AlgoConfig.NodeSelectorAlgo, config.ReadConfig.SlaveNodes)
			algoConfigMu.Unlock()

			fmt.Println("\n-------------------------------------------------------------------------------------")
			fmt.Println("Slave Node List:\n", config.ReadConfig.SlaveNodes)
			fmt.Println("\nBackup Node List:", config.ReadConfig.BackupNodes, "")

			fmt.Print("-------------------------------------------------------------------------------------\n\n")
			break
		}
	}
}

func handleDataTransfer(from string, to string) {

	log.Printf("\n\n Transferring data: [🟥]%v ---> %v[🟩]\n\n", from, to)
	metadata.mu.Lock()
	defer metadata.mu.Unlock()

	success := true

	for _, chunkMap := range metadata.Chunks {
		for _, chunkInfo := range chunkMap {
			contains := false
			sourceNode := from
			for _, node := range chunkInfo.SlaveNodeList {
				if node == from {
					contains = true
				} else {
					sourceNode = node
				}
			}
			if contains && sourceNode != to && sourceNode != from {
				log.Printf("📤 Master: Requesting internode transfer of chunk %s from %s to %s", chunkInfo.ChunkHash, sourceNode, to)
				success = success && SendInterNodeTransferRequest(sourceNode, to, chunkInfo.ChunkHash)
			} else if contains {
				log.Printf("⚠️  Master: No alternative replica found for chunk %s. Node %s was the only known holder.", chunkInfo.ChunkHash, from)
			}
		}
	}
	if !success {
		log.Printf("🔴 InterNode chunk transfer failed from %s to %s\n", from, to)
	} else {
		log.Printf("🟢 InterNode chunk transfer successful from %s to %s\n", from, to)
	}
	for _, chunkMap := range metadata.Chunks {
		for _, chunkInfo := range chunkMap {
			newList := []string{}
			replaced := false
			for _, nodeAddress := range chunkInfo.SlaveNodeList {
				if nodeAddress == from {
					replaced = true
					continue
				}
				newList = append(newList, nodeAddress)
			}
			if replaced {
				newList = append(newList, to)
				chunkInfo.SlaveNodeList = newList
			}
		}
	}
	SaveMetaDataToFile()
	log.Printf("✅ Metadata updated: replaced %s with %s", from, to)
}

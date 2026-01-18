package master

import (
	"encoding/json"
	"fmt"
	"godrive/config"
	"godrive/master/cache"
	"log"
	"net/http"
	"sync"
)

type uploadedFile struct {
	Name    string `json:"fileName"`
	Content string `json:"content"`
}

var MyNodeSelector *RoundRobinNodeSelector
var FileCache cache.Cache

func StartMasterHttp() {
	port := config.ReadConfig.Master.HttpPort
	fullAddress := fmt.Sprintf(":%d", port)
	MyNodeSelector = NewRoundRobinSelector(config.ReadConfig.SlaveNodes)

	FileCache = cache.NewLRUCache(5) // change policy here
	log.Println(" Cache Initialized: LRU Cache | Capacity = 5")

	http.HandleFunc("/", healthCheck)
	http.HandleFunc("/upload", handleFileUpload)
	http.HandleFunc("/download", handleFileDownload)
	// http.HandleFunc("/update", handleFileUpdate)
	http.HandleFunc("/delete", handleFileDelete)

	log.Printf(`
╔════════════════════════════════════════╗
║   HTTP SERVER STARTED ON PORT %v     ║
╚════════════════════════════════════════╝`, port)
	err := http.ListenAndServe(fullAddress, nil)
	if err != nil {
		log.Fatal("HTTP server crashed")
	}
}
func DeleteFileIfQuoramFails(filename string) {

	metadata.mu.Lock()
	fileInfo, exists := metadata.Chunks[filename]
	metadata.mu.Unlock()

	if !exists {
		return
	}

	deleteAckChannel := make(chan bool)
	var wg sync.WaitGroup

	for _, chunkInfo := range fileInfo {
		wg.Add(1)
		go deleteChunkFromSlaves(chunkInfo, deleteAckChannel, &wg)
	}

	go func() {
		wg.Wait()
		close(deleteAckChannel)
	}()

	allSuccessful := true
	for ack := range deleteAckChannel {
		if !ack {
			allSuccessful = false
			break
		}
	}

	if allSuccessful {
		metadata.mu.Lock()
		delete(metadata.Chunks, filename)
		metadata.mu.Unlock()
		SaveMetaDataToFile()
	}
}

// Route handler functions
func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Http server looks good"))
}

func handleFileUpload(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests allowed in this route", http.StatusBadRequest)
		return
	}
	var incomingFile uploadedFile
	err := json.NewDecoder(r.Body).Decode(&incomingFile)
	if err != nil {
		http.Error(w, "Bad format file", http.StatusBadRequest)
		return
	}
	if incomingFile.Name == "" || incomingFile.Content == "" {
		http.Error(w, "FileName or content is empty", http.StatusBadRequest)
		return
	}
	if _, exists := metadata.Chunks[incomingFile.Name]; exists {
		http.Error(w, "File already present in system. Delete file first to use upload or use 'update'.", http.StatusConflict)
		return
	}
	createdFile := BreakFilesIntoChunks(incomingFile)
	if success := DistriButeChunksToNode(createdFile); success {
		log.Println("────────────────────────────────────────")
		log.Printf("✅  Master: Splitted file into %v chunks", len(createdFile.Chunks))
		log.Println("────────────────────────────────────────")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("Accepted file: '%v'.", incomingFile.Name)))
	} else {
		log.Println("────────────────────────────────────────")
		log.Printf("⚠️  Master: Failed to upload file")
		log.Println("────────────────────────────────────────")
		http.Error(w, "Failed to upload file", http.StatusInternalServerError)
	}

}

func handleFileDownload(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests allowed in this route", http.StatusBadRequest)
		return
	}
	filename := r.URL.Query().Get("filename")
	if filename == "" {
		http.Error(w, "No key found", http.StatusBadRequest)
		return
	}
	//cache check
	if cachedData, ok := FileCache.Get(filename); ok {
		log.Println("⚡ Cache HIT for", filename)
		w.Write([]byte(cachedData))
		return
	}
	indexToFilechunkMap, exists := metadata.Chunks[filename]
	if !exists {
		http.Error(w, "No such file present in the system.", http.StatusNotFound)
		return
	}
	downloadedFile := FileStruct{
		Name:   filename,
		Chunks: make([]FileChunk, len(indexToFilechunkMap)),
	}
	incomingChunksChannel := make(chan FileChunk)
	var wg sync.WaitGroup
	for index, chunkInfo := range indexToFilechunkMap {
		wg.Add(1)
		go getChunk(index, chunkInfo, incomingChunksChannel, &wg)
	}
	go func() {
		wg.Wait()
		close(incomingChunksChannel)
	}()
	for incomingFileChunk := range incomingChunksChannel {
		if incomingFileChunk.Index == -1 {
			log.Println("────────────────────────────────────────")
			log.Printf("⚠️  Master: Error while downloading file")
			log.Println("────────────────────────────────────────")
			http.Error(w, "Error in Downloading file", http.StatusInternalServerError)
			return
		}
		downloadedFile.Chunks[incomingFileChunk.Index] = incomingFileChunk
	}
	log.Println("────────────────────────────────────────")
	log.Printf("✅ Master: Download %v sucessfully\n", downloadedFile.Name)
	log.Println("────────────────────────────────────────")
	createdFileAfterMerge := MergeChunksToFile(downloadedFile)
	// 🔵 CACHE PUT (ADD HERE)
	FileCache.Put(filename, createdFileAfterMerge.Content)
	log.Printf(" CACHE PUT | File: %s | CacheSize: %d\n", filename, FileCache.Size())

	createdFileJson, err := json.Marshal(createdFileAfterMerge)
	if err != nil {
		log.Println("Error")
		return
	}
	w.Write(createdFileJson)
}
func getChunk(index int, chunkInfo *ChunkInfo, channelToSendChunk chan FileChunk, wg *sync.WaitGroup) {
	defer wg.Done()
	chunkHash, slaveNodeList := chunkInfo.ChunkHash, chunkInfo.SlaveNodeList
	for ind := 0; ind < len(slaveNodeList); ind++ {
		obtainedFileChunk, err := RequestChunkFromSlave(slaveNodeList[ind], chunkHash)
		if err == nil {
			obtainedFileChunk.Index = index
			channelToSendChunk <- obtainedFileChunk
			return
		}
	}
	channelToSendChunk <- FileChunk{Index: -1}
}

func handleFileDelete(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodDelete {
		http.Error(w, "Only DELETE requests are allowed on this route", http.StatusMethodNotAllowed)
		return
	}

	filename := r.URL.Query().Get("filename")
	if filename == "" {
		http.Error(w, "No filename provided", http.StatusBadRequest)
		return
	}

	metadata.mu.Lock()
	fileInfo, exists := metadata.Chunks[filename]
	metadata.mu.Unlock()

	if !exists {
		http.Error(w, "No such file present in the system.", http.StatusNotFound)
		return
	}

	deleteAckChannel := make(chan bool)
	var wg sync.WaitGroup

	for _, chunkInfo := range fileInfo {
		wg.Add(1)
		go deleteChunkFromSlaves(chunkInfo, deleteAckChannel, &wg)
	}

	go func() {
		wg.Wait()
		close(deleteAckChannel)
	}()

	allSuccessful := true
	for ack := range deleteAckChannel {
		if !ack {
			allSuccessful = false
			break
		}
	}

	if allSuccessful {
		metadata.mu.Lock()
		delete(metadata.Chunks, filename)
		metadata.mu.Unlock()
		SaveMetaDataToFile()
		// 🔵 CACHE INVALIDATION (ADD HERE)
		FileCache.Delete(filename)
		log.Printf("🗑️ CACHE DELETE | File: %s | CacheSize: %d", filename, FileCache.Size())
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Deleted '%s' from the system.\n", filename)
	} else {
		log.Println("Couldn't delete file")
		http.Error(w, "Failed to delete file", http.StatusInternalServerError)
	}
}

// func handleFileUpdate(w http.ResponseWriter, r *http.Request) {

// 	if r.Method != http.MethodPost {
// 		http.Error(w, "Only POST requests allowed in this route", http.StatusBadRequest)
// 		return
// 	}
// 	var incomingFile uploadedFile
// 	err := json.NewDecoder(r.Body).Decode(&incomingFile)
// 	if err != nil {
// 		http.Error(w, "Bad format file", http.StatusBadRequest)
// 		return
// 	}
// 	if incomingFile.Name == "" || incomingFile.Content == "" {
// 		http.Error(w, "FileName or content is empty", http.StatusBadRequest)
// 		return
// 	}
// 	if metadata.Chunks[incomingFile.Name] == nil {
// 		http.Error(w, "No such file found to update", http.StatusNotFound)
// 		return
// 	}
// 	createdFile := BreakFilesIntoChunks(incomingFile)
// 	CompareChunksAndUpdate(createdFile)
// 	w.WriteHeader(http.StatusOK)
// 	w.Write([]byte(fmt.Sprintf("Accepted file: %v", incomingFile.Name)))
// }

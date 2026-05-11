package master

import (
	"encoding/json"
	"fmt"
	"godrive/config"
	"log"
	"net/http"
	"sync"
	"time"
)

type uploadedFile struct {
	Name    string `json:"fileName"`
	Content string `json:"content"`
}

type AlgorithmConfig struct {
	CacheAlgorithm   string `json:"cacheAlgorithm"`
	NodeSelectorAlgo string `json:"nodeSelectorAlgo"`
	CacheCapacity    int    `json:"cacheCapacity"`
}

var MyNodeSelector NodeSelector
var CurrentCache CacheProvider
var AlgoConfig AlgorithmConfig
var algoConfigMu sync.RWMutex

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func StartMasterHttp() {
	port := config.ReadConfig.Master.HttpPort
	fullAddress := fmt.Sprintf(":%d", port)

	// Initialize default algorithms
	AlgoConfig = AlgorithmConfig{
		CacheAlgorithm:   "lru",
		NodeSelectorAlgo: "leastNode",
		CacheCapacity:    1,
	}

	MyNodeSelector = NewLeastNodeSelector(config.ReadConfig.SlaveNodes)
	CurrentCache = InitCache("lru", 1)
	Metrics.Init(AlgoConfig.CacheAlgorithm, AlgoConfig.NodeSelectorAlgo)

	http.HandleFunc("/", corsMiddleware(healthCheck))
	http.HandleFunc("/upload", corsMiddleware(handleFileUpload))
	http.HandleFunc("/download", corsMiddleware(handleFileDownload))
	http.HandleFunc("/delete", corsMiddleware(handleFileDelete))
	http.HandleFunc("/update", corsMiddleware(handleFileUpdate))
	http.HandleFunc("/config/algorithms", corsMiddleware(handleAlgorithmConfig))
	http.HandleFunc("/config/cache-status", corsMiddleware(handleCacheStatus))
	http.HandleFunc("/metrics/performance", corsMiddleware(handlePerformanceMetrics))
	http.HandleFunc("/metrics/reset", corsMiddleware(handleMetricsReset))

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
	startTime := time.Now()
	success := false
	bytesTransferred := 0
	defer func() {
		Metrics.RecordRequest("upload", success, time.Since(startTime), bytesTransferred)
	}()

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests allowed in this route", http.StatusBadRequest)
		return
	}

	var incomingFile uploadedFile

	// Check if this is FormData (multipart) or JSON
	contentType := r.Header.Get("Content-Type")

	if contentType == "application/json" {
		// Handle JSON request (text mode from frontend)
		err := json.NewDecoder(r.Body).Decode(&incomingFile)
		if err != nil {
			http.Error(w, "Bad format file", http.StatusBadRequest)
			return
		}
	} else {
		// Handle multipart form data (file mode from frontend)
		err := r.ParseMultipartForm(100 << 20) // 100 MB max
		if err != nil {
			http.Error(w, "Failed to parse form data", http.StatusBadRequest)
			return
		}

		// Get filename from form
		fileName := r.FormValue("fileName")
		if fileName == "" {
			http.Error(w, "FileName is empty", http.StatusBadRequest)
			return
		}

		// Get file from form
		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Failed to read file from form", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Read file content as bytes
		fileBytes := make([]byte, 0)
		buffer := make([]byte, 1024)
		for {
			n, err := file.Read(buffer)
			if n > 0 {
				fileBytes = append(fileBytes, buffer[:n]...)
			}
			if err != nil {
				break
			}
		}

		// Convert bytes to string (base64 encoded for binary files)
		incomingFile.Name = fileName
		incomingFile.Content = string(fileBytes)
	}

	if incomingFile.Name == "" || incomingFile.Content == "" {
		http.Error(w, "FileName or content is empty", http.StatusBadRequest)
		return
	}
	bytesTransferred = len([]byte(incomingFile.Content))
	if _, exists := metadata.Chunks[incomingFile.Name]; exists {
		http.Error(w, "File already present in system. Delete file first to use upload or use 'update'.", http.StatusConflict)
		return
	}
	createdFile := BreakFilesIntoChunks(incomingFile)
	if distributeSuccess := DistriButeChunksToNode(createdFile); distributeSuccess {
		log.Println("────────────────────────────────────────")
		log.Printf("✅  Master: Splitted file into %v chunks", len(createdFile.Chunks))
		log.Println("────────────────────────────────────────")
		success = true
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
	startTime := time.Now()
	success := false
	bytesTransferred := 0
	defer func() {
		Metrics.RecordRequest("download", success, time.Since(startTime), bytesTransferred)
	}()

	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests allowed in this route", http.StatusBadRequest)
		return
	}
	filename := r.URL.Query().Get("filename")
	if filename == "" {
		http.Error(w, "No key found", http.StatusBadRequest)
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
	bytesTransferred = len([]byte(createdFileAfterMerge.Content))
	createdFileJson, err := json.Marshal(createdFileAfterMerge)
	if err != nil {
		log.Println("Error")
		return
	}
	success = true
	w.Write(createdFileJson)
}
func getChunk(index int, chunkInfo *ChunkInfo, channelToSendChunk chan FileChunk, wg *sync.WaitGroup) {
	defer wg.Done()
	chunkHash, slaveNodeList := chunkInfo.ChunkHash, chunkInfo.SlaveNodeList

	// Try to get from cache first
	algoConfigMu.RLock()
	cache := CurrentCache
	algoConfigMu.RUnlock()

	if cachedData, found := cache.Get(chunkHash); found {
		log.Printf("🟡 Cache HIT for chunk: %s\n", chunkHash)
		Metrics.RecordCacheResult(true)
		channelToSendChunk <- FileChunk{
			Index: index,
			Data:  []byte(cachedData),
			Hash:  chunkHash,
		}
		return
	}

	// Cache miss: fetch from slave
	Metrics.RecordCacheResult(false)
	for ind := 0; ind < len(slaveNodeList); ind++ {
		obtainedFileChunk, err := RequestChunkFromSlave(slaveNodeList[ind], chunkHash)
		if err == nil {
			obtainedFileChunk.Index = index
			Metrics.RecordNodeRequest(slaveNodeList[ind], len(obtainedFileChunk.Data))

			// Store in cache
			cache.Put(chunkHash, string(obtainedFileChunk.Data))
			log.Printf("🟢 Cache STORE for chunk: %s (from slave %s)\n", chunkHash, slaveNodeList[ind])

			channelToSendChunk <- obtainedFileChunk
			return
		}
	}
	channelToSendChunk <- FileChunk{Index: -1}
}

func handleFileDelete(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	success := false
	defer func() {
		Metrics.RecordRequest("delete", success, time.Since(startTime), 0)
	}()

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
		success = true
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Deleted '%s' from the system.\n", filename)
	} else {
		log.Println("Couldn't delete file")
		http.Error(w, "Failed to delete file", http.StatusInternalServerError)
	}
}

// handleFileUpdate handles PUT requests to update/replace a file
func handleFileUpdate(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	success := false
	bytesTransferred := 0
	defer func() {
		Metrics.RecordRequest("update", success, time.Since(startTime), bytesTransferred)
	}()

	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPut {
		http.Error(w, "Only PUT requests are allowed on this route", http.StatusMethodNotAllowed)
		return
	}

	var uploadRequest struct {
		FileName string `json:"fileName"`
		Content  string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&uploadRequest); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if uploadRequest.FileName == "" || uploadRequest.Content == "" {
		http.Error(w, "fileName and content are required", http.StatusBadRequest)
		return
	}
	bytesTransferred = len([]byte(uploadRequest.Content))

	// Delete the old file first
	metadata.mu.Lock()
	_, exists := metadata.Chunks[uploadRequest.FileName]
	metadata.mu.Unlock()

	if exists {
		// Delete all chunks of the old file
		deleteAckChannel := make(chan bool)
		var wg sync.WaitGroup

		metadata.mu.Lock()
		fileInfo := metadata.Chunks[uploadRequest.FileName]
		metadata.mu.Unlock()

		for _, chunkInfo := range fileInfo {
			wg.Add(1)
			go deleteChunkFromSlaves(chunkInfo, deleteAckChannel, &wg)
		}

		go func() {
			wg.Wait()
			close(deleteAckChannel)
		}()

		for range deleteAckChannel {
			// Consume all acks
		}

		metadata.mu.Lock()
		delete(metadata.Chunks, uploadRequest.FileName)
		metadata.mu.Unlock()
	}

	// Now upload as new file using same mechanism as regular upload
	incomingFile := uploadedFile{
		Name:    uploadRequest.FileName,
		Content: uploadRequest.Content,
	}
	createdFile := BreakFilesIntoChunks(incomingFile)
	if distributeSuccess := DistriButeChunksToNode(createdFile); distributeSuccess {
		log.Println("────────────────────────────────────────")
		log.Printf("✅  Master: Updated file '%s' with %v chunks", uploadRequest.FileName, len(createdFile.Chunks))
		log.Println("────────────────────────────────────────")
		success = true
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": fmt.Sprintf("File '%s' updated successfully", uploadRequest.FileName),
		})
	} else {
		log.Println("────────────────────────────────────────")
		log.Printf("⚠️  Master: Failed to update file")
		log.Println("────────────────────────────────────────")
		http.Error(w, "Failed to update file", http.StatusInternalServerError)
	}
}

// handleAlgorithmConfig handles GET/POST for algorithm configuration
func handleAlgorithmConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodGet {
		algoConfigMu.RLock()
		configCopy := AlgoConfig
		algoConfigMu.RUnlock()

		response := map[string]interface{}{
			"status":             "success",
			"currentCache":       configCopy.CacheAlgorithm,
			"currentSelector":    configCopy.NodeSelectorAlgo,
			"cacheCapacity":      configCopy.CacheCapacity,
			"availableCaches":    []string{"lru", "lfu", "fifo", "arc"},
			"availableSelectors": []string{"roundRobin", "random", "leastNode", "powerOfTwo"},
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		return
	}

	if r.Method == http.MethodPost {
		var newConfig AlgorithmConfig
		err := json.NewDecoder(r.Body).Decode(&newConfig)
		if err != nil {
			http.Error(w, "Invalid request format", http.StatusBadRequest)
			return
		}

		// Validate cache algorithm
		validCaches := map[string]bool{"lru": true, "lfu": true, "fifo": true, "arc": true}
		if !validCaches[newConfig.CacheAlgorithm] {
			http.Error(w, fmt.Sprintf("Invalid cache algorithm: %s", newConfig.CacheAlgorithm), http.StatusBadRequest)
			return
		}

		// Validate node selector algorithm
		validSelectors := map[string]bool{"roundRobin": true, "random": true, "leastNode": true, "powerOfTwo": true}
		if !validSelectors[newConfig.NodeSelectorAlgo] {
			http.Error(w, fmt.Sprintf("Invalid node selector algorithm: %s", newConfig.NodeSelectorAlgo), http.StatusBadRequest)
			return
		}

		// Validate cache capacity
		if newConfig.CacheCapacity <= 0 {
			http.Error(w, "Cache capacity must be greater than 0", http.StatusBadRequest)
			return
		}

		algoConfigMu.Lock()
		AlgoConfig = newConfig
		CurrentCache = InitCache(newConfig.CacheAlgorithm, newConfig.CacheCapacity)
		MyNodeSelector = InitNodeSelector(newConfig.NodeSelectorAlgo, config.ReadConfig.SlaveNodes)
		algoConfigMu.Unlock()
		Metrics.StartNewRun(newConfig.CacheAlgorithm, newConfig.NodeSelectorAlgo, "default", 1)

		log.Printf("✅ Algorithms updated: Cache=%s, Selector=%s, Capacity=%d\n",
			newConfig.CacheAlgorithm, newConfig.NodeSelectorAlgo, newConfig.CacheCapacity)

		response := map[string]interface{}{
			"status":        "success",
			"message":       "Algorithms updated",
			"cache":         newConfig.CacheAlgorithm,
			"nodeSelector":  newConfig.NodeSelectorAlgo,
			"cacheCapacity": newConfig.CacheCapacity,
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func handlePerformanceMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests allowed", http.StatusMethodNotAllowed)
		return
	}

	current, history := Metrics.Snapshot()
	runs := make([]runSummary, 0, len(history)+1)
	runs = append(runs, history...)
	if current.RunID != "" {
		runs = append(runs, current)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "success",
		"currentRun": current,
		"runs":       runs,
	})
}

func handleMetricsReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests allowed", http.StatusMethodNotAllowed)
		return
	}

	var resetRequest struct {
		WorkloadID  string `json:"workloadId"`
		Concurrency int    `json:"concurrency"`
	}

	if err := json.NewDecoder(r.Body).Decode(&resetRequest); err != nil {
		resetRequest.WorkloadID = "default"
		resetRequest.Concurrency = 1
	}

	if resetRequest.WorkloadID == "" {
		resetRequest.WorkloadID = "default"
	}
	if resetRequest.Concurrency <= 0 {
		resetRequest.Concurrency = 1
	}

	algoConfigMu.RLock()
	configCopy := AlgoConfig
	algoConfigMu.RUnlock()

	Metrics.Reset(configCopy.CacheAlgorithm, configCopy.NodeSelectorAlgo, resetRequest.WorkloadID, resetRequest.Concurrency)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":      "success",
		"message":     "Metrics reset",
		"workloadId":  resetRequest.WorkloadID,
		"concurrency": resetRequest.Concurrency,
	})
}

// handleCacheStatus returns current cache status and stats
func handleCacheStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	algoConfigMu.RLock()
	cacheSize := CurrentCache.Size()
	configCopy := AlgoConfig
	algoConfigMu.RUnlock()

	response := map[string]interface{}{
		"status":                "success",
		"cacheAlgo":             configCopy.CacheAlgorithm,
		"cacheSize":             cacheSize,
		"cacheCapacity":         configCopy.CacheCapacity,
		"utilizationPercentage": (cacheSize * 100) / configCopy.CacheCapacity,
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// func handleFileUpdate(w http.ResponseWriter, r *http.Request) {
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

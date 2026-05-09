package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// AlgorithmSelector provides interactive CLI for choosing algorithms
type AlgorithmSelector struct {
	masterURL string
	client    *http.Client
}

// Available algorithms
type AlgorithmChoice struct {
	CacheAlgorithm   string
	NodeSelectorAlgo string
	CacheCapacity    int
}

// Response structures
type ConfigResponse struct {
	Status             string   `json:"status"`
	CurrentCache       string   `json:"currentCache"`
	CurrentSelector    string   `json:"currentSelector"`
	CacheCapacity      int      `json:"cacheCapacity"`
	AvailableCaches    []string `json:"availableCaches"`
	AvailableSelectors []string `json:"availableSelectors"`
}

type AlgorithmUpdateResponse struct {
	Status        string `json:"status"`
	Message       string `json:"message"`
	Cache         string `json:"cache"`
	NodeSelector  string `json:"nodeSelector"`
	CacheCapacity int    `json:"cacheCapacity"`
}

type CacheStatusResponse struct {
	Status                string `json:"status"`
	CacheAlgo             string `json:"cacheAlgo"`
	CacheSize             int    `json:"cacheSize"`
	CacheCapacity         int    `json:"cacheCapacity"`
	UtilizationPercentage int    `json:"utilizationPercentage"`
}

// NewAlgorithmSelector creates a new selector instance
func NewAlgorithmSelector(masterURL string) *AlgorithmSelector {
	return &AlgorithmSelector{
		masterURL: masterURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetCurrentConfig fetches current algorithm configuration
func (as *AlgorithmSelector) GetCurrentConfig() (*ConfigResponse, error) {
	resp, err := as.client.Get(as.masterURL + "/config/algorithms")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch config: %v", err)
	}
	defer resp.Body.Close()

	var config ConfigResponse
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %v", err)
	}

	return &config, nil
}

// UpdateAlgorithms sends algorithm update to master
func (as *AlgorithmSelector) UpdateAlgorithms(choice *AlgorithmChoice) (*AlgorithmUpdateResponse, error) {
	body, _ := json.Marshal(choice)

	resp, err := as.client.Post(
		as.masterURL+"/config/algorithms",
		"application/json",
		strings.NewReader(string(body)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update algorithms: %v", err)
	}
	defer resp.Body.Close()

	var result AlgorithmUpdateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return &result, nil
}

// GetCacheStatus fetches current cache status
func (as *AlgorithmSelector) GetCacheStatus() (*CacheStatusResponse, error) {
	resp, err := as.client.Get(as.masterURL + "/config/cache-status")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch cache status: %v", err)
	}
	defer resp.Body.Close()

	var status CacheStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("failed to parse status: %v", err)
	}

	return &status, nil
}

// PrintHeader prints a formatted header
func (as *AlgorithmSelector) PrintHeader() {
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("                    🚀 GoDrive Algorithm Selector 🚀")
	fmt.Println(strings.Repeat("=", 70))
}

// PrintCurrentConfig displays current algorithm configuration
func (as *AlgorithmSelector) PrintCurrentConfig(config *ConfigResponse) {
	fmt.Println("\n📊 Current Configuration:")
	fmt.Println("  ├─ Cache Algorithm    : " + config.CurrentCache)
	fmt.Println("  ├─ Node Selector      : " + config.CurrentSelector)
	fmt.Println("  └─ Cache Capacity     : " + fmt.Sprintf("%d", config.CacheCapacity))

	fmt.Println("\n📋 Available Caches:")
	for i, cache := range config.AvailableCaches {
		fmt.Printf("    %d. %s\n", i+1, cache)
	}

	fmt.Println("\n📋 Available Selectors:")
	for i, selector := range config.AvailableSelectors {
		fmt.Printf("    %d. %s\n", i+1, selector)
	}
}

// PrintCacheStatus displays cache statistics
func (as *AlgorithmSelector) PrintCacheStatus(status *CacheStatusResponse) {
	fmt.Println("\n💾 Cache Status:")
	fmt.Println("  ├─ Algorithm          : " + status.CacheAlgo)
	fmt.Printf("  ├─ Utilization        : %d / %d items (%d%%)\n",
		status.CacheSize, status.CacheCapacity, status.UtilizationPercentage)
	fmt.Println("  └─ Status             : Active")
}

// PromptForCache shows cache selection menu
func (as *AlgorithmSelector) PromptForCache(reader *bufio.Reader, available []string) (string, error) {
	fmt.Println("\n🔹 Select Cache Algorithm:")
	for i, cache := range available {
		fmt.Printf("  %d. %s\n", i+1, cache)
	}

	fmt.Print("\nEnter choice (1-" + fmt.Sprintf("%d", len(available)) + "): ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	for i, cache := range available {
		if input == fmt.Sprintf("%d", i+1) || strings.EqualFold(input, cache) {
			return cache, nil
		}
	}

	return "", fmt.Errorf("invalid cache selection: %s", input)
}

// PromptForSelector shows selector selection menu
func (as *AlgorithmSelector) PromptForSelector(reader *bufio.Reader, available []string) (string, error) {
	fmt.Println("\n🔹 Select Node Selector Algorithm:")
	for i, selector := range available {
		fmt.Printf("  %d. %s\n", i+1, selector)
	}

	fmt.Print("\nEnter choice (1-" + fmt.Sprintf("%d", len(available)) + "): ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	for i, selector := range available {
		if input == fmt.Sprintf("%d", i+1) || strings.EqualFold(input, selector) {
			return selector, nil
		}
	}

	return "", fmt.Errorf("invalid selector selection: %s", input)
}

// PromptForCapacity asks for cache capacity
func (as *AlgorithmSelector) PromptForCapacity(reader *bufio.Reader) (int, error) {
	fmt.Print("\n🔹 Enter Cache Capacity (1-10000, default 100): ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return 100, nil
	}

	var capacity int
	_, err := fmt.Sscanf(input, "%d", &capacity)
	if err != nil {
		return 0, fmt.Errorf("invalid capacity: %s", input)
	}

	if capacity < 1 || capacity > 10000 {
		return 0, fmt.Errorf("capacity must be between 1 and 10000")
	}

	return capacity, nil
}

// InteractiveMode runs the interactive selector
func (as *AlgorithmSelector) InteractiveMode() {
	as.PrintHeader()

	// Fetch current configuration
	config, err := as.GetCurrentConfig()
	if err != nil {
		log.Fatalf("❌ Failed to connect to master server: %v\n", err)
	}

	as.PrintCurrentConfig(config)

	// Get cache status
	status, err := as.GetCacheStatus()
	if err == nil {
		as.PrintCacheStatus(status)
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("\n" + strings.Repeat("-", 70))
		fmt.Println("\n🎯 What would you like to do?")
		fmt.Println("  1. Change Cache Algorithm")
		fmt.Println("  2. Change Node Selector")
		fmt.Println("  3. Change Both Algorithms")
		fmt.Println("  4. View Cache Status")
		fmt.Println("  5. View Current Config")
		fmt.Println("  6. Exit")

		fmt.Print("\nEnter choice (1-6): ")
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			as.ChangeCache(reader, config)
		case "2":
			as.ChangeSelector(reader, config)
		case "3":
			as.ChangeBoth(reader, config)
		case "4":
			status, _ := as.GetCacheStatus()
			as.PrintCacheStatus(status)
		case "5":
			config, _ := as.GetCurrentConfig()
			as.PrintCurrentConfig(config)
		case "6":
			fmt.Println("\n👋 Goodbye!\n")
			return
		default:
			fmt.Println("❌ Invalid choice. Please enter 1-6.")
		}
	}
}

// ChangeCache changes only the cache algorithm
func (as *AlgorithmSelector) ChangeCache(reader *bufio.Reader, config *ConfigResponse) {
	cache, err := as.PromptForCache(reader, config.AvailableCaches)
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		return
	}

	capacity, err := as.PromptForCapacity(reader)
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		return
	}

	fmt.Println("\n⏳ Updating algorithm...")

	result, err := as.UpdateAlgorithms(&AlgorithmChoice{
		CacheAlgorithm:   cache,
		NodeSelectorAlgo: config.CurrentSelector,
		CacheCapacity:    capacity,
	})

	if err != nil {
		fmt.Printf("❌ Failed to update: %v\n", err)
		return
	}

	if result.Status == "success" {
		fmt.Println("\n✅ SUCCESS! Algorithm updated immediately:")
		fmt.Printf("  ├─ Cache Algorithm    : %s\n", result.Cache)
		fmt.Printf("  ├─ Node Selector      : %s\n", result.NodeSelector)
		fmt.Printf("  └─ Cache Capacity     : %d\n", result.CacheCapacity)
		fmt.Println("\n🟢 New algorithm is active and ready to use!")

		// Update local config
		config.CurrentCache = cache
		config.CacheCapacity = capacity
	} else {
		fmt.Printf("❌ Update failed: %s\n", result.Message)
	}
}

// ChangeSelector changes only the node selector algorithm
func (as *AlgorithmSelector) ChangeSelector(reader *bufio.Reader, config *ConfigResponse) {
	selector, err := as.PromptForSelector(reader, config.AvailableSelectors)
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		return
	}

	fmt.Println("\n⏳ Updating algorithm...")

	result, err := as.UpdateAlgorithms(&AlgorithmChoice{
		CacheAlgorithm:   config.CurrentCache,
		NodeSelectorAlgo: selector,
		CacheCapacity:    config.CacheCapacity,
	})

	if err != nil {
		fmt.Printf("❌ Failed to update: %v\n", err)
		return
	}

	if result.Status == "success" {
		fmt.Println("\n✅ SUCCESS! Algorithm updated immediately:")
		fmt.Printf("  ├─ Cache Algorithm    : %s\n", result.Cache)
		fmt.Printf("  ├─ Node Selector      : %s\n", result.NodeSelector)
		fmt.Printf("  └─ Cache Capacity     : %d\n", result.CacheCapacity)
		fmt.Println("\n🟢 New algorithm is active and ready to use!")

		// Update local config
		config.CurrentSelector = selector
	} else {
		fmt.Printf("❌ Update failed: %s\n", result.Message)
	}
}

// ChangeBoth changes both cache and selector algorithms
func (as *AlgorithmSelector) ChangeBoth(reader *bufio.Reader, config *ConfigResponse) {
	cache, err := as.PromptForCache(reader, config.AvailableCaches)
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		return
	}

	selector, err := as.PromptForSelector(reader, config.AvailableSelectors)
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		return
	}

	capacity, err := as.PromptForCapacity(reader)
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		return
	}

	fmt.Println("\n⏳ Updating algorithms...")

	result, err := as.UpdateAlgorithms(&AlgorithmChoice{
		CacheAlgorithm:   cache,
		NodeSelectorAlgo: selector,
		CacheCapacity:    capacity,
	})

	if err != nil {
		fmt.Printf("❌ Failed to update: %v\n", err)
		return
	}

	if result.Status == "success" {
		fmt.Println("\n✅ SUCCESS! Both algorithms updated immediately:")
		fmt.Printf("  ├─ Cache Algorithm    : %s\n", result.Cache)
		fmt.Printf("  ├─ Node Selector      : %s\n", result.NodeSelector)
		fmt.Printf("  └─ Cache Capacity     : %d\n", result.CacheCapacity)
		fmt.Println("\n🟢 New algorithms are active and ready to use!")

		// Update local config
		config.CurrentCache = cache
		config.CurrentSelector = selector
		config.CacheCapacity = capacity
	} else {
		fmt.Printf("❌ Update failed: %s\n", result.Message)
	}
}

// DirectModeFromArgs allows passing algorithms as command-line arguments
func (as *AlgorithmSelector) DirectModeFromArgs(cacheAlgo, selectorAlgo string, capacity int) {
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("            🚀 Applying Algorithm Configuration 🚀")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Printf("\nConfiguration:\n")
	fmt.Printf("  ├─ Cache Algorithm    : %s\n", cacheAlgo)
	fmt.Printf("  ├─ Node Selector      : %s\n", selectorAlgo)
	fmt.Printf("  └─ Cache Capacity     : %d\n\n", capacity)

	fmt.Println("⏳ Applying configuration...")

	result, err := as.UpdateAlgorithms(&AlgorithmChoice{
		CacheAlgorithm:   cacheAlgo,
		NodeSelectorAlgo: selectorAlgo,
		CacheCapacity:    capacity,
	})

	if err != nil {
		log.Fatalf("❌ Failed to apply configuration: %v\n", err)
	}

	if result.Status == "success" {
		fmt.Println("\n✅ SUCCESS! Configuration applied immediately:")
		fmt.Printf("  ├─ Cache Algorithm    : %s\n", result.Cache)
		fmt.Printf("  ├─ Node Selector      : %s\n", result.NodeSelector)
		fmt.Printf("  └─ Cache Capacity     : %d\n", result.CacheCapacity)
		fmt.Println("\n🟢 New algorithms are active and ready to use!\n")
	} else {
		log.Fatalf("❌ Configuration failed: %s\n", result.Message)
	}
}

// main function for standalone execution
func runAlgorithmSelector(masterURL string) {
	if masterURL == "" {
		masterURL = "http://127.0.0.1:4000"
	}

	selector := NewAlgorithmSelector(masterURL)

	// Check if arguments provided for direct mode
	if len(os.Args) > 1 && os.Args[1] != "" {
		// Direct mode: go run main.go select <cache> <selector> <capacity>
		// Example: go run main.go select lru powerOfTwo 100
		if len(os.Args) >= 4 && os.Args[1] == "select" {
			var capacity int
			_, err := fmt.Sscanf(os.Args[4], "%d", &capacity)
			if err != nil {
				capacity = 100
			}
			selector.DirectModeFromArgs(os.Args[2], os.Args[3], capacity)
			return
		}
	}

	// Interactive mode
	selector.InteractiveMode()
}

// Note: If you want to use this in the main program, add this to main.go:
//
// import "os"
//
// func init() {
//     // Add this flag to main.go if you want algorithm selector as a feature
//     if len(os.Args) > 1 && os.Args[1] == "algo-select" {
//         runAlgorithmSelector("http://127.0.0.1:4000")
//         os.Exit(0)
//     }
// }

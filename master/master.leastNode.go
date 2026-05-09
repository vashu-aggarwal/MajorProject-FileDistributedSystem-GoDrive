package master

import (
	"godrive/config"
	"log"
	"math"
	"sync"
)

type LeastNodeSelector struct {
	mu    sync.Mutex
	nodes []config.Node
}

func NewLeastNodeSelector(nodes []config.Node) *LeastNodeSelector {
	log.Println("[LeastNodeSelector] Initialized with nodes:", nodes)
	return &LeastNodeSelector{
		nodes: nodes,
	}
}

func (L *LeastNodeSelector) GiveNode() config.Node {
	L.mu.Lock()
	defer L.mu.Unlock()

	minLoad := math.MaxInt
	var selectedNode config.Node

	log.Println("[LeastNodeSelector] Calculating node loads...")

	for _, node := range L.nodes {
		load := getNodeLoad(node.Port)

		log.Printf(
			"[LeastNodeSelector] Node %s:%s → load=%d\n",
			node.Host, node.Port, load,
		)

		if load < minLoad {
			minLoad = load
			selectedNode = node
		}
	}

	log.Printf(
		"[LeastNodeSelector] Selected node %s:%s with load=%d\n",
		selectedNode.Host, selectedNode.Port, minLoad,
	)

	return selectedNode
}

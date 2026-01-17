package master

import (
	"godrive/config"
	"log"
	"sync"
)

/* ---------- ROUND ROBIN ---------- */

type RoundRobinNodeSelector struct {
	mutex     sync.Mutex
	nodeIndex int
	nodeList  []config.Node
}

func NewRoundRobinSelector(nodes []config.Node) *RoundRobinNodeSelector {
	log.Println("🎯 Node Selection Algorithm: ROUND ROBIN")
	return &RoundRobinNodeSelector{
		nodeIndex: 0,
		nodeList:  nodes,
	}
}

func (r *RoundRobinNodeSelector) GiveNode() config.Node {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	node := r.nodeList[r.nodeIndex]
	r.nodeIndex = (r.nodeIndex + 1) % len(r.nodeList)

	// Log which node was selected
	log.Printf("📦 RoundRobin selected node: %s:%s\n", node.Host, node.Port)

	return node
}

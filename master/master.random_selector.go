package master

import (
	"log"
	"math/rand"
	"time"

	"godrive/config"
)

type RandomNodeSelector struct {
	nodes []config.Node
}

func NewRandomNodeSelector(nodes []config.Node) NodeSelector {
	rand.Seed(time.Now().UnixNano())

	log.Println("🔀 Random Node Selector initialized")

	return &RandomNodeSelector{
		nodes: nodes,
	}
}

func (r *RandomNodeSelector) GiveNode() config.Node {
	if len(r.nodes) == 0 {
		log.Println("❌ RandomNodeSelector: No nodes available")
		return config.Node{}
	}

	index := rand.Intn(len(r.nodes))
	selected := r.nodes[index]

	log.Printf("🎯 RandomNodeSelector selected node: %s:%s\n",
		selected.Host,
		selected.Port,
	)

	return selected
}

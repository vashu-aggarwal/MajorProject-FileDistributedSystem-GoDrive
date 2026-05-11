package master

import (
	"godrive/config"
	"log"
	"math/rand"
	"sync"
	"time"
)

type PowerOfTwoSelector struct {
	nodes []config.Node
	mu    sync.Mutex
}

func NewPowerOfTwoSelector(nodes []config.Node) *PowerOfTwoSelector {
	rand.Seed(time.Now().UnixNano())
	log.Println("🔁 Power-of-Two Node Selector ENABLED")
	return &PowerOfTwoSelector{
		nodes: nodes,
	}
}

func (p *PowerOfTwoSelector) GiveNode() config.Node {
	p.mu.Lock()
	defer p.mu.Unlock()

	n1 := p.nodes[rand.Intn(len(p.nodes))]
	n2 := p.nodes[rand.Intn(len(p.nodes))]

	load1 := getNodeLoad(n1.Port)
	load2 := getNodeLoad(n2.Port)

	log.Printf(
		"[POWER-OF-TWO] Candidates: %s(load=%d), %s(load=%d)",
		n1.Port, load1,
		n2.Port, load2,
	)

	if load1 <= load2 {
		log.Printf("[POWER-OF-TWO] Selected: %s", n1.Port)
		return n1
	}

	log.Printf("[POWER-OF-TWO] Selected: %s", n2.Port)
	return n2
}

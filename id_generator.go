package ws

import (
	"sync"
)

type IDGenerator struct {
	mu      sync.Mutex
	current uint
}

func (g *IDGenerator) Next() uint {
	g.mu.Lock()
	defer func() {
		g.current++
		g.mu.Unlock()
	}()

	return g.current
}

var globalIDGenerator = new(IDGenerator)

func NextGlobalID() uint {
	return globalIDGenerator.Next()
}

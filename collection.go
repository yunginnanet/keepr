package main

import (
	"fmt"
	"sync"
)

// Collection contains taxonomy information and relationship mapping for our Sample collectiion.
type Collection struct {
	Tempos map[int][]*Sample
	mu     *sync.RWMutex
}

var Library = &Collection{
	Tempos: make(map[int][]*Sample),
	mu:     &sync.RWMutex{},
}

// IngestTempo creates a map of tempo to sample.
func (c *Collection) IngestTempo(sample *Sample) {
	if !(sample.Tempo > 0) {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Tempos[sample.Tempo] = append(c.Tempos[sample.Tempo], sample)
}

// TempoStats outputs the amount of samples with each known tempo.
func (c *Collection) TempoStats() {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for t, ss := range c.Tempos {
		if len(ss) > 1 {
			fmt.Printf("%dBPM: %d\n", t, len(ss))
		}
	}
}

package main

import (
	"errors"
	"os"
	"strconv"
	"sync"
	"time"
)

// SampleType represents the type of sample we think it is.
type SampleType uint8

const (
	Unknown SampleType = iota
	Percussion
	Ambient
	Melodic
	DrumLoop
	Loop
	Kick
	Snare
	HatClosed
	HatOpen
)

// Sample represents an audio sample and contains relevant information regarding said sample.
type Sample struct {
	Name     string
	Path     string
	Modified time.Time
	Key      string
	Tempo    int
	Type     []SampleType
}

// TODO: make a "Collector" interface

// Collection contains taxonomy information and relationship mapping for our Sample collectiion.
type Collection struct {
	Tempos map[int][]*Sample
	mu     *sync.RWMutex
}

// Library is a global default instance of a Collection.
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
			log.Printf("%dBPM: %d", t, len(ss))
		}
	}
}

func (c *Collection) TempoSymlinks() (err error) {
	log.Trace().Msg("TempoSymlinks start")
	defer log.Trace().Err(err).Msg("TempoSymlinks finish")
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.Tempos) < 1 {
		return errors.New("no tempos recorded")
	}

	dst := apath(destination + "Tempo")
	err = os.MkdirAll(dst, os.ModePerm)
	if err != nil && !os.IsNotExist(err) {
		return
	}

	for t, ss := range c.Tempos {
		tempopath := dst + "/" + strconv.Itoa(t) + "/"
		err = os.MkdirAll(tempopath, os.ModePerm)
		if err != nil && !os.IsNotExist(err) {
			return
		}
		for _, sample := range ss {
			finalPath := tempopath + sample.Name
			log.Trace().Str("caller", sample.Path).Msg(finalPath)
			err = freshLink(finalPath)
			if err != nil {
				return
			}
			if _, err = os.Stat(sample.Path); err != nil {
				return
			}
			err = os.Symlink(sample.Path, finalPath)
			if err != nil && !os.IsNotExist(err) {
				return
			}
		}
	}
	return nil
}

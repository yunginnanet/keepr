package main

import (
	"errors"
	"os"
	"strconv"
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
	_, err = os.Stat(dst)

	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return
	}

	err = os.MkdirAll(dst, os.ModePerm)
	if err != nil {
		return
	}

	for t, ss := range c.Tempos {
		tempopath := dst + "/" + strconv.Itoa(t) + "/"

		if err != nil {

		}
		_, err = os.Stat(tempopath)
		if !errors.Is(err, os.ErrNotExist) {
			return err
		} else {
			// os.MkdirAll(tempopath+strconv.Itoa(t), os.ModePerm)
		}
		for _, sample := range ss {
			log.Debug().Str("caller", sample.FullPath()).Msg("to exist in " + tempopath)
		}
	}
	return nil
}

package collect

import "sync/atomic"

var Backlog int32

// IngestKey creates a map of tempo to sample.
func (c *Collection) IngestKey(sample *Sample) {
	atomic.AddInt32(&Backlog, 1)
	defer atomic.AddInt32(&Backlog, -1)
	log.Debug().Str("caller", sample.Name).Msgf("Key: %s", sample.Key.Root.String(sample.Key.AdjSymbol))
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Keys[sample.Key] = append(c.Keys[sample.Key], sample)
}

// IngestTempo creates a map of tempo to sample.
func (c *Collection) IngestTempo(sample *Sample) {
	if sample.Tempo == 0 || sample.Tempo < 50 || sample.Tempo > 250 {
		return
	}
	atomic.AddInt32(&Backlog, 1)
	defer atomic.AddInt32(&Backlog, -1)
	c.mu.Lock()
	defer c.mu.Unlock()
	log.Debug().Str("caller", sample.Name).Msgf("Tempo: %d", sample.Tempo)
	c.Tempos[sample.Tempo] = append(c.Tempos[sample.Tempo], sample)
}

// IngestMelodicLoop appends to a list of [pointers to] melodic loop samples.
func (c *Collection) IngestMelodicLoop(sample *Sample) {
	atomic.AddInt32(&Backlog, 1)
	defer atomic.AddInt32(&Backlog, -1)
	log.Debug().Str("caller", sample.Name).Msg("Melodic Loop")
	c.mu.Lock()
	defer c.mu.Unlock()
	c.MelodicLoops = append(c.MelodicLoops, sample)
}

// IngestMIDI appends to a list of [pointers to] TypeMIDI/SMF files.
func (c *Collection) IngestMIDI(sample *Sample) {
	atomic.AddInt32(&Backlog, 1)
	defer atomic.AddInt32(&Backlog, -1)
	log.Debug().Str("caller", sample.Name).Msg("MIDI")
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Midis = append(c.Midis, sample)
}

// IngestDrum creates a map of different drum types to samples.
func (c *Collection) IngestDrum(sample *Sample, drumType DrumType) {
	atomic.AddInt32(&Backlog, 1)
	defer atomic.AddInt32(&Backlog, -1)
	log.Debug().Str("caller", sample.Name).Msgf("Drum: %s", drumToDirMap[drumType])
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Drums[drumType] = append(c.Drums[drumType], sample)
}

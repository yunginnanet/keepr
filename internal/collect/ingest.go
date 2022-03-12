package collect

// IngestKey creates a map of tempo to sample.
func (c *Collection) IngestKey(sample *Sample) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Keys[sample.Key] = append(c.Keys[sample.Key], sample)
}

// IngestTempo creates a map of tempo to sample.
func (c *Collection) IngestTempo(sample *Sample) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Tempos[sample.Tempo] = append(c.Tempos[sample.Tempo], sample)
}

// IngestMelodicLoop appends to a list of [pointers to] melodic loop samples.
func (c *Collection) IngestMelodicLoop(sample *Sample) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.MelodicLoops = append(c.MelodicLoops, sample)
}

// IngestMIDI appends to a list of [pointers to] MIDI/SMF files.
func (c *Collection) IngestMIDI(sample *Sample) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Midis = append(c.Midis, sample)
}

// IngestDrum creates a map of different drum types to samples.
func (c *Collection) IngestDrum(sample *Sample, drumType DrumType) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Drums[drumType] = append(c.Drums[drumType], sample)
}

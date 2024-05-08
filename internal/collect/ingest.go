package collect

import (
	"strings"
	"sync/atomic"

	"github.com/davecgh/go-spew/spew"
)

var Backlog int32

// IngestKey creates a map of tempo to sample.
func (c *Collection) IngestKey(sample *Sample) {
	if sample.Key.Root == 0 {
		return
	}
	atomic.AddInt32(&Backlog, 1)
	defer atomic.AddInt32(&Backlog, -1)
	log.Debug().Str("caller", sample.Name).Msgf("Key: %s", sample.Key.Root.String(sample.Key.AdjSymbol)+modeStr(sample.Key))
	c.mu.Lock()
	c.Keys[sample.Key] = append(c.Keys[sample.Key], sample)
	c.mu.Unlock()
}

// IngestTempo creates a map of tempo to sample.
func (c *Collection) IngestTempo(sample *Sample) {
	if sample.Tempo == 0 || sample.Tempo < 50 || sample.Tempo > 250 {
		return
	}
	atomic.AddInt32(&Backlog, 1)
	defer atomic.AddInt32(&Backlog, -1)
	c.mu.Lock()
	log.Debug().Str("caller", sample.Name).Msgf("Tempo: %d", sample.Tempo)
	c.Tempos[sample.Tempo] = append(c.Tempos[sample.Tempo], sample)
	c.mu.Unlock()
}

// IngestMelodicLoop appends to a list of [pointers to] melodic loop samples.
func (c *Collection) IngestMelodicLoop(sample *Sample) {
	if !sample.IsType(TypeMelodic) || !sample.IsType(TypeLoop) {
		if sample.IsType(TypeMelodic) || sample.IsType(TypeLoop) {
			log.Warn().Str("caller", sample.Name).Msg(spew.Sdump(sample))
		}
		return
	}
	atomic.AddInt32(&Backlog, 1)
	defer atomic.AddInt32(&Backlog, -1)
	log.Debug().Str("caller", sample.Name).Msg("Melodic Loop")
	c.mu.Lock()
	c.MelodicLoops = append(c.MelodicLoops, sample)
	c.mu.Unlock()
}

// IngestMIDI appends to a list of [pointers to] TypeMIDI/SMF files.
func (c *Collection) IngestMIDI(sample *Sample) {
	if !sample.IsType(TypeMIDI) {
		return
	}
	atomic.AddInt32(&Backlog, 1)
	defer atomic.AddInt32(&Backlog, -1)
	log.Debug().Str("caller", sample.Name).Msg("MIDI")
	c.mu.Lock()
	c.MIDIs = append(c.MIDIs, sample)
	c.mu.Unlock()
}

// IngestDrum creates a map of different drum types to samples.
func (c *Collection) IngestDrum(sample *Sample, drumType DrumType) {
	if !sample.IsType(TypeDrum) && !sample.IsType(TypeDrumLoop) {
		return
	}
	atomic.AddInt32(&Backlog, 1)
	defer atomic.AddInt32(&Backlog, -1)
	log.Debug().Str("caller", sample.Name).Msgf("Drum: %s", drumToDirMap[drumType])
	c.mu.Lock()
	c.Drums[drumType] = append(c.Drums[drumType], sample)
	c.mu.Unlock()
}

func (c *Collection) IngestArtist(sample *Sample) {
	if sample.Metadata == nil || sample.Metadata.Artist == "" {
		return
	}
	atomic.AddInt32(&Backlog, 1)
	defer atomic.AddInt32(&Backlog, -1)
	log.Debug().Str("caller", sample.Name).Msgf("Artist: %s", sample.Metadata.Artist)
	c.mu.Lock()
	c.Artists[sample.Metadata.Artist] = append(c.Artists[sample.Metadata.Artist], sample)
	c.mu.Unlock()
}

func (c *Collection) IngestGenre(sample *Sample) {
	if sample.Metadata == nil || sample.Metadata.Genre == "" {
		return
	}
	atomic.AddInt32(&Backlog, 1)
	defer atomic.AddInt32(&Backlog, -1)
	log.Debug().Str("caller", sample.Name).Msgf("Genre: %s", sample.Metadata.Genre)
	c.mu.Lock()
	c.Genres[sample.Metadata.Genre] = append(c.Genres[sample.Metadata.Genre], sample)
	c.mu.Unlock()
}

func (c *Collection) IngestSource(sample *Sample) {
	if sample.Metadata == nil || sample.Metadata.Source == "" {
		return
	}
	atomic.AddInt32(&Backlog, 1)
	defer atomic.AddInt32(&Backlog, -1)
	log.Debug().Str("caller", sample.Name).Msgf("Source: %s", sample.Metadata.Source)
	c.mu.Lock()
	c.Sources[sample.Metadata.Source] = append(c.Sources[sample.Metadata.Source], sample)
	c.mu.Unlock()
}

func (c *Collection) IngestCreationDate(sample *Sample) {
	if sample.Metadata == nil || sample.Metadata.CreationDate == "" {
		return
	}
	atomic.AddInt32(&Backlog, 1)
	defer atomic.AddInt32(&Backlog, -1)
	log.Debug().Str("caller", sample.Name).Msgf("Creation Date: %s", sample.Metadata.CreationDate)
	c.mu.Lock()
	c.CreationDates[sample.Metadata.CreationDate] = append(c.CreationDates[sample.Metadata.CreationDate], sample)
	c.mu.Unlock()
}

func (c *Collection) IngestSoftware(sample *Sample) {
	if sample.Metadata == nil || sample.Metadata.Software == "" {
		return
	}
	atomic.AddInt32(&Backlog, 1)
	defer atomic.AddInt32(&Backlog, -1)
	log.Debug().Str("caller", sample.Name).Msgf("Software: %s", sample.Metadata.Software)
	c.mu.Lock()
	c.Software[sample.Metadata.Software] = append(c.Software[sample.Metadata.Software], sample)
	c.mu.Unlock()
}

func (c *Collection) IngestMetadata(sample *Sample) {
	if sample.Metadata == nil {
		return
	}
	c.IngestArtist(sample)
	c.IngestGenre(sample)
	c.IngestSource(sample)
	c.IngestCreationDate(sample)
	c.IngestSoftware(sample)
}

var replacers = []string{
	" ", "-", ".", "/", "\\",
	":", ",", ";", "!", "?",
	"'", "(", ")", "[", "]",
	"{", "}", "<", ">", "|",
	"&", "^", "%", "$", "#",
	"@", "!", "~", "`", "+",
}

func (c *Collection) IngestSample(sample *Sample) {
	c.IngestMetadata(sample)
	c.IngestKey(sample)
	c.IngestTempo(sample)
	// c.IngestDrum(sample, sample.DrumType) // happens during filename parsing
	c.IngestMelodicLoop(sample)
	c.IngestMIDI(sample)
	c.DeDupe()
}

func (c *Collection) DeDupe() {
	sampMaps := []map[string][]*Sample{c.Artists, c.Genres, c.Sources, c.CreationDates}
	for _, sampMap := range sampMaps {
		dupes := map[string]map[string][]*Sample{}
		for title, values := range sampMap {
			ogTitle := title
			title = strings.ToLower(title)
			title = strings.TrimSpace(title)
			for _, replacer := range replacers {
				title = strings.ReplaceAll(title, replacer, "_")
			}
			if _, ok := dupes[title]; !ok {
				dupes[title] = make(map[string][]*Sample)
				dupes[title][ogTitle] = values
				continue
			}
			if len(values) == 0 {
				delete(sampMap, ogTitle)
				continue
			}
			dupes[title][ogTitle] = values
		}
		for dedupedTitle, titles := range dupes {
			if len(titles) == 1 {
				continue
			}
			log.Debug().Msgf("deduping %s", dedupedTitle)
			for ogTitle, values := range titles {
				sampMap[dedupedTitle] = append(sampMap[dedupedTitle], values...)
				delete(sampMap, ogTitle)
			}
			var paths = map[string]struct{}{}
			var newVals []*Sample
			for _, smp := range sampMap[dedupedTitle] {
				if _, ok := paths[smp.Path]; ok {
					continue
				}
				paths[smp.Path] = struct{}{}
				newVals = append(newVals, smp)
			}
			sampMap[dedupedTitle] = newVals
			log.Trace().Msgf("deduped %s", dedupedTitle)
		}
	}
}

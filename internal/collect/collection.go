package collect

import (
	"errors"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/go-audio/wav"
	"github.com/rs/zerolog"
	"gopkg.in/music-theory.v0/key"

	"git.tcp.direct/kayos/keepr/internal/config"
	"git.tcp.direct/kayos/keepr/internal/util"
)

var log *zerolog.Logger

func init() {
	go func() {
		for log == nil {
			log = config.GetLogger()
		}
	}()
}

// SampleType represents the type of sample we think it is.
type SampleType uint8

//goland:noinspection GoUnusedConst
const (
	Unknown SampleType = iota
	Ambient
	Melodic
	DrumLoop
	OneShot
	Drum
	Loop
	MIDI
)

type DrumType uint8

const (
	Kick DrumType = iota
	Snare
	HiHat
	HatClosed
	HatOpen
	Tom
	Percussion
	EightOhEight
)

// Sample represents an audio sample and contains relevant information regarding said sample.
type Sample struct {
	Name     string
	Path     string
	ModTime  time.Time
	Duration time.Duration
	Key      key.Key
	Tempo    int
	Type     []SampleType
	Metadata *wav.Metadata
}

// TODO: make a "Collector" interface

// Collection contains taxonomy information and relationship mapping for our Sample collectiion.
type Collection struct {
	Tempos       map[int][]*Sample
	Keys         map[key.Key][]*Sample
	Drums        map[DrumType][]*Sample
	DrumLoops    []*Sample
	MelodicLoops []*Sample
	Midis        []*Sample
	mu           *sync.RWMutex
}

// Library is a global default instance of a Collection.
var Library = &Collection{
	Tempos: make(map[int][]*Sample),
	Keys:   make(map[key.Key][]*Sample),
	Drums:  make(map[DrumType][]*Sample),
	mu:     &sync.RWMutex{},
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

func (c *Collection) SymlinkTempos() (err error) {
	log.Trace().Msg("SymlinkTempos start")
	defer log.Trace().Err(err).Msg("SymlinkTempos finish")
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.Tempos) < 1 {
		return errors.New("no tempos recorded")
	}

	dst := util.APath(config.Destination+"Tempo", config.Relative)
	err = os.MkdirAll(dst, os.ModePerm)
	if err != nil && !os.IsNotExist(err) {
		return
	}

	for t, ss := range c.Tempos {
		tempopath := dst + "/" + strconv.Itoa(t) + "/"
		err = os.MkdirAll(tempopath, os.ModePerm)
		if err != nil && !os.IsExist(err) {
			return
		}
		for _, s := range ss {
			go func(sample *Sample) {
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
					log.Error().Err(err).Msg("failed to create symlink")
				}
			}(s)
		}
	}
	return nil
}

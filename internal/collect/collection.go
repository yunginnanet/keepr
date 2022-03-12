package collect

import (
	"errors"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
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

// DrumStats outputs the amount of samples of each known drum type.
func (c *Collection) DrumStats() {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for t, ss := range c.Drums {
		if len(ss) > 1 {
			log.Printf("%s: %d", drumToDirMap[t], len(ss))
		}
	}
}

// KeyStats outputs the amount of samples with each known key.
func (c *Collection) KeyStats() {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for t, ss := range c.Keys {
		if len(ss) > 1 {
			log.Printf("%s: %d", t.Root.String(t.AdjSymbol), len(ss))
		}
	}
}

func link(sample *Sample, kp string) {
	atomic.AddInt32(&Backlog, 1)
	defer atomic.AddInt32(&Backlog, -1)
	slog := log.With().Str("caller", sample.Path).Logger()
	finalPath := kp + sample.Name
	slog.Trace().Msg(finalPath)
	err := freshLink(finalPath)
	if err != nil && !os.IsNotExist(err) {
		slog.Trace().Msgf(err.Error())
	}
	if _, err = os.Stat(sample.Path); err != nil {
		slog.Warn().Err(err).Msg("can't stat original file")
	}
	if config.Simulate {
		log.Printf("would have linked %s -> %s", sample.Path, finalPath)
		return
	}
	symerr := os.Symlink(sample.Path, finalPath)
	if symerr != nil && !os.IsExist(symerr) && !os.IsNotExist(symerr) {
		slog.Error().Err(symerr).Msg("failed to create symlink")
	}
}

func (c *Collection) SymlinkTempos() (err error) {
	atomic.AddInt32(&Backlog, 1)
	defer atomic.AddInt32(&Backlog, -1)
	log.Trace().Msg("SymlinkTempos start")
	defer log.Trace().Err(err).Msg("SymlinkTempos finish")
	c.mu.RLock()
	defer c.mu.RUnlock()
	if len(c.Tempos) < 1 {
		return errors.New("no known tempos")
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
			go link(s, tempopath)
		}
	}
	return nil
}

func (c *Collection) SymlinkKeys() (err error) {
	atomic.AddInt32(&Backlog, 1)
	defer atomic.AddInt32(&Backlog, -1)
	log.Trace().Msg("SymlinkKeys start")
	defer log.Trace().Err(err).Msg("SymlinkKeys finish")
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.Keys) < 1 {
		return errors.New("no known keys")
	}
	dst := util.APath(config.Destination+"Key", config.Relative)
	err = os.MkdirAll(dst, os.ModePerm)
	if err != nil && !os.IsNotExist(err) {
		return
	}
	for t, ss := range c.Keys {
		keypath := dst + "/" + t.Root.String(t.AdjSymbol) + "/"
		err = os.MkdirAll(keypath, os.ModePerm)
		if err != nil && !os.IsExist(err) {
			return
		}
		for _, s := range ss {
			go link(s, keypath)
		}
	}
	return nil
}

func (c *Collection) SymlinkDrums() (err error) {
	atomic.AddInt32(&Backlog, 1)
	defer atomic.AddInt32(&Backlog, -1)
	log.Trace().Msg("SymlinkDrums start")
	defer log.Trace().Err(err).Msg("SymlinkDrums finish")
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.Drums) < 1 {
		return errors.New("no known drums")
	}
	dst := util.APath(config.Destination+"Drums", config.Relative)
	err = os.MkdirAll(dst, os.ModePerm)
	if err != nil && !os.IsNotExist(err) {
		return
	}
	for t, ss := range c.Drums {
		drumpath := dst + "/" + drumToDirMap[t] + "/"
		err = os.MkdirAll(drumpath, os.ModePerm)
		if err != nil && !os.IsExist(err) {
			return
		}
		for _, s := range ss {
			go link(s, drumpath)
		}
	}
	return nil
}

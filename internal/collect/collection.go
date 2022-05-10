package collect

import (
	"errors"
	"fmt"
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
	TypeUnknown SampleType = iota
	TypeAmbient
	TypeMelodic
	TypeDrumLoop
	TypeOneShot
	TypeDrum
	TypeLoop
	TypeMIDI
)

type DrumType uint8

const (
	DrumKick DrumType = iota
	DrumSnare
	DrumHiHat
	DrumHatClosed
	DrumHatOpen
	DrumTom
	DrumPercussion
	Drum808
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

func (c *Collection) TypeStats() {
	println(fmt.Sprintf("Drum loops: %d", len(c.DrumLoops)))
	println(fmt.Sprintf("Melodic loops: %d", len(c.MelodicLoops)))
}

// TempoStats outputs the amount of samples with each known tempo.
func (c *Collection) TempoStats() {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for t, ss := range c.Tempos {
		if len(ss) > 1 {
			println(fmt.Sprintf("%dBPM: %d", t, len(ss)))
		}
	}
}

// DrumStats outputs the amount of samples of each known drum type.
func (c *Collection) DrumStats() {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for t, ss := range c.Drums {
		if len(ss) > 1 {
			println(fmt.Sprintf("%s: %d", drumToDirMap[t], len(ss)))
		}
	}
}

// KeyStats outputs the amount of samples with each known key.
func (c *Collection) KeyStats() {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for t, ss := range c.Keys {
		if len(ss) > 1 {
			println(fmt.Sprintf("%s: %d", t.Root.String(t.AdjSymbol), len(ss)))
		}
	}
}

func link(sample *Sample, kp string) {
	atomic.AddInt32(&Backlog, 1)
	defer atomic.AddInt32(&Backlog, -1)

	mapMu.RLock()
	if _, ok := lockMap[sample.Path]; !ok {
		mapMu.RUnlock()
		mapMu.Lock()
		lockMap[sample.Path] = &sync.Mutex{}
		mapMu.Unlock()
		mapMu.RLock()
	}
	defer mapMu.RUnlock()

	lockMap[sample.Path].Lock()
	defer lockMap[sample.Path].Unlock()

	slog := log.With().Str("caller", sample.Path).Logger()
	finalPath := kp + sample.Name
	slog.Trace().Msg(finalPath)
	err := freshLink(finalPath)
	if err != nil && !os.IsNotExist(err) {
		slog.Warn().Msgf(err.Error())
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

func (c *Collection) SymlinkMelodicLoops() (err error) {
	atomic.AddInt32(&Backlog, 1)
	defer atomic.AddInt32(&Backlog, -1)
	log.Trace().Msg("SymlinkMelodicLoops start")
	defer log.Trace().Err(err).Msg("SymlinkMelodicLoops finish")
	c.mu.RLock()
	defer c.mu.RUnlock()
	if len(c.MelodicLoops) < 1 {
		return errors.New("no known tempos")
	}
	dest := config.Output + "Melodic_Loops"
	dst := util.APath(dest, config.Relative)
	err = os.MkdirAll(dst, os.ModePerm)
	mlpath := dst + "/"
	if err != nil && !os.IsNotExist(err) {
		return
	}
	for _, s := range c.MelodicLoops {
		var oneshot = false
		for _, t := range s.Type {
			if t == TypeOneShot {
				break
			}
		}
		if !oneshot {
			go link(s, mlpath)
		}
	}
	return nil
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
	dst := util.APath(config.Output+"Tempo", config.Relative)
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
	dst := util.APath(config.Output+"Key", config.Relative)
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
	dst := util.APath(config.Output+"Drums", config.Relative)
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

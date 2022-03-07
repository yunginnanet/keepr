package main

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"kr.dev/walk"
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
	Name      string
	Directory string
	Modified  time.Time
	Key       string
	Tempo     int
	Type      []SampleType
}

func (s *Sample) FullPath() string {
	return basepath + s.Directory + "/" + s.Name
}

func apath(path string) string {
	if relative {
		return path
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to get absolute path")
	}
	return abs
}

func main() {
	var lastpath = ""
	cripwalk := walk.New(os.DirFS("/"), target)
	for cripwalk.Next() {
		if err := cripwalk.Err(); err != nil {
			log.Warn().Str("caller", lastpath).Msg(err.Error())
		}
		lastpath = cripwalk.Path()
		slog := log.With().Str("caller", cripwalk.Path()).Logger()
		switch {
		case cripwalk.Entry() == nil:
			slog.Trace().Msg("nil")
			continue
		case cripwalk.Entry().IsDir():
			if strings.Contains(cripwalk.Path(), destination) {
				slog.Debug().Msg("skiping directory entirely")
				cripwalk.SkipDir()
			}
			slog.Trace().Msg("directory")
		case cripwalk.Path() == os.Args[1]:
			slog.Debug().Msg("skiping samplesimp directory entirely")
			cripwalk.SkipParent()
		default:
			sample := process(cripwalk.Entry(), apath(cripwalk.Path()))
			Library.IngestTempo(sample)
		}
	}
	Library.TempoStats()
	err := Library.TempoSymlinks()
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}
}

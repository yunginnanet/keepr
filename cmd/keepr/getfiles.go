package main

import (
	"os"
	"runtime"
	"strings"

	"github.com/rs/zerolog"
	"kr.dev/walk"

	"git.tcp.direct/kayos/keepr/internal/collect"
	"git.tcp.direct/kayos/keepr/internal/config"
	"git.tcp.direct/kayos/keepr/internal/util"
)

var (
	log      *zerolog.Logger
	basepath = "/"
)

func init() {
	if runtime.GOOS == "windows" {
		// TODO: fix this garbage
		basepath = "C:\\"
	}
}

func main() {
	log = config.GetLogger()
	var lastpath = ""
	cripwalk := walk.New(os.DirFS(basepath), config.Target)
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
			if strings.Contains(cripwalk.Path(), config.Destination) {
				slog.Debug().Msg("skiping directory entirely")
				cripwalk.SkipDir()
			}
			slog.Trace().Msg("directory")
		case cripwalk.Path() == os.Args[1]:
			slog.Debug().Msg("skipping self-parent directory entirely")
			cripwalk.SkipParent()
		default:
			sample, err := collect.Process(cripwalk.Entry(), util.APath(cripwalk.Path(), config.Relative))
			if err != nil {
				log.Warn().Err(err).Msgf("failed to process %s", cripwalk.Entry().Name())
				continue
			}
			if sample == nil {
				log.Trace().Msgf("skipping unknown file %s", cripwalk.Entry().Name())
				continue
			}
			collect.Library.IngestTempo(sample)
		}
	}

	if zerolog.GlobalLevel() == zerolog.TraceLevel {
		collect.Library.TempoStats()
	}

	if config.StatsOnly {
		return
	}

	err := collect.Library.SymlinkTempos()
	if err != nil {
		log.Fatal().Err(err).Msg("returned from symlinkTempos")
	}
}

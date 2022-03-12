package main

import (
	"os"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

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
			// slog.Trace().Msg("directory")
		case cripwalk.Path() == os.Args[1]:
			slog.Debug().Msg("skipping self-parent directory entirely")
			cripwalk.SkipParent()
		default:
			sample, err := collect.Process(cripwalk.Entry(), util.APath(cripwalk.Path(), config.Relative))
			if err != nil {
				slog.Warn().Err(err).Msgf("failed to process")
				continue
			}
			if sample == nil {
				slog.Trace().Msgf("skipping unknown file")
				continue
			}
		}
	}

	for !atomic.CompareAndSwapInt32(&collect.Backlog, 0, -1) {
		time.Sleep(1 * time.Second)
		print(".")
	}

	if config.StatsOnly {
		collect.Library.TempoStats()
		collect.Library.KeyStats()
		collect.Library.DrumStats()
	}

	var errs []error
	errs = append(errs, collect.Library.SymlinkTempos())
	errs = append(errs, collect.Library.SymlinkKeys())
	errs = append(errs, collect.Library.SymlinkDrums())

	log.Info().Errs("errs", errs).Msg("fin.")
}

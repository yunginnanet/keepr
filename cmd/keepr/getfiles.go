package main

import (
	"os"
	"path/filepath"
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
	target := strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(config.Source), "/"), "/")
	cripwalk := walk.New(os.DirFS(basepath), target)
	_, output := filepath.Split(strings.TrimSuffix(config.Output, "/"))
	log.Trace().Msgf("output is %s", output)
	for cripwalk.Next() {
		if err := cripwalk.Err(); err != nil {
			log.Fatal().Caller().Str("caller", lastpath).Msg(err.Error())
			continue
		}
		lastpath = cripwalk.Path()
		slog := log.With().Str("caller", cripwalk.Path()).Logger()
		switch {
		case cripwalk.Entry() == nil:
			slog.Trace().Msg("nil")
			continue
		case cripwalk.Entry().IsDir():
			if strings.Contains(cripwalk.Path(), output) {
				slog.Info().Msg("skiping directory entirely")
				cripwalk.SkipDir()
			}
			continue
			// slog.Trace().Msg("directory")
		default:
			if strings.Contains(cripwalk.Path(), config.Output) {
				log.Trace().Msg("skipping file in destination")
				cripwalk.SkipDir()
				continue
			}
			sample, err := collect.Process(cripwalk.Entry(), util.APath(cripwalk.Path(), config.Relative))
			if err != nil {
				slog.Warn().Caller().Str("caller", cripwalk.Path()).Err(err).Msgf("failed to process")
				continue
			}
			if sample == nil {
				slog.Trace().Msgf("skipping unknown file")
				continue
			}
			slog.Info().Interface("sample", sample).Msg("processed")
		}
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

	for !atomic.CompareAndSwapInt32(&collect.Backlog, 0, -1) {
		time.Sleep(1 * time.Second)
		print(".")
	}

	log.Info().Errs("errs", errs).Msg("fin.")
}

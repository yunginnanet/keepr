package config

import (
	"bytes"
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"

	"git.tcp.direct/kayos/keepr/internal/art"
)

const defDestination = "001-LINKED_SORTED_DIRECTORIES"

var (
	log    zerolog.Logger
	Source = ""
	// Output is the base path for our symlink library.
	Output = defDestination
	// Relative will determine if we use relative pathing for symlinks.
	Relative      = false
	Simulate      = false
	StatsOnly     = false
	NoMIDI        = false
	SkipWavDecode = false
)

// GetLogger retrieves a pointer to our zerolog instance.
func GetLogger() *zerolog.Logger {
	return &log
}

func required(i int) {
	if !(len(os.Args) < i) {
		return
	}
	println("invalid syntax, missing argument")
	os.Exit(1)
}

var helpStr = `
           Flags:

--output, -o     (required) output directory
--source, -s     (required) source directory

--debug, -v      enable debug output
--trace, -vv     enable trace output
--relative, -r   enable relative pathing
--stats          only output stats, no symlinking
--no-op, -n      simulate actions only, change nothing (read only)
--no-midi, -m    do not parse MIDI files
--fast, -f       do not parse WAV files

--help, -h       it me

`

func help() {
	rdr := bytes.NewReader([]byte(helpStr))
	io.Copy(os.Stdout, rdr)
}

func KeeprInit() {
	println(art.String())

	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	log = zerolog.New(zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
		w.TimeFormat = time.RFC822
	})).With().Timestamp().Logger()

	for i, arg := range os.Args {
		if i == 0 {
			continue
		}
		switch arg {
		case "_":
			continue
		case "-o", "--output":
			required(i)
			Output = os.Args[i+1]
			os.Args[i+1] = "_"
		case "--debug", "-v":
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		case "--trace", "-vv":
			zerolog.SetGlobalLevel(zerolog.TraceLevel)
		case "--relative", "-r":
			Relative = true
		case "--stats":
			StatsOnly = true
		case "--no-op", "-n":
			Simulate = true
		case "--no-midi", "-m":
			NoMIDI = true
		case "--fast", "-f":
			SkipWavDecode = true
		case "--source", "-s":
			required(i)
			Source = os.Args[i+1]
			os.Args[i+1] = "_"
		case "--help", "-h":
			help()
			os.Exit(0)
		default:
			log.Fatal().Msg("unknown argument: " + arg)
			help()
		}
	}

	if Source == "" {
		log.Error().Msg("missing target search directory")
		help()
		os.Exit(1)
	}

	f, err := os.Stat(Output)
	switch {
	case err != nil:
		if !os.IsNotExist(err) {
			log.Fatal().Caller().Str("caller", Output).Err(err).Msg("")
		}
		if err := os.MkdirAll(Output, os.ModePerm); err != nil {
			log.Fatal().Caller().Str("caller", Output).Err(err).Msg("could not make directory")
		}
	case !f.IsDir():
		if Output != "./"+defDestination {
			log.Error().Caller().Str("caller", Output).Msg("not a directory")
			help()
			os.Exit(1)
		}
		if err := os.MkdirAll(Output, os.ModePerm); err != nil {
			log.Fatal().Caller().Str("caller", Output).Err(err).Msg("could not make directory")
		}
	}
	if !strings.HasSuffix(Output, "/") {
		Output = Output + "/"
	}
}

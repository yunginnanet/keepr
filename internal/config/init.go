package config

import (
	"os"
	"strings"
	"time"

	"git.tcp.direct/kayos/common/squish"
	"github.com/rs/zerolog"
)

const defDestination = "001-LINKED_SORTED_DIRECTORIES"

var (
	log    zerolog.Logger
	Target = ""
	// Destination is the base path for our symlink library.
	Destination = defDestination
	// Relative will determine if we use relative pathing for symlinks.
	Relative      = false
	Simulate      = false
	StatsOnly     = false
	NoMIDI        = false
	SkipWavDecode = false
)

const (
	banner  = "H4sIAAAAAAACA7VWz2qDMBy++wq95NBrtVuLZWzsRSpI6aSUrmuxMtjwEGRHD85m4mHvsfsexSdZqo35XzvFECHB7/t+f6MBbqdhWA7oMhzDcUMABvO72f10vC1QXGQfeGdXO4iX02qZDOaTEnMBXaLG29MjgCopioNEjbMByexfhtA1YqJeeyXRrVIJgNDAD5vKqAZmR0I3KTlSpV/F+f3hSJLXkQKYHU/REM/FLOZ9qlBuGcKZFUvFpPlNNKajZgOsf0x9rhFmK2ZfyEFaICT2AvEm52CxkBNbcjmt5PjuonLMiQQ9C6nj1OlBIdQspZWN/m9gxjZCLkYlVYcYKrJPPJmeks50hRB8Or//Ji9jxrIyS20JkUSAisPDtshE7mpWIeG3uS4xeDkul18JnrpPzBVoOQc6MOLjY0qa9CdzTaDlaMdl3OKOUs0T/U40BUGKhqGKSP0voU2SK9xHQkvpPnmwX3renOiG4BU/UL6vR4+rdWAGy735tPa9ZWBtFm+7g7XxvL0PRuB1eGMOb/HisNjunz2w81eLl/W75z+MagnQ+vYVusbpDuVaHW5wlvEHTKOp9AMKAAA="
	version = "0.1"
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

func init() {
	c := squish.UnpackStr(banner)
	c = strings.ReplaceAll(c, "$1", strings.Split(version, ".")[0]) + "."
	println(strings.ReplaceAll(c, "$2", strings.Split(version, ".")[1]))

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
		case "--destination", "-d":
			required(i)
			Destination = os.Args[i+1]
			os.Args[i+1] = "_"
		case "--debug", "-v":
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		case "--trace", "-vv":
			zerolog.SetGlobalLevel(zerolog.TraceLevel)
		case "--relative", "-r":
			Relative = true
		case "--stats", "-s":
			StatsOnly = true
		case "--no-op", "-n":
			Simulate = true
		case "--no-midi", "-m":
			NoMIDI = true
		case "--fast", "-f":
			SkipWavDecode = true
		default:
			Target = strings.Trim(arg, "/")
		}
	}

	if Target == "" {
		log.Fatal().Msg("missing target search directory")
	}

	f, err := os.Stat(Destination)
	switch {
	case err != nil:
		if !os.IsNotExist(err) {
			log.Fatal().Caller().Str("caller", Destination).Err(err).Msg("")
		}
		if err := os.MkdirAll(Destination, os.ModePerm); err != nil {
			log.Fatal().Caller().Str("caller", Destination).Err(err).Msg("could not make directory")
		}
	case !f.IsDir():
		if Destination != "./"+defDestination {
			log.Fatal().Caller().Str("caller", Destination).Msg("not a directory")
		}
		if err := os.MkdirAll(Destination, os.ModePerm); err != nil {
			log.Fatal().Caller().Str("caller", Destination).Err(err).Msg("could not make directory")
		}
	}
	if !strings.HasSuffix(Destination, "/") {
		Destination = Destination + "/"
	}
}

package config

import (
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

const defDestination = "001-LINKED_SORTED_DIRECTORIES"

var (
	log    zerolog.Logger
	Target = ""
	// Destination is the base path for our symlink library.
	Destination = defDestination
	// Relative will determine if we use relative pathing for symlinks.
	Relative = false
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
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
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
		default:
			Target = strings.Trim(arg, "/")
			println("search target detected: " + Target)
		}
	}

	log = zerolog.New(zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
		w.TimeFormat = time.RFC822
	})).With().Timestamp().Logger()

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

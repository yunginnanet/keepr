package main

import (
	"errors"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

const defDestination = "001-LINKED_SORTED_DIRECTORIES"

var (
	log         zerolog.Logger
	target      = ""
	destination = "./" + defDestination
	basepath    = "/"
	relative    = false
)

func required(i int) {
	if !(len(os.Args) < i) {
		return
	}
	println("invalid syntax, missing argument")
	os.Exit(1)
}

func init() {
	if runtime.GOOS == "windows" {
		basepath = "C:/"
	}
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
			destination = os.Args[i+1]
			os.Args[i+1] = "_"
		case "--debug", "-v":
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		case "--trace", "-vv":
			zerolog.SetGlobalLevel(zerolog.TraceLevel)
		case "--relative", "-r":
			relative = true
		default:
			target = strings.Trim(arg, "/")
			println("search target detected: " + target)
		}
	}

	log = zerolog.New(zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
		w.TimeFormat = time.RFC822
	})).With().Timestamp().Logger()

	if target == "" {
		log.Fatal().Msg("missing target search directory")
	}

	f, err := os.Stat(destination)
	switch {
	case err != nil:
		if !errors.Is(err, os.ErrNotExist) {
			log.Fatal().Caller().Str("caller", destination).Err(err).Msg("")
		}
		if err := os.MkdirAll(destination, os.ModePerm); err != nil {
			log.Fatal().Caller().Str("caller", destination).Err(err).Msg("could not make directory")
		}
	case !f.IsDir():
		if destination != "./"+defDestination {
			log.Fatal().Caller().Str("caller", destination).Msg("not a directory")
		}
		if err := os.MkdirAll(destination, os.ModePerm); err != nil {
			log.Fatal().Caller().Str("caller", destination).Err(err).Msg("could not make directory")
		}
	}
	if !strings.HasSuffix(destination, "/") {
		destination = destination + "/"
	}
}

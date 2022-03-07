package main

import (
	"io/fs"
	"strconv"
	"strings"
)

func checkbpm(piece string) (bpm int, ok bool) {
	frg := strings.Split(piece, "bpm")[0]
	m := strings.Split(frg, "")
	var start = 0
	var numfound = false
	for b, p := range m {
		if _, e := strconv.Atoi(p); e != nil {
			start = b
			continue
		}
		numfound = true
		break
	}
	if !numfound {
		return 0, false
	}
	var fail error
	if bpm, fail = strconv.Atoi(frg[start:]); fail == nil {
		return bpm, true
	}
	return 0, false
}

func process(entry fs.DirEntry, dir string) *Sample {
	finfo, err := entry.Info()
	if err != nil {
		log.Warn().Err(err).Msg(entry.Name())
	}
	s := &Sample{
		Name:      entry.Name(),
		Directory: dir,
		Modified:  finfo.ModTime(),
	}
	name := strings.ToLower(entry.Name())

	var (
		spl  []string
		sep  string = " "
		seps        = []string{"_", "-", " - "}
	)

	for _, s := range seps {
		if strings.Contains(name, s) {
			sep = s
		}
	}

	spl = strings.Split(name, sep)
	for _, piece := range spl {
		switch {
		case strings.Contains(piece, "bpm"):
			bpm, ok := checkbpm(piece)
			if ok {
				s.Tempo = bpm
			}
		}
	}
	return s
}

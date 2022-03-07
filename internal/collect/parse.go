package collect

import (
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"strings"

	"github.com/go-audio/wav"
)

func freshLink(path string) error {
	if _, err := os.Lstat(path); err == nil {
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("failed to unlink: %+v", err)
		}
	}
	return nil
}

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

func guessBPM(name string) int {
	name = strings.ToLower(name)
	var (
		spl  []string
		sep  = " "
		seps = []string{"_", "-", " - "}
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
				return bpm
			}
		}
	}
	return 0
}

func readWAV(s *Sample) error {
	f, err := os.Open(s.Path)
	if err != nil {
		return fmt.Errorf("couldn't open %s: %s", s.Path, err.Error())
	}
	defer f.Close()

	decoder := wav.NewDecoder(f)
	s.Duration, err = decoder.Duration()
	if err != nil {
		return fmt.Errorf("failed to get duration for %s: %s", s.Name, err.Error())
	}
	decoder.ReadMetadata()
	if decoder.Err() != nil {
		return decoder.Err()
	}
	s.Metadata = decoder.Metadata

	return nil
}

func Process(entry fs.DirEntry, dir string) (s *Sample, err error) {
	var finfo os.FileInfo
	finfo, err = entry.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to Process %s: %s", entry.Name(), err.Error())
	}

	s = &Sample{
		Name:    entry.Name(),
		Path:    dir,
		ModTime: finfo.ModTime(),
	}
	err = readWAV(s)
	s.Tempo = guessBPM(s.Name)

	return
}

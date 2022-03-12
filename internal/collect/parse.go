package collect

import (
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"strings"

	"github.com/go-audio/wav"
	"gopkg.in/music-theory.v0/key"
	"gopkg.in/music-theory.v0/note"
)

func freshLink(path string) error {
	if _, err := os.Lstat(path); err == nil {
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("failed to unlink: %+v", err)
		}
	}
	return nil
}

func checkbpm(piece string) (bpm int) {
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
		return 0
	}
	var fail error
	if bpm, fail = strconv.Atoi(frg[start:]); fail == nil {
		return bpm
	}
	return 0
}

func guessSeperator(name string) (spl []string) {
	var (
		sep  = " "
		seps = []string{"-", "_", " - "}
	)
	for _, s := range seps {
		if strings.Contains(name, s) {
			sep = s
		}
	}
	return strings.Split(name, sep)
}

func (s *Sample) getParentDir() string {
	spl := strings.Split(s.Path, "/")
	return strings.ToLower(spl[len(spl)-2])
}

func (s *Sample) IsType(st SampleType) bool {
	for _, t := range s.Type {
		if t == st {
			return true
		}
	}
	return false
}

var drumDirMap = map[string]DrumType{
	"snares": Snare, "kicks": Kick, "hats": HiHat, "hihats": HiHat, "closed_hihats": HatClosed,
	"open_hihats": HatOpen, "808s": EightOhEight, "808": EightOhEight, "toms": Tom,
}

func (s *Sample) ParseFilename() {
	for _, piece := range guessSeperator(s.Name) {
		piece = strings.ToLower(piece)
		if strings.Contains(piece, "bpm") {
			s.Tempo = checkbpm(piece)
			if s.Tempo != 0 {
				s.Type = append(s.Type, Loop)
			}
		}

		drumtype, isdrum := drumDirMap[s.getParentDir()]

		switch {
		case s.getParentDir() == "melodic_loops":
			if !s.IsType(Loop) {
				s.Type = append(s.Type, Loop)
				go Library.IngestMelodicLoop(s)
			}
			if isdrum {
				go Library.IngestDrum(s, drumtype)
			}
		}

		spl := strings.Split(piece, "")

		// if our fragment starts with a known root note, then try to parse the fragment, else dip-set.
		switch spl[0] {
		case "C", "D", "E", "F", "G", "A", "B":
			break
		default:
			continue
		}

		k := key.Of(piece)
		if k.Root != note.Nil {
			s.Key = k
			Library.IngestKey(s)
		}
	}
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
	log.Trace().Msg(fmt.Sprintf("metadata: %v", s.Metadata))

	decoder.ReadInfo()

	return nil
}

func Process(entry fs.DirEntry, dir string) (s *Sample, err error) {
	var finfo os.FileInfo
	finfo, err = entry.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to Process %s: %s", entry.Name(), err.Error())
	}

	spl := strings.Split(entry.Name(), ".")
	ext := spl[len(spl)-1]

	s = &Sample{
		Name:    entry.Name(),
		Path:    dir,
		ModTime: finfo.ModTime(),
	}

	switch ext {
	case "midi", "mid":
		s.Type = append(s.Type, MIDI)
	case "wav":
		err = readWAV(s)
		if err != nil {
			return nil, err
		}
		s.ParseFilename()
	default:
		return nil, nil
	}

	return
}

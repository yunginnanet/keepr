package collect

import (
	"fmt"
	"io/fs"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/go-audio/wav"
	"gopkg.in/music-theory.v0/key"

	"git.tcp.direct/kayos/keepr/internal/config"
)

func freshLink(path string) error {
	if _, err := os.Lstat(path); err == nil {
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("failed to unlink: %+v", err)
		}
	}
	return nil
}

func guessBPM(piece string) (bpm int) {
	// TODO: don't trust this lol?
	if num, numerr := strconv.Atoi(piece); numerr != nil {
		return num
	}
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

var drumToDirMap = map[DrumType]string{
	Snare: "Snares", Kick: "Kicks", HiHat: "HiHats", HatClosed: "HiHat/Closed",
	HatOpen: "HiHat/Open", EightOhEight: "808", Tom: "Toms", Percussion: "Other",
}

var (
	rgxSharpIn, _    = regexp.Compile("[♯#]|major")
	rgxFlatIn, _     = regexp.Compile("^F|[♭b]")
	rgxSharpBegin, _ = regexp.Compile("^[♯#]")
	rgxFlatBegin, _  = regexp.Compile("^[♭b]")
	rgxSharpishIn, _ = regexp.Compile("(M|maj|major|aug)")
	rgxFlattishIn, _ = regexp.Compile("([^a-z]|^)(m|min|minor|dim)")
	mustMatchOne     = []*regexp.Regexp{rgxFlatIn, rgxFlatBegin, rgxSharpBegin, rgxSharpIn, rgxFlattishIn, rgxSharpishIn}
)

func (s *Sample) ParseFilename() {
	atomic.AddInt32(&Backlog, 1)
	defer atomic.AddInt32(&Backlog, -1)
	drumtype, isdrum := drumDirMap[s.getParentDir()]

	switch {
	case s.getParentDir() == "melodic_loops":
		if !s.IsType(Loop) {
			s.Type = append(s.Type, Loop)
			go Library.IngestMelodicLoop(s)
		}
	case isdrum:
		go Library.IngestDrum(s, drumtype)
	}

	for _, opiece := range guessSeperator(s.Name) {
		piece := strings.ToLower(opiece)
		if num, numerr := strconv.Atoi(piece); numerr == nil {
			if num > 50 && num != 808 {
				s.Tempo = num
			}
		}
		if strings.Contains(piece, "bpm") {
			s.Tempo = guessBPM(piece)
		}

		if s.Tempo != 0 {
			go Library.IngestTempo(s)
		}

		spl := strings.Split(opiece, "")
		if len(spl) < 1 || len(spl) > 5 {
			continue
		}
		// if our fragment starts with a known root note, then try to parse the fragment, else dip-set.
		switch spl[0] {
		case "C", "D", "E", "F", "G", "A", "B":
			if len(spl) < 2 {
				break
			}
			var found = false
			for _, rgx := range mustMatchOne {
				if rgx.MatchString(opiece) {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		default:
			continue
		}

		s.Key = key.Of(opiece)
		go Library.IngestKey(s)
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
	if meta := decoder.Metadata; meta == nil {
		return nil
	}

	s.Metadata = decoder.Metadata
	log.Trace().Msg(fmt.Sprintf("metadata: %v", s.Metadata))

	decoder.ReadInfo()

	return nil
}

func Process(entry fs.DirEntry, dir string) (s *Sample, err error) {
	log.Trace().Str("caller", entry.Name()).Msg("Processing")
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
		if !config.NoMIDI {
			s.Type = append(s.Type, MIDI)
			go Library.IngestMIDI(s)
		}
	case "wav":
		if !config.SkipWavDecode {
			err = readWAV(s)
		}
		if err != nil {
			return nil, err
		}
		go s.ParseFilename()
	default:
		return nil, nil
	}

	return
}

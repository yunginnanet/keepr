package collect

import (
	"fmt"
	"io/fs"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-audio/wav"
	"gopkg.in/music-theory.v0/key"

	"git.tcp.direct/kayos/keepr/internal/config"
)

var lockMap = make(map[string]*sync.Mutex)

var mapMu = &sync.RWMutex{}

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
	log.Trace().Msgf("found seperator for %s: %s", name, sep)
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
	"snares": DrumSnare, "kicks": DrumKick, "hats": DrumHiHat, "hihats": DrumHiHat, "closed_hihats": DrumHatClosed,
	"open_hihats": DrumHatOpen, "808s": Drum808, "808": Drum808, "toms": DrumTom,
}

var drumToDirMap = map[DrumType]string{
	DrumSnare: "Snares", DrumKick: "Kicks", DrumHiHat: "HiHats", DrumHatClosed: "HiHats/Closed",
	DrumHatOpen: "HiHats/Open", Drum808: "808", DrumTom: "Toms", DrumPercussion: "Other",
}

var (
	rgxSharpIn, _    = regexp.Compile("[♯#]|major")
	rgxFlatIn, _     = regexp.Compile("^F|[♭b]")
	rgxSharpBegin, _ = regexp.Compile("^[♯#]")
	rgxFlatBegin, _  = regexp.Compile("^[♭b]")
	rgxSharpishIn, _ = regexp.Compile("(maj|major|aug)")
	rgxFlattishIn, _ = regexp.Compile("([^a-z]|^)(m|min|minor|dim)")

	mustMatchOne = map[string]*regexp.Regexp{
		"flat": rgxFlatIn, "flat_begin": rgxFlatBegin,
		"sharp_begin": rgxSharpBegin, "sharp": rgxSharpIn,
		"flattish": rgxFlattishIn, "sparpish": rgxSharpishIn}
)

func (s *Sample) ParseFilename() {
	atomic.AddInt32(&Backlog, 1)
	defer atomic.AddInt32(&Backlog, -1)
	drumtype, isdrum := drumDirMap[s.getParentDir()]

	switch {
	case s.getParentDir() == "melodic_loops", strings.Contains(s.getParentDir(), "melod"):
		if !s.IsType(TypeLoop) {
			s.Type = append(s.Type, TypeLoop)
			go Library.IngestMelodicLoop(s)
		}
	case isdrum:
		go Library.IngestDrum(s, drumtype)
	}

	var fallback = ""
	var keyFound = false

	roots := []string{"C", "D", "E", "F", "G", "A", "B"}

	opieces := guessSeperator(s.Name)
	for _, opiece := range opieces {
		opiece = strings.TrimSuffix(opiece, ".wav")
		for _, r := range roots {
			if strings.TrimSpace(opiece) == r {
				fallback = opiece
			}
		}
	}

	for _, opiece := range opieces {
		opiece = strings.TrimSuffix(opiece, ".wav")
		log.Trace().Msgf("parse %s, piece: %s", s.Name, opiece)
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
		if len(spl) > 6 || len(spl) == 0 {
			continue
		}

		// if our fragment starts with a known root note, then try to parse the fragment, else dip-set.
		switch spl[0] {
		case "C", "D", "E", "F", "G", "A", "B":
			var found = false
			for desc, rgx := range mustMatchOne {
				if rgx.MatchString(opiece) {
					log.Trace().Msgf("matched regex for %s", desc)
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
		if s.Key.Root != 0 {
			keyFound = true
			go Library.IngestKey(s)
		}
	}
	if !keyFound && fallback != "" {
		log.Warn().Msgf("using fallback key for %s: %s", s.Name, fallback)
		s.Key = key.Of(fallback)
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

	decoder.ReadMetadata()
	if decoder.Err() != nil {
		return decoder.Err()
	}
	if s.Metadata = decoder.Metadata; s.Metadata == nil {
		return fmt.Errorf("%s had nil metadata despite no error", s.Name)
	}

	s.Duration, err = decoder.Duration()
	if err != nil {
		return fmt.Errorf("failed to get duration for %s: %s", s.Name, err.Error())
	}

	if s.Duration != 0 && s.Duration < 2*time.Second {
		s.Type = append(s.Type, TypeOneShot)
		var newTypes []SampleType
		for _, t := range s.Type {
			if t != TypeLoop {
				newTypes = append(newTypes, t)
			}
		}
		s.Type = newTypes
	}

	log.Trace().Msg(fmt.Sprintf("metadata: %v", s.Metadata))

	//

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
			s.Type = append(s.Type, TypeMIDI)
			go Library.IngestMIDI(s)
		}
	case "wav":
		if !config.SkipWavDecode {
			wavErr := readWAV(s)
			if wavErr != nil {
				log.Debug().Caller().Str("caller", s.Name).Msgf("failed to parse wav data: %s", wavErr.Error())
			}
		}
		s.ParseFilename()
	default:
		return nil, nil
	}

	return
}

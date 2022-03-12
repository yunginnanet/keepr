package collect

import (
	"testing"
	"time"

	"github.com/go-audio/wav"
	"github.com/go-music-theory/music-theory/key"
)

func TestSample_getParentDir(t *testing.T) {
	type fields struct {
		Name     string
		Path     string
		ModTime  time.Time
		Duration time.Duration
		Key      key.Key
		Tempo    int
		Type     []SampleType
		Metadata *wav.Metadata
	}
	type test struct {
		name   string
		fields fields
		want   string
	}
	var tests = []test{test{
		name: "testparentdir",
		fields: fields{
			Name: "OS_BB_808_E_RARRI.wav",
			Path: "/home/fuckholejones/808_Bass/OS_BB_808_E_RARRI.wav",
		},
		want: "808_bass",
	},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Sample{
				Name: tt.fields.Name,
				Path: tt.fields.Path,
			}
			if got := s.getParentDir(); got != tt.want {
				t.Errorf("getParentDir() = %v, want %v", got, tt.want)
			} else {
				t.Logf("parent directory: " + got)
			}
		})
	}
}

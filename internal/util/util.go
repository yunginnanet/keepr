package util

import (
	"path/filepath"
	"runtime"

	"github.com/rs/zerolog/log"
)

// APath is a wrapper for filepath.Abs for convenience and flexibility.
func APath(path string, relative bool) string {
	basepath := "/"
	// TODO: make this more gooder
	if runtime.GOOS == "windows" {
		basepath = "c:\\"
	}
	if relative {
		return path
	}
	abs, err := filepath.Abs(basepath + path)
	if err != nil {
		log.Warn().Caller(1).Str("caller", path).Err(err).Msg("unable to get absolute path")
		return path
	}
	return abs
}

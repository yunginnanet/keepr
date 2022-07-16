package util

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
	"runtime"
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

func FreshLink(path string) error {
	if _, err := os.Lstat(path); err == nil {
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("failed to unlink: %+v", err)
		}
	}
	return nil
}

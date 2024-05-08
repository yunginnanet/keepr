package art

import (
	"os"
	"testing"
)

func TestArt(t *testing.T) {
	_, err := os.Stdout.WriteString(String())
	if err != nil {
		t.Error(err)
	}
}

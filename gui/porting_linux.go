// +build gui

package gui
import (
	"path/filepath"
	"os"
)

func getUserPath() string {
	home := os.Getenv("HOME")
	if home == "" {
		panic("Required environment variable HOME is not set")
	}
	return filepath.Join(home, ".kidscode")
}

// +build gui

package gui
import (
	"path/filepath"
	"os"
)

func getUserPath() string {
	appdata := os.Getenv("APPDATA")
	if appdata == "" {
		panic("Required environment variable HOME is not set")
	}
	return filepath.Join(appdata, "KidscodeIGN")
}

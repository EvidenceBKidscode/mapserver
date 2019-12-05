package app

import (
	"encoding/json"
	"io/ioutil"
)

type RasterOverlay struct {
	Label        string        `json:"label"`
	Texture      string        `json:"texture"`
}

func RasterOverlaysParse(filename string) ([]RasterOverlay, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	overlays := make([]RasterOverlay,0)
	err = json.Unmarshal(data, &overlays)
	if err != nil {
		return nil, err
	}

	return overlays, nil
}

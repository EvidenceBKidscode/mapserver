package geometry

import (
	"encoding/json"
	"io/ioutil"
)

func Parse(filename string) (*Geometry, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	geometry := Geometry{}

	err = json.Unmarshal(data, &geometry)
	if err != nil {
		return nil, err
	}

	return &geometry, nil
}

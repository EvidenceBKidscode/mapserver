package web

import (
	"encoding/json"
	"mapserver/app"
	"mapserver/layer"
	"mapserver/geometry"
	"net/http"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"path"
)

//Public facing config
type PublicConfig struct {
	Version         string               `json:"version"`
	Layers          []*layer.Layer       `json:"layers"`
	MapObjects      *app.MapObjectConfig `json:"mapobjects"`
	DefaultOverlays []string             `json:"defaultoverlays"`
	EnableSearch    bool                 `json:"enablesearch"`

	WorldName       string               `json:"worldname"`
	WorldId         string               `json:"worldid"`
	Geometry        *geometry.Geometry   `json:"geometry"`
}

type ConfigHandler struct {
	ctx *app.App
}

func (h *ConfigHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Add("content-type", "application/json")

	webcfg := PublicConfig{}
	webcfg.Layers = h.ctx.Config.Layers
	webcfg.MapObjects = h.ctx.Config.MapObjects
	webcfg.Version = app.Version
	webcfg.DefaultOverlays = h.ctx.Config.DefaultOverlays
	webcfg.EnableSearch = h.ctx.Config.EnableSearch

	webcfg.WorldName = path.Base(h.ctx.WorldDir)

	webcfg.Geometry = h.ctx.Geometry

	// Create a hash ID from map coordinates and map name
	// TODO: Find a better solution (file stored ID ? what should ID represent?)
	// Id is used to locally distinguish different maps
	hasher := sha1.New()
	hasher.Write([]byte(webcfg.WorldName))

	if (webcfg.Geometry != nil) {
		hasher.Write([]byte(fmt.Sprintf("%f:", webcfg.Geometry.CoordinatesCarto[0][0])))
		hasher.Write([]byte(fmt.Sprintf("%f:", webcfg.Geometry.CoordinatesCarto[0][1])))
		hasher.Write([]byte(fmt.Sprintf("%f:", webcfg.Geometry.CoordinatesCarto[1][0])))
		hasher.Write([]byte(fmt.Sprintf("%f:", webcfg.Geometry.CoordinatesCarto[1][1])))
		hasher.Write([]byte(fmt.Sprintf("%f:", webcfg.Geometry.CoordinatesCarto[2][0])))
		hasher.Write([]byte(fmt.Sprintf("%f:", webcfg.Geometry.CoordinatesCarto[2][1])))
		hasher.Write([]byte(fmt.Sprintf("%f:", webcfg.Geometry.CoordinatesCarto[3][0])))
		hasher.Write([]byte(fmt.Sprintf("%f:", webcfg.Geometry.CoordinatesCarto[3][1])))
	}
	webcfg.WorldId = base64.URLEncoding.EncodeToString(hasher.Sum(nil))

	json.NewEncoder(resp).Encode(webcfg)
}

package app

import (
	"mapserver/blockaccessor"
	"mapserver/colormapping"
	"mapserver/db"
	"mapserver/eventbus"
	"mapserver/mapblockaccessor"
	"mapserver/mapblockrenderer"
	"mapserver/mapobjectdb"
	"mapserver/params"
	"mapserver/settings"
	"mapserver/tiledb"
	"mapserver/tilerenderer"
	"path/filepath"
)

type App struct {
	Params      params.ParamsType
	Config      *Config
	WorldDir    string
	Worldconfig map[string]string

	Blockdb  db.DBAccessor
	Objectdb mapobjectdb.DBAccessor
	TileDB   *tiledb.TileDB
	Settings settings.Settings

	MapBlockAccessor *mapblockaccessor.MapBlockAccessor
	BlockAccessor    *blockaccessor.BlockAccessor
	Colormapping     *colormapping.ColorMapping
	Mapblockrenderer *mapblockrenderer.MapBlockRenderer
	Tilerenderer     *tilerenderer.TileRenderer

	MediaRepo map[string][]byte

	WebEventbus *eventbus.Eventbus
}

func (ctx *App) GetWorldPath(filename string) string {
	return filepath.Join(ctx.WorldDir, filename)
}

package app

import (
	"mapserver/blockaccessor"
	"mapserver/colormapping"
	"mapserver/db/postgres"
	"mapserver/db/sqlite"
	"mapserver/eventbus"
	"mapserver/mapblockaccessor"
	"mapserver/mapblockrenderer"
	postgresobjdb "mapserver/mapobjectdb/postgres"
	sqliteobjdb "mapserver/mapobjectdb/sqlite"
	"mapserver/media"
	"mapserver/params"
	"mapserver/settings"
	"mapserver/tiledb"
	"mapserver/tilerenderer"
	"mapserver/worldconfig"
	"mapserver/geometry"
	"time"

	"io/ioutil"
	"os"

	"github.com/sirupsen/logrus"

	"errors"
)

func Setup(p params.ParamsType, cfg *Config, worlddir string) *App {
	a := App{}
	a.Params = p
	a.Config = cfg
	a.WorldDir = worlddir
	a.WebEventbus = eventbus.New()

	var err error

	//Parse world config
	a.Worldconfig = worldconfig.Parse(a.GetWorldPath("world.mt"))

	//Parse geometry data
	a.Geometry, err = geometry.Parse(a.GetWorldPath("geometry.dat"))
	if (err != nil) {
		logrus.WithFields(logrus.Fields{"error": err}).Warn("No geometry data for this world")
	}

	logrus.WithFields(logrus.Fields{"version": Version}).Info("Starting mapserver")

	switch a.Worldconfig[worldconfig.CONFIG_BACKEND] {
	case worldconfig.BACKEND_SQLITE3:
		a.Blockdb, err = sqlite.New(a.GetWorldPath("map.sqlite"))
		if err != nil {
			panic(err)
		}

	case worldconfig.BACKEND_POSTGRES:
		a.Blockdb, err = postgres.New(a.Worldconfig[worldconfig.CONFIG_PSQL_CONNECTION])
		if err != nil {
			panic(err)
		}

	default:
		panic(errors.New("map-backend not supported: " + a.Worldconfig[worldconfig.CONFIG_BACKEND]))
	}

	//migrate block db
	err = a.Blockdb.Migrate()
	if err != nil {
		panic(err)
	}

	//mapblock accessor
	expireDuration, err := time.ParseDuration(cfg.MapBlockAccessorCfg.Expiretime)
	if err != nil {
		panic(err)
	}

	purgeDuration, err := time.ParseDuration(cfg.MapBlockAccessorCfg.Purgetime)
	if err != nil {
		panic(err)
	}

	// mapblock accessor
	a.MapBlockAccessor = mapblockaccessor.NewMapBlockAccessor(
		a.Blockdb,
		expireDuration, purgeDuration,
		cfg.MapBlockAccessorCfg.MaxItems)

	// block accessor
	a.BlockAccessor = blockaccessor.New(a.MapBlockAccessor)

	//color mapping
	a.Colormapping = colormapping.NewColorMapping()

	colorfiles := []string{
		//https://daconcepts.com/vanessa/hobbies/minetest/colors.txt
		//"/colors/vanessa.txt",
		//"/colors/advtrains.txt",
		//"/colors/scifi_nodes.txt",
		//"/colors/custom.txt",
		"/colors/kidscode.txt",
	}

	for _, colorfile := range colorfiles {
		_, err := a.Colormapping.LoadVFSColors(false, colorfile)
		if err != nil {
			panic(err.Error() + " file:" + colorfile)
		}
	}

	//load provided colors, if available
	info, err := os.Stat(a.GetWorldPath("colors.txt"))
	if info != nil && err == nil {
		logrus.WithFields(logrus.Fields{"filename": "colors.txt"}).Info("Loading colors from filesystem")

		data, err := ioutil.ReadFile(a.GetWorldPath("colors.txt"))
		if err != nil {
			panic(err)
		}

		count, err := a.Colormapping.LoadBytes(data)
		if err != nil {
			panic(err)
		}

		logrus.WithFields(logrus.Fields{"count": count}).Info("Loaded custom colors")

	}

	//mapblock renderer
	a.Mapblockrenderer = mapblockrenderer.NewMapBlockRenderer(a.MapBlockAccessor, a.Colormapping)

	//mapserver database
	if a.Worldconfig[worldconfig.CONFIG_PSQL_MAPSERVER] != "" {
		a.Objectdb, err = postgresobjdb.New(a.Worldconfig[worldconfig.CONFIG_PSQL_MAPSERVER])
	} else {
		a.Objectdb, err = sqliteobjdb.New(a.GetWorldPath("mapserver.sqlite"))
	}

	if err != nil {
		panic(err)
	}

	//migrate object database
	err = a.Objectdb.Migrate()

	if err != nil {
		panic(err)
	}

	//create tiledb
	a.TileDB, err = tiledb.New(a.GetWorldPath("mapserver.tiles"))

	if err != nil {
		panic(err)
	}

	//settings
	a.Settings = settings.New(a.Objectdb)

	//setup tile renderer
	a.Tilerenderer = tilerenderer.NewTileRenderer(
		a.Mapblockrenderer,
		a.TileDB,
		a.Blockdb,
		a.Config.Layers,
	)

	//create media repo
	repo := make(map[string][]byte)

	if a.Config.EnableMediaRepository {
		mediasize, _ := media.ScanDir(repo, a.WorldDir, []string{"mapserver.tiles", ".git"})
		fields := logrus.Fields{
			"count": len(repo),
			"bytes": mediasize,
		}
		logrus.WithFields(fields).Info("Created media repository")
	}

	a.MediaRepo = repo

	return &a
}

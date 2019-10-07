package tilerendererjob

import (
	"mapserver/app"
	"mapserver/coords"
	"mapserver/mapblockparser"
	"github.com/sirupsen/logrus"
)

const (
	MAP_SIZE = 65536 / 16
)

var renderedSectors map[int]bool

func clearRenderedSectors() {
	renderedSectors = nil
}

func markSectorRendered(X int, Y int) {
	if renderedSectors == nil {
		renderedSectors = make(map[int]bool)
	}
	renderedSectors[Y*MAP_SIZE+X] = true
}

func isRendered(X int, Y int) (bool) {
	return renderedSectors[Y*MAP_SIZE+X]
}

func renderMapblocks(ctx *app.App, mblist []*mapblockparser.MapBlock) int {
	tilecount := 0
	totalRenderedMapblocks.Add(float64(len(mblist)))

	for _, mb := range mblist {
		if isRendered(mb.Pos.X, mb.Pos.Z) {
			continue
		}
		markSectorRendered(mb.Pos.X, mb.Pos.Z)
		tc := coords.GetTileCoordsFromMapBlock(mb.Pos, ctx.Config.Layers)
		ctx.TileDB.MarkOutdated(tc)
	}

	for z := coords.MAX_ZOOM; z >= coords.MIN_ZOOM; z-- {
		//Spin up workers
		jobs := make(chan coords.TileCoords, ctx.Config.RenderingQueue)
		done := make(chan bool, 1)

		for j := 0; j < ctx.Config.RenderingJobs; j++ {
			go worker(ctx, jobs, done)
		}

		for _, tc := range ctx.TileDB.GetOutdatedByZoom(z) {
			fields := logrus.Fields{
				"pos":    tc,
				"prefix": "tilerenderjob",
			}
			logrus.WithFields(fields).Debug("Tile render job mapblock")

			tilecount++

			fields = logrus.Fields{
				"X":       tc.X,
				"Y":       tc.Y,
				"Zoom":    tc.Zoom,
				"LayerId": tc.LayerId,
				"prefix":  "tilerenderjob",
			}
			logrus.WithFields(fields).Debug("Tile render job dispatch tile")

			//dispatch re-render
			jobs <- tc
		}

		//spin down worker pool
		close(jobs)

		for j := 0; j < ctx.Config.RenderingJobs; j++ {
			<-done
		}
	}

	return tilecount
}

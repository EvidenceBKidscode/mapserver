package tilerendererjob

import (
	"mapserver/app"
	"mapserver/coords"
	"mapserver/settings"
	"time"
	"fmt"
	"github.com/sirupsen/logrus"
)

type IncrementalRenderEvent struct {
	LastMtime int64 `json:"lastmtime"`
	Progress float64 `json:"progress"`
}

func incrementalRender(ctx *app.App) {
	// Lastpos is not stored. In case of crash, blocks will be processed again
	mtime := ctx.Settings.GetInt64(settings.SETTING_LAST_MTIME, 0) + 1

	fields := logrus.Fields{
		"LastMtime": mtime,
	}
	logrus.WithFields(fields).Info("Starting incremental rendering job")

	for true {
		count, newMtime, err := ctx.MapBlockAccessor.CountModifiedBlocks(mtime)
		if err != nil {
			panic(err)
		}
		if (count == 0) {
			time.Sleep(2 * time.Second)
			continue
		}
		fmt.Printf("Found %d modified blocks\n", count)

		ev := IncrementalRenderEvent{ Progress: 0 }
		ctx.WebEventbus.Emit("incremental-render-progress", &ev)
		done := 0
		pos := coords.NewMapBlockCoords(coords.MinCoord, coords.MinCoord, coords.MinCoord)

		for true {
			start := time.Now()
			result, err := ctx.MapBlockAccessor.FindModifiedBlocks(mtime, pos, ctx.Config.IncrementalFetchLimit, ctx.Config.Layers)

			if err != nil {
				panic(err)
			}

			if len(result.List) == 0 && !result.HasMore {
				break;
			}

			tiles := renderMapblocks(ctx, result.List)

			done = done + len(result.List)
			pos.X = result.LastPos.X
			pos.Y = result.LastPos.Y
			pos.Z = result.LastPos.Z
			mtime = result.LastMtime

			fmt.Printf("Done so far :%d\n", done)

			t := time.Now()
			elapsed := t.Sub(start)

			ev = IncrementalRenderEvent{ Progress: float64(done) / float64(count) }
			ctx.WebEventbus.Emit("incremental-render-progress", &ev)

			fields := logrus.Fields{
				"mapblocks": len(result.List),
				"tiles":     tiles,
				"elapsed":   elapsed,
			}
			logrus.WithFields(fields).Info("incremental rendering")

			//tile gc
			ctx.TileDB.GC()
		}
		ctx.Settings.SetInt64(settings.SETTING_LAST_MTIME, newMtime)
		mtime = newMtime + 1

		ev = IncrementalRenderEvent{ Progress: 1 }
		ctx.WebEventbus.Emit("incremental-render-progress", &ev)
	}
}

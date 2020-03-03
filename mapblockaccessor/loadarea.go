package mapblockaccessor

import (
	"mapserver/coords"
	"mapserver/eventbus"
	"mapserver/mapblockparser"

	cache "github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)


func (a *MapBlockAccessor) PreloadArea(pos1 *coords.MapBlockCoords, pos2 *coords.MapBlockCoords) (error) {
/*	fields := logrus.Fields{
		"lastmtime": mtime,
		"limit":     limit,
	}
	logrus.WithFields(fields).Debug("FindModifiedBlocks")
*/
	lock.Lock()
	defer lock.Unlock()

	timer := prometheus.NewTimer(dbGetMtimeDuration)
	blocks, err := a.accessor.FindBlocksInArea(pos1, pos2)
	timer.ObserveDuration()

	if err != nil {
		return err
	}

	for _, block := range blocks {
		key := getKey(block.Pos)

		cached, found := a.blockcache.Get(key)
		if found && cached != nil {
			continue;
		}

		mapblock, err := mapblockparser.Parse(block.Data, block.Mtime, block.Pos)
		if err != nil {
			fields := logrus.Fields{
				"x":   block.Pos.X,
				"y":   block.Pos.Y,
				"z":   block.Pos.Z,
				"err": err,
			}
			logrus.WithFields(fields).Error("parse error")
			continue
		}

		a.Eventbus.Emit(eventbus.MAPBLOCK_RENDERED, mapblock)

		a.blockcache.Set(key, mapblock, cache.DefaultExpiration)
		cacheBlockCount.Inc()
	}

	return nil
}

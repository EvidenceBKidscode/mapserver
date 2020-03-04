package mapblockaccessor

import (
	"mapserver/coords"
	"mapserver/eventbus"
	"mapserver/mapblockparser"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

func (a *MapBlockAccessor) LoadSector(x int, z int, miny int, maxy int) (map[int]*mapblockparser.MapBlock, error) {
	result := make(map[int]*mapblockparser.MapBlock)

	lock.Lock()
	defer lock.Unlock()

	timer := prometheus.NewTimer(dbGetMtimeDuration)
	blocks, err := a.accessor.FindBlocksInArea(coords.NewMapBlockCoords(x, miny, z), coords.NewMapBlockCoords(x, maxy, z))
	timer.ObserveDuration()

	if err != nil {
		return nil, err
	}

	for _, block := range blocks {
		mb, found := a.blockcache.Get(block.Pos)
		if found {
			// Block could have been cached as inexistant
			if mb != nil {
				result[block.Pos.Y] = mb
			}
			continue
		}

		mb, err := mapblockparser.Parse(block.Data, block.Mtime, block.Pos)
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

		a.Eventbus.Emit(eventbus.MAPBLOCK_RENDERED, mb)
		a.blockcache.Set(block.Pos, mb)
		cacheBlockCount.Inc()
		result[block.Pos.Y] = mb
	}

	return result, nil
}

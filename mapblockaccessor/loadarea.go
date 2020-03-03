package mapblockaccessor

import (
	"mapserver/coords"
	"mapserver/eventbus"
	"mapserver/mapblockparser"

	cache "github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

func (a *MapBlockAccessor) LoadSector(x int, z int, miny int, maxy int) (map[int]*mapblockparser.MapBlock, error) {
	// CODE DUPLIQUE de get.go
	//maintenance
	cacheBlocks.Set(float64(a.blockcache.ItemCount()))
	if a.blockcache.ItemCount() > a.maxcount {
		//flush cache
		fields := logrus.Fields{
			"cached items": a.blockcache.ItemCount(),
			"maxcount":     a.maxcount,
		}
		logrus.WithFields(fields).Debug("Flushing cache")

		a.blockcache.Flush()
	}

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
		key := getKey(block.Pos)

		cached, found := a.blockcache.Get(key)

		if found && cached != nil {
			result[block.Pos.Y] = cached.(*mapblockparser.MapBlock)
		} else {
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

			a.blockcache.Set(key, mb, cache.DefaultExpiration)
			cacheBlockCount.Inc()
			result[block.Pos.Y] = mb
		}
	}
	return result, nil
}

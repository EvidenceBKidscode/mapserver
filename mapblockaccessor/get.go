package mapblockaccessor

import (
	"mapserver/coords"
	"mapserver/eventbus"
	"mapserver/mapblockparser"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var lock = &sync.RWMutex{}

func (a *MapBlockAccessor) GetMapBlock(pos *coords.MapBlockCoords) (*mapblockparser.MapBlock, error) {
	//read section
	lock.RLock()

	cachedblock, found := a.blockcache.Get(pos)
	if found {
		defer lock.RUnlock()

		getCacheHitCount.Inc()
		return cachedblock, nil
	}

	//end read
	lock.RUnlock()

	timer := prometheus.NewTimer(dbGetDuration)
	defer timer.ObserveDuration()

	//write section
	lock.Lock()
	defer lock.Unlock()

	//try read
	cachedblock, found = a.blockcache.Get(pos)
	if found {
		getCacheHitCount.Inc()
		return cachedblock, nil
	}

	block, err := a.accessor.GetBlock(pos)
	if err != nil {
		return nil, err
	}

	if block == nil {
		//no mapblock here
		cacheBlockCount.Inc()
		a.blockcache.Set(pos, nil)
		return nil, nil
	}

	getCacheMissCount.Inc()

	mapblock, err := mapblockparser.Parse(block.Data, block.Mtime, pos)
	if err != nil {
		return nil, err
	}

	a.Eventbus.Emit(eventbus.MAPBLOCK_RENDERED, mapblock)

	cacheBlockCount.Inc()
	a.blockcache.Set(pos, mapblock)

	return mapblock, nil
}

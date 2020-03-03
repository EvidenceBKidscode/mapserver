package mapblockaccessor

import (
	"mapserver/coords"
	"mapserver/mapblockparser"
)

func (a *MapBlockAccessor) Update(pos *coords.MapBlockCoords, mb *mapblockparser.MapBlock) {
	cacheBlockCount.Inc()
	a.blockcache.Set(pos, mb)
}

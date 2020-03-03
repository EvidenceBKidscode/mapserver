package mapblockaccessor

import (
	"fmt"
	"mapserver/coords"
	"mapserver/db"
	"mapserver/eventbus"

	"time"
)

type MapBlockAccessor struct {
	accessor   db.DBAccessor
	blockcache *BlockCache
	Eventbus   *eventbus.Eventbus
}

func getKey(pos *coords.MapBlockCoords) string {
	return fmt.Sprintf("Coord %d/%d/%d", pos.X, pos.Y, pos.Z)
}

func NewMapBlockAccessor(accessor db.DBAccessor, expiretime, purgetime time.Duration, maxcount int) *MapBlockAccessor {
	blockcache := NewBlockCache(maxcount)

	return &MapBlockAccessor{
		accessor:   accessor,
		blockcache: blockcache,
		Eventbus:   eventbus.New(),
	}
}

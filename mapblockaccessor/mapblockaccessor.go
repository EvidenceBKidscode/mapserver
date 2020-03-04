package mapblockaccessor

import (
	"mapserver/db"
	"mapserver/eventbus"

	"time"
)

type MapBlockAccessor struct {
	accessor   db.DBAccessor
	blockcache *BlockCache
	Eventbus   *eventbus.Eventbus
}

func NewMapBlockAccessor(accessor db.DBAccessor, expiretime, purgetime time.Duration, maxcount int) *MapBlockAccessor {
	blockcache := NewBlockCache(maxcount)

	return &MapBlockAccessor{
		accessor:   accessor,
		blockcache: blockcache,
		Eventbus:   eventbus.New(),
	}
}

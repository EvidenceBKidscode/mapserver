package mapblockaccessor

import (
	"fmt"
	"mapserver/coords"
	"mapserver/db"
	"mapserver/layer"
	"mapserver/mapblockparser"
	"time"

	cache "github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
)

type MapBlockAccessor struct {
	accessor  db.DBAccessor
	c         *cache.Cache
	listeners []MapBlockListener
}

type MapBlockListener interface {
	OnParsedMapBlock(block *mapblockparser.MapBlock)
}

func getKey(pos coords.MapBlockCoords) string {
	return fmt.Sprintf("Coord %d/%d/%d", pos.X, pos.Y, pos.Z)
}

func NewMapBlockAccessor(accessor db.DBAccessor) *MapBlockAccessor {
	c := cache.New(500*time.Millisecond, 1000*time.Millisecond)

	return &MapBlockAccessor{accessor: accessor, c: c}
}

func (a *MapBlockAccessor) AddListener(l MapBlockListener) {
	a.listeners = append(a.listeners, l)
}

func (a *MapBlockAccessor) Update(pos coords.MapBlockCoords, mb *mapblockparser.MapBlock) {
	key := getKey(pos)
	a.c.Set(key, mb, cache.DefaultExpiration)
}

type FindMapBlocksResult struct {
	HasMore         bool
	LastPos         *coords.MapBlockCoords
	List            []*mapblockparser.MapBlock
	UnfilteredCount int
}

func (a *MapBlockAccessor) FindMapBlocks(lastpos coords.MapBlockCoords, lastmtime int64, limit int, layerfilter []layer.Layer) (*FindMapBlocksResult, error) {

	blocks, err := a.accessor.FindBlocks(lastpos, lastmtime, limit)

	if err != nil {
		return nil, err
	}

	result := FindMapBlocksResult{}

	mblist := make([]*mapblockparser.MapBlock, 0)
	var newlastpos *coords.MapBlockCoords
	result.HasMore = len(blocks) == limit
	result.UnfilteredCount = len(blocks)

	for _, block := range blocks {
		newlastpos = &block.Pos

		inLayer := false
		for _, l := range layerfilter {
			if (block.Pos.Y*16) >= l.From && (block.Pos.Y*16) <= l.To {
				inLayer = true
				break
			}
		}

		if !inLayer {
			continue
		}

		fields := logrus.Fields{
			"x": block.Pos.X,
			"y": block.Pos.Y,
			"z": block.Pos.Z,
		}
		logrus.WithFields(fields).Trace("legacy mapblock")

		key := getKey(block.Pos)

		mapblock, err := mapblockparser.Parse(block.Data, block.Mtime, block.Pos)
		if err != nil {
			return nil, err
		}

		for _, listener := range a.listeners {
			listener.OnParsedMapBlock(mapblock)
		}

		a.c.Set(key, mapblock, cache.DefaultExpiration)
		mblist = append(mblist, mapblock)

	}

	result.LastPos = newlastpos
	result.List = mblist

	return &result, nil
}


func (a *MapBlockAccessor) GetMapBlock(pos coords.MapBlockCoords) (*mapblockparser.MapBlock, error) {
	key := getKey(pos)

	cachedblock, found := a.c.Get(key)
	if found {
		return cachedblock.(*mapblockparser.MapBlock), nil
	}

	block, err := a.accessor.GetBlock(pos)
	if err != nil {
		return nil, err
	}

	if block == nil {
		return nil, nil
	}

	mapblock, err := mapblockparser.Parse(block.Data, block.Mtime, pos)
	if err != nil {
		return nil, err
	}

	for _, listener := range a.listeners {
		listener.OnParsedMapBlock(mapblock)
	}

	a.c.Set(key, mapblock, cache.DefaultExpiration)

	return mapblock, nil
}
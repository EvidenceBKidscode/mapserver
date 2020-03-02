package mapblockaccessor

import (
	"mapserver/mapblockparser"
	"mapserver/coords"
	"time"
	"sync"
)

type CachedBlock struct {
	block *mapblockparser.MapBlock
	timestamp int64
}

type BlockCache struct {
	blocks map[int64]*CachedBlock
	maxcount int
	mutex sync.RWMutex
}

func NewBlockCache(maxcount int) (*BlockCache) {
	return &BlockCache{
		blocks: make(map[int64]*CachedBlock),
		maxcount: maxcount,
	}
}

func (c *BlockCache) Get(pos *coords.MapBlockCoords) (*mapblockparser.MapBlock, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if block, ok := c.blocks[coords.CoordToPlain(pos)]; ok {
		return block.block, true
	} else {
		return nil, false
	}
}

func (c* BlockCache) Set(pos *coords.MapBlockCoords, block *mapblockparser.MapBlock) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Remove oldest elements
	if (len(c.blocks) >= c.maxcount) {
		timestamp := time.Now().UnixNano()

		for _, v := range c.blocks {
			if v.timestamp < timestamp {
				timestamp = v.timestamp
			}
		}

		for k, v := range c.blocks {
			if v.timestamp <= timestamp {
				delete(c.blocks, k)
			}
		}
	}

	// Add new element
	c.blocks[coords.CoordToPlain(pos)] = &CachedBlock{block:block, timestamp:time.Now().UnixNano()}
}

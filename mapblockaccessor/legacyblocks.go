package mapblockaccessor

import (
//	"mapserver/eventbus"
	"mapserver/layer"
	"mapserver/mapblockparser"
	"mapserver/settings"

	cache "github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
)

type FindNextLegacyBlocksResult struct {
	HasMore         bool
	List            []*mapblockparser.MapBlock
	UnfilteredCount int
	Progress        float64
	LastMtime       int64
}

func (a *MapBlockAccessor) FindNextLegacyBlocks(s settings.Settings, layers []*layer.Layer, limit int) (*FindNextLegacyBlocksResult, error) {

	nextResult, err := a.accessor.FindNextInitialBlocks(s, layers, limit)

	if err != nil {
		return nil, err
	}

	blocks := nextResult.List
	result := FindNextLegacyBlocksResult{}

	mblist := make([]*mapblockparser.MapBlock, 0)
	result.HasMore = nextResult.HasMore
	result.UnfilteredCount = nextResult.UnfilteredCount
	result.Progress = nextResult.Progress
	result.LastMtime = nextResult.LastMtime

	for _, block := range blocks {

		fields := logrus.Fields{
			"x": block.Pos.X,
			"y": block.Pos.Y,
			"z": block.Pos.Z,
		}
		logrus.WithFields(fields).Trace("mapblock")
		key := getKey(block.Pos)
		/*

		mapblock, err := mapblockparser.Parse(block.Data, block.Mtime, block.Pos)
		if err != nil {
			fields := logrus.Fields{
				"x":   block.Pos.X,
				"y":   block.Pos.Y,
				"z":   block.Pos.Z,
				"err": err,
			}
			logrus.WithFields(fields).Error("mapblock-pars")

			return nil, err
		}

		a.Eventbus.Emit(eventbus.MAPBLOCK_RENDERED, mapblock)

		a.blockcache.Set(key, mapblock, cache.DefaultExpiration)
		cacheBlockCount.Inc()
*/
	mapblock := mapblockparser.NewMapblock()
	mapblock.Mtime = block.Mtime
	mapblock.Pos = block.Pos
	mapblock.Size = 0

	mapblock.Parsed = false
	mapblock.RawData = block.Data
	a.blockcache.Set(key, mapblock, cache.DefaultExpiration)
	cacheBlockCount.Inc()
/**/

		mblist = append(mblist, mapblock)

	}

	result.List = mblist

	fields := logrus.Fields{
		"len(List)":       len(result.List),
		"unfilteredCount": result.UnfilteredCount,
		"hasMore":         result.HasMore,
		"limit":           limit,
	}
	logrus.WithFields(fields).Debug("FindMapBlocksByPos:Result")

	return &result, nil
}

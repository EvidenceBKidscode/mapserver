package mapblockaccessor

import (
	"mapserver/coords"
	"mapserver/eventbus"
	"mapserver/layer"
	"mapserver/mapblockparser"

	cache "github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type FindModifiedBlocksResult struct {
	HasMore         bool
	List            []*mapblockparser.MapBlock
	UnfilteredCount int
	LastPos         coords.MapBlockCoords
	LastMtime       int64
}

func (a *MapBlockAccessor) FindModifiedBlocks(mtime int64, pos *coords.MapBlockCoords, limit int, layerfilter []*layer.Layer) (*FindModifiedBlocksResult, error) {
	fields := logrus.Fields{
		"lastmtime": mtime,
		"limit":     limit,
	}
	logrus.WithFields(fields).Debug("FindModifiedBlocks")

	timer := prometheus.NewTimer(dbGetMtimeDuration)
	blocks, err := a.accessor.FindModifiedBlocks(mtime, pos, limit)
	timer.ObserveDuration()

	if err != nil {
		return nil, err
	}

	changedBlockCount.Add(float64(len(blocks)))

	result := FindModifiedBlocksResult{}

	mblist := make([]*mapblockparser.MapBlock, 0)

	result.HasMore = len(blocks) == limit
	result.UnfilteredCount = len(blocks)

	for _, block := range blocks {
		result.LastPos.X = block.Pos.X
		result.LastPos.Y = block.Pos.Y
		result.LastPos.Z = block.Pos.Z
		result.LastMtime = block.Mtime

		currentLayer := layer.FindLayerByY(layerfilter, block.Pos.Y)

		if currentLayer == nil {
			continue
		}

		fields := logrus.Fields{
			"x": block.Pos.X,
			"y": block.Pos.Y,
			"z": block.Pos.Z,
		}
		logrus.WithFields(fields).Debug("mapblock")

		key := getKey(block.Pos)

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
		mblist = append(mblist, mapblock)

	}

	result.List = mblist

	return &result, nil
}

func (a *MapBlockAccessor) CountModifiedBlocks(mtime int64) (int, int64, error) {
	count, newmtime, err := a.accessor.CountModifiedBlocks(mtime)
	if err != nil {
		return 0, 0, err
	}
	return int(count), newmtime, nil
}

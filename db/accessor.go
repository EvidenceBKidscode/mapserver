package db

import (
	"mapserver/coords"
	"mapserver/layer"
	"mapserver/settings"
)

type Block struct {
	Pos   *coords.MapBlockCoords
	Data  []byte
	Mtime int64
}

type InitialBlocksResult struct {
	List            []*Block
	UnfilteredCount int
	HasMore         bool
	Progress        float64
	LastMtime       int64
}

type DBAccessor interface {
	Migrate() error

	GetTimestamp() (int64, error)
	FindBlocksByMtime(gtmtime int64, limit int) ([]*Block, error)
	FindNextInitialBlocks(s settings.Settings, layers []*layer.Layer, limit int) (*InitialBlocksResult, error)
	GetBlock(pos *coords.MapBlockCoords) (*Block, error)
	FindModifiedBlocks(mtime int64, pos *coords.MapBlockCoords, limit int) ([]*Block, error)
	CountModifiedBlocks(mtime int64) (int64, int64, error)
	FindBlocksInArea(pos1 *coords.MapBlockCoords, pos2 *coords.MapBlockCoords) ([]*Block, error)
}

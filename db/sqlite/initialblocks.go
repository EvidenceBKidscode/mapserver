package sqlite

import (
	"mapserver/coords"
	"mapserver/db"
	"mapserver/layer"
	"mapserver/settings"

	_ "github.com/mattn/go-sqlite3"
)

const (
	SETTING_LAST_POS               = "last_pos"
	SETTING_TOTAL_LEGACY_COUNT     = "total_legacy_count"
	SETTING_PROCESSED_LEGACY_COUNT = "total_processed_legacy_count"
)

/*const getLastBlockQuery = `
select pos,data,mtime
from blocks b
where b.pos > ?
order by b.pos asc, b.mtime asc
limit ?
`*/
const getLastBlockQuery = `
select b.x, b.z, max(b.mtime)
from blocks b
where b.x > ?1 or b.x == ?1 and b.z > ?2
group by b.x, b.z
order by b.x, b.z
limit ?
`

// TODO:ADD LAYER Y in query

// TODO:Return sectors, not blocks

func (this *Sqlite3Accessor) FindNextInitialBlocks(s settings.Settings, layers []*layer.Layer, limit int) (*db.InitialBlocksResult, error) {
	result := &db.InitialBlocksResult{}

	blocks := make([]*db.Block, 0)
	lastpos := coords.PlainToCoord(s.GetInt64(SETTING_LAST_POS, coords.MinPlainCoord-1))

	processedcount := s.GetInt64(SETTING_PROCESSED_LEGACY_COUNT, 0)
	totallegacycount := s.GetInt64(SETTING_TOTAL_LEGACY_COUNT, -1)
	if totallegacycount == -1 {
		//Query from db
		totallegacycount, err := this.CountBlocks()

		if err != nil {
			panic(err)
		}

		s.SetInt64(SETTING_TOTAL_LEGACY_COUNT, int64(totallegacycount))
	}


	rows, err := this.db.Query(getLastBlockQuery, lastpos.X, lastpos.Z, limit)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		result.HasMore = true
		result.UnfilteredCount++

		var x int
		var z int
		var mtime int64

		err = rows.Scan(&x, &z, &mtime)
		if err != nil {
			return nil, err
		}

		if mtime > result.LastMtime {
			result.LastMtime = mtime
		}

		mb := db.Block{Pos: coords.NewMapBlockCoords(x, 20, z), Data: nil, Mtime: 0}

		// new position
		lastpos.X = x
		lastpos.Z = z

/*TODO:MANAGE LAYERS

		blockcoordy := mb.Pos.Y
		currentlayer := layer.FindLayerByY(layers, blockcoordy)

		if currentlayer != nil {
*/			blocks = append(blocks, &mb)
//		}
	}

	s.SetInt64(SETTING_PROCESSED_LEGACY_COUNT, int64(result.UnfilteredCount)+processedcount)

	result.Progress = float64(processedcount) / float64(totallegacycount)
	result.List = blocks

	//Save current positions of initial run
	s.SetInt64(SETTING_LAST_POS, coords.CoordToPlain(lastpos))

	return result, nil
}

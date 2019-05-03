package mapobject

import (
	"mapserver/mapblockparser"
	"mapserver/mapobjectdb"
)

type PoiBlock struct {
	Color string
}

func (this *PoiBlock) onMapObject(x, y, z int, block *mapblockparser.MapBlock) *mapobjectdb.MapObject {
	md := block.Metadata.GetMetadata(x, y, z)

	o := mapobjectdb.NewMapObject(block.Pos, x, y, z, "poi")
	o.Attributes["name"] = md["name"]
	o.Attributes["category"] = md["category"]
	o.Attributes["url"] = md["url"]
	o.Attributes["owner"] = md["owner"]
	o.Attributes["icon"] = md["icon"]
	o.Attributes["color"] = this.Color

	return o
}

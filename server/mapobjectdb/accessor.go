package mapobjectdb

import (
	"mapserver/coords"
	"time"

	"github.com/sirupsen/logrus"
)

/*
sqlite perf: https://stackoverflow.com/questions/1711631/improve-insert-per-second-performance-of-sqlite?rq=1
PRAGMA synchronous = OFF
PRAGMA journal_mode = MEMORY
*/

type Tile struct {
	Pos   *coords.TileCoords
	Data  []byte
	Mtime int64
}

type MapObject struct {
	//mapblock position
	MBPos *coords.MapBlockCoords `json:"mapblock"`

	//block position
	X int `json:"x"`
	Y int `json:"y"`
	Z int `json:"z"`

	Type       string            `json:"type"`
	Mtime      int64             `json:"mtime"`
	Attributes map[string]string `json:"attributes"`
}

func NewMapObject(MBPos *coords.MapBlockCoords, x int, y int, z int, _type string) *MapObject {

	fields := logrus.Fields{
		"mbpos": MBPos,
		"x":     x,
		"y":     y,
		"z":     z,
		"type":  _type,
	}
	log.WithFields(fields).Debug("NewMapObject")

	o := MapObject{
		MBPos:      MBPos,
		Type:       _type,
		X:          (MBPos.X * 16) + x,
		Y:          (MBPos.Y * 16) + y,
		Z:          (MBPos.Z * 16) + z,
		Mtime:      time.Now().Unix(),
		Attributes: make(map[string]string),
	}

	return &o
}

type SearchAttributeLike struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type SearchQuery struct {
	//mapblock position
	Pos1          *coords.MapBlockCoords `json:"pos1"`
	Pos2          *coords.MapBlockCoords `json:"pos2"`
	Type          string                 `json:"type"`
	AttributeLike *SearchAttributeLike   `json:"attributelike"`
}

type Subscription struct {
	Id       *int64
	Endpoint string
	Auth     string
	P256DH   string
	X, Y, Z  int
}

type DBAccessor interface {
	//migrates the database
	Migrate() error

	//Generic map objects (poi, etc)
	GetMapData(q *SearchQuery) ([]*MapObject, error)
	RemoveMapData(pos *coords.MapBlockCoords) error
	AddMapData(data *MapObject) error

	//Settings
	GetSetting(key string, defaultvalue string) (string, error)
	SetSetting(key string, value string) error

	//subscriptions
	GetSubscriptionsByMapblockPos(*coords.MapBlockCoords) ([]Subscription, error)
	GetSubscriptionsByEndpoint(endpoint string) ([]Subscription, error)
	AddSubscription(sub *Subscription) error
	RemoveSubscription(sub *Subscription) error
}

package mapobjectdb

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
	"time"
)

const migrateScript = `
create table if not exists objects(
	id integer primary key autoincrement,
  x int,
  y int,
	z int,
	posx int,
	posy int,
	posz int,
	type varchar,
	data blob,
  mtime bigint
);

create table if not exists tiles(
  data blob,
  mtime bigint,
  layerid int,
  x int,
  y int,
  zoom int,
  primary key(x,y,zoom,layerid)
);
`

type Sqlite3Accessor struct {
	db       *sql.DB
	filename string
}

func (db *Sqlite3Accessor) Migrate() error {
	log.WithFields(logrus.Fields{"filename": db.filename}).Info("Migrating database")
	start := time.Now()
	_, err := db.db.Exec(migrateScript)
	if err != nil {
		return err
	}
	t := time.Now()
	elapsed := t.Sub(start)
	log.WithFields(logrus.Fields{"elapsed": elapsed}).Info("Migration completed")

	return nil
}
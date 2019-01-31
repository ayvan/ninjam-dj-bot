package tracks

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

type JamTracksDB interface {
	Tracks() ([]*Track, error)
	Track(id uint) (*Track, error)
	Playlists() ([]*Playlist, error)
	Playlist(id uint) (*Playlist, error)
}

var _ JamTracksDB = &JamDB{} // check interface implementation

type JamDB struct {
	db *gorm.DB
}

func NewJamDB(file string) (jamDB *JamDB, err error) {
	var db *gorm.DB
	db, err = gorm.Open("sqlite3", file)
	if err != nil {
		err = fmt.Errorf("failed to connect database: %s", err)
		return
	}

	if err = db.AutoMigrate(&Track{}, &Tag{}, &Playlist{}, &PlaylistTrack{}).Error; err != nil {
		err = fmt.Errorf("failed to migrate database: %s", err)
		return
	}

	jamDB = &JamDB{
		db: db,
	}

	return
}

func (jdb *JamDB) DBClose() {
	jdb.db.Close()
}

func (jdb *JamDB) DB() *gorm.DB {
	return jdb.db
}

func (jdb *JamDB) Tags() (tags []*Tag, err error) {
	tags = []*Tag{}
	err = jdb.db.Find(&tags).Error
	return
}

func (jdb *JamDB) Tracks() (tracks []*Track, err error) {
	tracks = []*Track{}
	err = jdb.db.Preload("Tags").Find(&tracks).Error
	return
}

func (jdb *JamDB) Track(id uint) (res *Track, err error) {
	track := &Track{}
	dbRes := jdb.db.Preload("Tags").First(&track, "id", id)
	if dbRes.RecordNotFound() || dbRes.Error != nil {
		return
	}

	res = track

	return
}

func (jdb *JamDB) TrackByPath(path string) (res *Track, err error) {
	track := &Track{}
	dbRes := jdb.db.Preload("Tags").First(&track, "file_path", "path")
	if dbRes.RecordNotFound() || dbRes.Error != nil {
		return
	}

	res = track

	return
}

func (jdb *JamDB) Playlists() (playlists []*Playlist, err error) {
	playlists = []*Playlist{}
	err = jdb.db.Preload("Tags").Find(&playlists).Error
	return
}

func (jdb *JamDB) Playlist(id uint) (res *Playlist, err error) {
	playlist := &Playlist{}
	dbRes := jdb.db.Preload("Tags").First(&playlist, "id", id)
	if dbRes.RecordNotFound() || dbRes.Error != nil {
		return
	}

	res = playlist

	return
}

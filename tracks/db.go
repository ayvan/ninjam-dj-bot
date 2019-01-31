package tracks

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"time"
)

type Model struct {
	ID        uint       `gorm:"primary_key" json:"id"`
	CreatedAt time.Time  `json:"-"`
	UpdatedAt time.Time  `json:"-"`
	DeletedAt *time.Time `sql:"index" json:"-"`
}

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

var ErrorNotFound = fmt.Errorf("track not found")

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
	dbRes := jdb.db.Preload("Tags").First(&track, id)
	if dbRes.RecordNotFound() {
		err = ErrorNotFound
		return
	}
	if dbRes.Error != nil {
		err = dbRes.Error
	}

	res = track

	return
}

func (jdb *JamDB) TrackUpdate(id uint, req *Track) (res *Track, err error) {
	track, err := jdb.Track(uint(id))
	if err != nil {
		return
	}

	// данные модели ORM, путь к файлу, число проигрываний менять запрещено,
	// остальное - разрешено
	req.Model = track.Model
	req.FilePath = track.FilePath
	req.Played = track.Played

	db := jdb.db.Omit("tags").Save(&req)
	if db.Error != nil {
		err = db.Error
		return
	}

	association := jdb.db.Model(&req).Association("Tags").Replace(req.Tags)
	if association.Error != nil {
		err = association.Error
		return
	}

	res = &Track{}
	dbRes := jdb.db.Preload("Tags").First(res, id)
	if dbRes.RecordNotFound() {
		err = ErrorNotFound
		return
	}
	if dbRes.Error != nil {
		err = dbRes.Error
	}

	return
}

func (jdb *JamDB) Tag(id uint) (res *Tag, err error) {
	tag := &Tag{}
	dbRes := jdb.db.First(&tag, id)
	if dbRes.RecordNotFound() {
		err = ErrorNotFound
		return
	}
	if dbRes.Error != nil {
		err = dbRes.Error
	}

	res = tag

	return
}

func (jdb *JamDB) TagUpdate(id uint, req *Tag) (res *Tag, err error) {
	tag, err := jdb.Tag(uint(id))
	if err != nil {
		return
	}

	// данные модели ORM, путь к файлу, число проигрываний менять запрещено,
	// остальное - разрешено
	req.Model = tag.Model

	db := jdb.db.Save(&req)
	if db.Error != nil {
		err = db.Error
		return
	}

	res = &Tag{}
	dbRes := jdb.db.First(res, id)
	if dbRes.RecordNotFound() {
		err = ErrorNotFound
		return
	}
	if dbRes.Error != nil {
		err = dbRes.Error
	}

	return
}

func (jdb *JamDB) TrackByPath(path string) (res *Track, err error) {
	track := &Track{}
	dbRes := jdb.db.Preload("Tags").First(&track, "file_path = ?", path)
	if dbRes.RecordNotFound() {
		err = ErrorNotFound
		return
	}
	if dbRes.Error != nil {
		err = dbRes.Error
	}

	res = track

	return
}

func (jdb *JamDB) Playlists() (playlists []*Playlist, err error) {
	playlists = []*Playlist{}
	err = jdb.db.Preload("Tracks").Find(&playlists).Error
	return
}

func (jdb *JamDB) Playlist(id uint) (res *Playlist, err error) {
	playlist := &Playlist{}
	dbRes := jdb.db.Preload("Tracks").First(&playlist, id)
	if dbRes.RecordNotFound() {
		err = ErrorNotFound
		return
	}
	if dbRes.Error != nil {
		err = dbRes.Error
	}

	res = playlist

	return
}

func (jdb *JamDB) PlaylistUpdate(id uint, req *Playlist) (res *Playlist, err error) {
	playlist, err := jdb.Playlist(uint(id))
	if err != nil {
		return
	}

	// данные модели ORM, путь к файлу, число проигрываний менять запрещено,
	// остальное - разрешено
	req.Model = playlist.Model

	db := jdb.db.Save(&req)
	if db.Error != nil {
		err = db.Error
		return
	}

	res = &Playlist{}
	dbRes := jdb.db.First(res, id)
	if dbRes.RecordNotFound() {
		err = ErrorNotFound
		return
	}
	if dbRes.Error != nil {
		err = dbRes.Error
	}

	return
}

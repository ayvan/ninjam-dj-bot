package tracks

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/sirupsen/logrus"
)

var dbCacheID = make(map[uint]*Track)
var db *gorm.DB

func DBClose() {
	db.Close()
}

func DB() *gorm.DB {
	return db
}

func Init(file string) {
	var err error
	db, err = gorm.Open("sqlite3", file)
	if err != nil {
		panic("failed to connect database")
	}

	if err := db.AutoMigrate(&Track{}, &Tag{}).Error; err != nil {
		panic("failed to migrate database")
	}

	//tags := []Tag{
	//	{Name:"tag1"},
	//	{Name:"tag2"},
	//}
	//
	//for _,t:=range tags{
	//	db.Save(&t)
	//}
}

func LoadCache() {
	tracks := []*Track{}

	if err := db.Preload("Tags").Find(&tracks).Error; err != nil {
		logrus.Errorf("db.Find error in LoadCache: %s", err)
	}

	for _, t := range tracks {
		dbCacheID[t.ID] = t
	}
}

func Tags() (tags []*Tag, err error) {
	tags = []*Tag{}
	err = db.Find(&tags).Error
	return
}

func Tracks() (tracks []*Track, err error) {
	tracks = []*Track{}
	err = db.Preload("Tags").Find(&tracks).Error
	return
}

func TrackFromCache(id uint) *Track {
	return dbCacheID[id]
}

func TrackByPath(path string) (res *Track, err error) {
	track := &Track{}
	dbRes := db.Preload("Tags").First(&track, "file_path", "path")
	if dbRes.RecordNotFound() || dbRes.Error != nil {
		return
	}

	res = track

	return
}

package tracks

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/sirupsen/logrus"
)

var dbCache = make(map[uint]*Track)
var db *gorm.DB

func CloseDB() {
	db.Close()
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
		dbCache[t.ID] = t
	}
}

func GetTracks() (res []*Track) {
	for _, track := range dbCache {
		res = append(res, track)
	}

	return
}

func GetTrack(id uint) *Track {
	return dbCache[id]
}

func GetTags() []Tag {
	tags := []Tag{}
	if err := db.Find(&tags).Error; err != nil {
		logrus.Errorf("db.Find error in GetTags: %s", err)
	}

	return tags
}

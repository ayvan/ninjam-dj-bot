package auth

import (
	"fmt"
	"github.com/jinzhu/gorm"
)

var ErrorNotFound = fmt.Errorf("not found")

type User struct {
	gorm.Model
	Username     string
	PasswordHash string
}

type DB struct {
	db *gorm.DB
}

func NewDB(file string) (jamDB *DB, err error) {
	var db *gorm.DB
	db, err = gorm.Open("sqlite3", file)
	if err != nil {
		err = fmt.Errorf("failed to connect database: %s", err)
		return
	}

	if err = db.AutoMigrate(&User{}).Error; err != nil {
		err = fmt.Errorf("failed to migrate database: %s", err)
		return
	}

	jamDB = &DB{
		db: db,
	}

	return
}

func (db *DB) DBClose() {
	db.DB().Close()
}

func (db *DB) DB() *gorm.DB {
	return db.db
}

func (db *DB) UserCreate(req *User) (res *User, err error) {
	dbRes := db.DB().Save(&req)
	if dbRes.Error != nil {
		err = dbRes.Error
		return
	}

	res = req

	return
}

func (db *DB) UserByName(username string) (res *User, err error) {
	track := &User{}
	dbRes := db.DB().First(&track, "username = ?", username)
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

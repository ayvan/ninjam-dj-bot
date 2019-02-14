package main

import (
	"github.com/ayvan/ninjam-dj-bot/tracks"
	"github.com/ayvan/ninjam-dj-bot/tracks_sync"
	"github.com/sirupsen/logrus"
	"os"
	"path"
	"path/filepath"
)

func main() {
	if len(os.Args) == 1 {
		logrus.Fatalf("you must specify tracks directory")
	}

	var err error
	dir, err := filepath.Abs(os.Args[1])
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Infof("start processing tracks in %s", dir)

	// помещаем БД в директорию с треками
	db, err := tracks.NewJamDB(path.Join(dir, "tracks.db"))
	if err != nil {
		logrus.Fatal(err)
	}
	defer db.DBClose()

	tracks_sync.Init(dir, db)

	if err = filepath.Walk(dir, tracks_sync.Walk); err != nil {
		logrus.Fatal(err)
	}
}

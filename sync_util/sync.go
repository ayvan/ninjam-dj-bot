package main

import (
	"github.com/Ayvan/ninjam-dj-bot/tracks"
	"github.com/Ayvan/ninjam-dj-bot/tracks_sync"
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
	dir, err := filepath.Abs(filepath.Dir(os.Args[1]))
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Infof("start processing tracks in %s", dir)

	// помещаем БД в директорию с треками
	tracks.Init(path.Join(dir, "tracks.db"))
	defer tracks.DBClose()
	tracks.LoadCache()

	tracks_sync.Init(dir)

	if err = filepath.Walk(dir, tracks_sync.Walk); err != nil {
		logrus.Fatal(err)
	}
}

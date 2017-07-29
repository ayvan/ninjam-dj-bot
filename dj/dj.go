package dj

import (
	"fmt"
	"github.com/Ayvan/ninjam-dj-bot/config"
	"github.com/Sirupsen/logrus"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
	"time"
)

type trackInfo struct {
	URL      string
	Name     string
	Key      string
	BPM      string
	FileName string
}

var loadedFiles map[string]bool = make(map[string]bool)
var tracks []*trackInfo = make([]*trackInfo, 0)
var tracksByKey map[string][]*trackInfo = make(map[string][]*trackInfo)
var sigChan chan bool = make(chan bool, 1)
var stopPlayChan chan bool = make(chan bool, 1)
var nextTrackChan chan *trackInfo = make(chan *trackInfo, 1)

var player *os.Process

func Start() {
	load()

	go func() {
		for {
			select {
			case <-stopPlayChan:
				stopMP3()
			case s := <-sigChan:
				stopMP3()
				sigChan <- s
				return
			case t := <-nextTrackChan:
				launchMP3(t.FileName)
			}
		}
	}()
}

func Stop() {
	sigChan <- true
}

func TracksCount() map[string]int {
	ts := map[string]int{}

	for key, t := range tracksByKey {
		ts[key] = len(t)
	}

	return ts
}

func load() {
	dir := config.Get().Player.Dir

	files, err := ioutil.ReadDir(dir)

	if err != nil {
		logrus.Error(err)
		return
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		r := regexp.MustCompile(`^([a-zA-Z#]+)___([\d]+)___([\s\S]+)\.mp3$`)

		s := r.FindStringSubmatch(file.Name())
		if len(s) > 0 {
			logrus.Info(file.Name())
			loadedFiles[file.Name()] = true
			ti := trackInfo{
				Name:     s[3],
				Key:      s[1],
				BPM:      s[2],
				FileName: path.Join(dir, file.Name()),
			}

			tracks = append(tracks, &ti)
			tracksByKey[ti.Key] = append(tracksByKey[ti.Key], &ti)
		}
	}
}

func StopMP3() {
	stopPlayChan <- true
}

func stopMP3() {
	if player != nil {
		err := player.Signal(os.Interrupt)
		if err != nil {
			logrus.Error("player.Signal(os.Interrupt):", err)
		}
	}
}

func Random() (string, string) {
	r := rand.New(rand.NewSource(time.Now().Unix()))

	l := len(tracks)

	if l == 0 {
		return "Playlist is empty:(", ""
	}

	n := r.Intn(l)

	t := tracks[n]

	nextTrack(t)

	return fmt.Sprintf("Playing %s, key %s", t.Name, t.Key), t.BPM
}

func RandomKey(key string) (string, string) {
	r := rand.New(rand.NewSource(time.Now().Unix()))

	l := len(tracksByKey[key])

	if l == 0 {
		return "Playlist is empty:(", ""
	}

	n := r.Intn(l)

	t := tracksByKey[key][n]

	nextTrack(t)

	return fmt.Sprintf("Playing %s, key %s", t.Name, t.Key), t.BPM
}

func nextTrack(t *trackInfo) {
	nextTrackChan <- t
}

func launchMP3(mp3File string) {

	if player != nil {
		err := player.Signal(os.Interrupt)
		logrus.Error("player.Signal(os.Interrupt):", err)
	}

	app := config.Get().Player.Command
	args := config.Get().Player.Args

	argsSlice := strings.Split(args, " ")

	argsSlice = append(argsSlice, mp3File)

	cmd := exec.Command(app, argsSlice...)
	logrus.Debug("Player path: ", cmd.Path)

	err := cmd.Start()

	if err != nil {
		logrus.Error("Player start error: ", err)
		return
	}

	player = cmd.Process
	logrus.Info("Player started")
}

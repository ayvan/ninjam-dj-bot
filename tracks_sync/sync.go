package tracks_sync

import (
	"fmt"
	"github.com/Ayvan/ninjam-dj-bot/tracks"
	"github.com/bogem/id3v2"
	"github.com/burillo-se/bs1770wrap"
	"github.com/sirupsen/logrus"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var mp3regex = regexp.MustCompile(`\.mp3$`)

var dir string
var jamDB *tracks.JamDB

func Init(d string, db *tracks.JamDB) {
	dir = d
	jamDB = db
}

func Walk(path string, info os.FileInfo, err error) error {
	if err != nil {
		logrus.Fatal(err)
	}
	if !info.IsDir() {
		if mp3regex.MatchString(info.Name()) {
			ProcessMP3Track(path)
		}
	}
	return nil
}

func AnalyzeMP3Track(trackPath string) (track *tracks.Track, err error) {
	// сделаем путь файла относительным, от текущей директории
	relativePath := strings.TrimLeft(strings.TrimPrefix(trackPath, dir), "./")

	tag, err := id3v2.Open(trackPath, id3v2.Options{Parse: true})
	if err != nil {
		err = fmt.Errorf("id3v2.Open error for %s: %s", trackPath, err)
		logrus.Error(err)
		return
	}

	var trackNumber int
	num := strings.Trim(tag.GetTextFrame(tag.CommonID("Track number/Position in set")).Text, "\x00")

	trackNumber, _ = strconv.Atoi(num)

	track = &tracks.Track{
		FilePath:         relativePath,
		Title:            tag.Title(),
		Artist:           tag.Artist(),
		Album:            tag.Album(),
		AlbumTrackNumber: uint(trackNumber),
	}

	frames := tag.GetFrames("PRIV")

	for _, frame := range frames {
		// тут ошибки не критичны, трек сохранится в БД "как есть" без информации - его можно будет редактировать
		f, ok := frame.(id3v2.UnknownFrame)
		if ok {
			frameStruct := private_ext_frame_data{}
			_, data := getFrameNameAndData(f.Body)
			err := frameStruct.Unmarshal(data)
			if err != nil {
				logrus.Warn(err)
				continue
			}

			if frameStruct.version != 2 {
				err = fmt.Errorf("bad tag version: %d", frameStruct.version)
				logrus.Warn(err)
				continue
			}

			trackData := frameStruct.data

			track.Key = trackData.Key()
			track.Mode = trackData.Mode()
			track.BPM = trackData.BPM()
			track.BPI = trackData.BPI()
			track.LoopStart = trackData.LoopStart()
			track.LoopEnd = trackData.LoopEnd()

		}
	}

	ldata, err := bs1770wrap.CalculateLoudness(trackPath)
	if err != nil {
		err = fmt.Errorf("bs1770wrap.CalculateLoudness: %s", err)
		logrus.Error(err)
		return
	}

	track.Length = ldata.Length

	track.Integrated = ldata.Integrated
	track.Range = ldata.Range
	track.Peak = ldata.Peak
	track.Shortterm = ldata.Shortterm
	track.Momentary = ldata.Momentary

	return
}

func ProcessMP3Track(path string) (track *tracks.Track, err error) {
	logrus.Infof("starting analyze track %s", path)
	track, err = AnalyzeMP3Track(path)
	if err != nil {
		err = fmt.Errorf("AnalyzeMP3Track for %s: %s", path, err)
		logrus.Error(err)
		return
	}

	// проверяем, есть ли уже трек в базе
	if trackInDB, _ := jamDB.TrackByPath(track.FilePath); trackInDB != nil {
		// если трек есть - назначим ID нашему треку и запись обновится вместо добавления
		track.ID = trackInDB.ID
		// данные, которые из тегов MP3 не извлекаем, тоже следует перенести
		track.Tags = trackInDB.Tags
		track.Played = trackInDB.Played
	}

	if err = jamDB.DB().Save(track).Error; err != nil {
		err = fmt.Errorf("add track error for %s: %s", path, err)
		logrus.Error(err)
		return
	}

	return
}

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

func Init(d string) {
	dir = d
}

func Walk(path string, info os.FileInfo, err error) error {
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
		f, ok := frame.(id3v2.UnknownFrame)
		if ok {
			frameStruct := private_ext_frame_data{}
			_, data := getFrameNameAndData(f.Body)
			frameStruct.Unmarshal(data)

			if frameStruct.version != 2 {
				err = fmt.Errorf("bad tag version: %d", frameStruct.version)
				logrus.Error(err)
				return
			}

			trackData := frameStruct.data

			track.Key = uint(trackData.key)
			track.Mode = uint(trackData.mode)
			track.BPM = uint(trackData.bpm)
			track.BPI = uint(trackData.bpi)
			track.LoopStart = uint64(trackData.ls)
			track.LoopEnd = uint64(trackData.le)

		}
	}

	loudnessData, err := bs1770wrap.CalculateLoudness(trackPath)
	if err != nil {
		err = fmt.Errorf("bs1770wrap.CalculateLoudness: %s", err)
		logrus.Error(err)
		return
	}

	track.Loudness = loudnessData.IntegratedLoudness
	track.LoudnessPeak = loudnessData.TruePeak
	track.LoudnessRange = loudnessData.LoudnessRange
	track.Length = loudnessData.Length

	return
}

func ProcessMP3Track(path string) (track *tracks.Track, err error) {
	track, err = AnalyzeMP3Track(path)
	if err != nil {
		err = fmt.Errorf("AnalyzeMP3Track for %s: %s", path, err)
		logrus.Error(err)
		return
	}

	// проверяем, есть ли уже трек в базе
	if trackInDB, _ := tracks.TrackByPath(path); trackInDB != nil {
		// если трек есть - назначим ID нашему треку и запись обновится вместо добавления
		track.ID = trackInDB.ID
		// данные, которые из тегов MP3 не извлекаем, тоже следует перенести
		track.Tags = trackInDB.Tags
		track.Played = trackInDB.Played
	}

	if err = tracks.DB().Save(track).Error; err != nil {
		err = fmt.Errorf("add track error for %s: %s", path, err)
		logrus.Error(err)
		return
	}

	return
}

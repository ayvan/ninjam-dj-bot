package sync

import (
	"fmt"
	"github.com/Ayvan/ninjam-dj-bot/tracks"
	"github.com/bogem/id3v2"
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

func analyzeMP3Track(path string) (track *tracks.Track, err error) {
	// сделаем путь файла относительным, от текущей директории
	relativePath := strings.TrimLeft(strings.TrimPrefix(path, dir), "/")
	logrus.Infof("%s", relativePath)

	tag, err := id3v2.Open(path, id3v2.Options{Parse: true})
	if err != nil {
		err = fmt.Errorf("id3v2.Open error for %s: %s", path, err)
		logrus.Error(err)
		return
	}
	fmt.Println(string(tag.Version()))

	getReplayGain(tag)

	var trackNumber int
	num := strings.Trim(tag.GetTextFrame(tag.CommonID("Track number/Position in set")).Text, "\x00")

	trackNumber, _ = strconv.Atoi(num)

	track = &tracks.Track{
		FilePath:         relativePath,
		Title:            tag.Title(),
		Artist:           tag.Artist(),
		Album:            tag.Album(),
		AlbumTrackNumber: uint(trackNumber),

		// TODO извлечь прочие данные из тегов
	}

	frames := tag.GetFrames("PRIV")

	for _, frame := range frames {
		f, ok := frame.(id3v2.UnknownFrame)
		if ok {
			frameStruct := private_ext_frame_data{}
			_, data := getFrameNameAndData(f.Body)
			frameStruct.Unmarshal(data)

			fmt.Printf("%x %v\n", frameStruct.magic, frameStruct)
		}
	}

	return
}

func ProcessMP3Track(path string) (track *tracks.Track, err error) {
	track, err = analyzeMP3Track(path)
	if err != nil {
		err = fmt.Errorf("analyzeMP3Track for %s: %s", path, err)
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

func getReplayGain(tag *id3v2.Tag) {
	frames := tag.GetFrames("TXXX")

	for _, frame := range frames {
		frameData, ok := frame.(id3v2.UserDefinedTextFrame)
		if !ok {
			continue
		}
		fmt.Println(frameData.Description, frameData.Value)
	}
}

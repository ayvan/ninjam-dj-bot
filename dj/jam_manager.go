package dj

import (
	"github.com/Ayvan/ninjam-dj-bot/config"
	"github.com/Ayvan/ninjam-dj-bot/tracks"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"math/rand"
	"time"
)

const (
	messageAlreadyStarted           = "playing already started"
	messageCantStartRandomTrack     = "can't start random track"
	messageUnableToRecognizeCommand = "unable to recognize command, please use \"dj help\" to get the list and format of the available commands"

	errorGeneral = "an error has occurred"
)

var p *message.Printer

func init() {
	message.SetString(language.Russian, messageAlreadyStarted, "воспроизведение уже запущено")
	message.SetString(language.Russian, messageCantStartRandomTrack, "не удалось запустить случайный трек")
	message.SetString(language.Russian, messageUnableToRecognizeCommand, "невозможно распознать команду, используйте \"dj help\" для получения списка и формата доступных команд")
	message.SetString(language.Russian, errorGeneral, "произошла ошибка")

	p = message.NewPrinter(config.Language)
}

type Manager interface {
	Playlists() []tracks.Playlist
	PlayRandom(command JamCommand) string
	StartPlaylist(id uint) string
	StartTrack(id uint) string
	Stop() string
}

var _ Manager = &JamManager{} // check interface implementation

type playingMode uint

const (
	playingTrack playingMode = iota + 1
	playingPlaylist
)

type JamManager struct {
	playingMode playingMode // playing single track or playing list of tracks
	playlistID  uint
	trackID     uint
	playing     bool // играем или нет в данный момент

	bpm int
	bpi int

	jamPlayer *JamPlayer
	jamDB     tracks.JamTracksDB
}

type JamChatCommand struct {
	Command string
	Param   string
	Tags    []string
	ID      uint
}

type JamCommand struct {
	Command uint
	Param   string
	Key     uint
	Mode    uint
	ID      uint
	Tags    []uint
}

func NewJamManager(jamDB tracks.JamTracksDB, player *JamPlayer) *JamManager {
	jm := &JamManager{
		jamPlayer: player,
		jamDB:     jamDB,
	}
	player.SetOnStop(jm.onStop)
	return jm
}

func (jm *JamManager) Playlists() (res []tracks.Playlist) {
	return
}

func (jm *JamManager) PlayRandom(command JamCommand) (msg string) {
	jm.Stop()

	count, err := jm.jamDB.CountTracks()

	if err != nil {
		logrus.Error(err)
		return p.Sprint(errorGeneral)
	}

	randSource := rand.NewSource(time.Now().UnixNano())
	randomizer := rand.New(randSource)

	i := 0
	var track *tracks.Track
	for {
		i++
		if i > 1000 {
			msg = p.Sprint(messageCantStartRandomTrack)
			return
		}
		id := uint(randomizer.Intn(int(count)))

		track, err = jm.jamDB.Track(id)
		if err == tracks.ErrorNotFound {
			continue
		} else if err != nil {
			logrus.Error(err)
			return p.Sprint(errorGeneral)
		}

		if command.Key != 0 {
			if track.Key != command.Key {
				continue
			}
		}
		break
	}

	jm.jamPlayer.LoadTrack(track)
	jm.jamPlayer.SetRepeats(0)

	jm.trackID = track.ID
	jm.playlistID = 0
	jm.playingMode = playingTrack

	jm.Start()
	return // TODO msg
}

func (jm *JamManager) StartPlaylist(id uint) (msg string) {
	return
}

func (jm *JamManager) StartTrack(id uint) (msg string) {
	return
}

func (jm *JamManager) Stop() (msg string) {
	jm.jamPlayer.Stop()
	jm.playing = false
	return // todo msg
}

func (jm *JamManager) Start() (msg string) {
	if jm.playing == true {
		return p.Sprint(messageAlreadyStarted)
	}
	jm.playing = true
	err := jm.jamPlayer.Start()
	if err != nil {
		logrus.Error(err)
		return p.Sprint(errorGeneral)
	}
	return
}

func (jm *JamManager) Command(chatCommand string) string {
	command := Command(CommandParse(chatCommand))

	switch command.Command {
	case CommandRandom:
		return jm.PlayRandom(command)
	case CommandTrack:
	case CommandPlaylist:
	case CommandStop:
		return jm.Stop()
	case CommandPlay:
		return jm.Start()
	case CommandNext:
	case CommandPrev:
	case CommandPlaying:
	default:
		return p.Sprint(messageUnableToRecognizeCommand)
	}

	return ""
}

func (jm *JamManager) onStop() {
	if jm.playingMode == playingPlaylist {
		// TODO найти какой трек в листе мы играем и запустить следующий, или сообщить что плейлист окончен
		// если у нас jm.playing == false значит стоп пришёл т.к. мы сами дали команды на стоп - тогда ничего не делаем
	}

	jm.playing = false
}

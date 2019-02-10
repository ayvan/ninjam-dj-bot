package dj

import (
	"fmt"
	"github.com/Ayvan/ninjam-dj-bot/tracks"
	"math/rand"
	"time"
)

type Manager interface {
	Playlists() []tracks.Playlist
	PlayRandom(command JamCommand) error
	StartPlaylist(id uint) error
	StartTrack(id uint) error
	Stop()
}

var _ Manager = &JamManager{} // check interface implementation

const (
	playingTrack = iota + 1
	playingPlaylist
)

type JamManager struct {
	playingMode uint // играем трек или плейлист
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

func (jm *JamManager) PlayRandom(command JamCommand) (err error) {
	jm.Stop()

	count, err := jm.jamDB.CountTracks()

	if err != nil {
		return
	}

	randSource := rand.NewSource(time.Now().UnixNano())
	randomizer := rand.New(randSource)

	i := 0
	var track *tracks.Track
	for {
		i++
		if i > 1000 {
			err = fmt.Errorf("не удалось запустить случайный трек")
			return
		}
		id := uint(randomizer.Intn(int(count)))

		track, err = jm.jamDB.Track(id)
		if err != nil && err == tracks.ErrorNotFound {
			continue
		} else if err != nil {
			return
		}

		if command.Key != 0 {
			if track.Key != command.Key {
				continue
			}
		}
		break
	}

	jm.jamPlayer.setMP3Source(track.FilePath)
	jm.jamPlayer.SetRepeats(0)
	jm.jamPlayer.setBPI(track.BPI)
	jm.jamPlayer.setBPM(track.BPM)

	jm.trackID = track.ID
	jm.playlistID = 0
	jm.playingMode = playingTrack

	jm.Start()
	return
}

func (jm *JamManager) StartPlaylist(id uint) (err error) {
	return
}

func (jm *JamManager) StartTrack(id uint) (err error) {
	return
}

func (jm *JamManager) Stop() {
	jm.jamPlayer.Stop()
	jm.playing = false
	return
}

func (jm *JamManager) Start() (err error) {
	if jm.playing == true {
		return fmt.Errorf("проигрывание уже запущено")
	}
	jm.playing = true
	err = jm.jamPlayer.Start()
	return
}

func (jm *JamManager) Command(chatCommand string) string {
	command := Command(CommandParse(chatCommand))

	switch command.Command {
	case CommandRandom:
		jm.PlayRandom(command)
	case CommandTrack:
	case CommandPlaylist:
	case CommandStop:
		jm.Stop()
	case CommandPlay:
		jm.Start()
	case CommandNext:
	case CommandPrev:
	case CommandPlaying:
	default:
		return `невозможно распознать команду, используйте "help" для получения списка и формата доступных команд`
	}

	return ""
}

func (jm *JamManager) onStop() {
	jm.playing = false

	if jm.playingMode == playingPlaylist {
		// TODO найти какой трек в листе мы играем и запустить следующий, или сообщить что плейлист окончен
	}
}

package dj

import "github.com/Ayvan/ninjam-dj-bot/tracks"

type Manager interface {
	Playlists() []tracks.Playlist
	PlayRandom() error
	PlayRandomKey(key string) error
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
	playingMode int // играем трек или плейлист
	playlistID  uint
	trackID     uint
	playing     bool // играем или нет в данный момент

	bpm int
	bpi int

	jamPlayer *JamPlayer
	jamDB     tracks.JamTracksDB
}

func NewJamManager(jamDB tracks.JamTracksDB, player *JamPlayer) *JamManager {
	return &JamManager{
		jamPlayer: player,
		jamDB:     jamDB,
	}
}

func (jm *JamManager) Playlists() (res []tracks.Playlist) {
	return
}

func (jm *JamManager) PlayRandom() (err error) {
	return
}

func (jm *JamManager) PlayRandomKey(key string) (err error) {
	return
}

func (jm *JamManager) StartPlaylist(id uint) (err error) {
	return
}

func (jm *JamManager) StartTrack(id uint) (err error) {
	return
}

func (jm *JamManager) Stop() {
	return
}

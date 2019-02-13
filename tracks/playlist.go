package tracks

import (
	"encoding/json"
	"fmt"
	"strings"
)

type PlaylistSlice []Playlist

type PlaylistTrack struct {
	TrackID uint `json:"track_id"`
	Repeats uint `json:"repeats"` // число повторений зацикленной части трека
	Timeout uint `json:"timeout"` // пауза после трека
	Queue   bool `json:"queue"`   // действует ли очередь во время трека
}

type Playlist struct {
	Model
	Name        string `json:"name"`
	Description string `json:"description"`
	// TargetTrackTime время трека в секундах, по-умолчанию для добавляемого трека, на его основе будет рассчитано число повторов трека
	TargetTrackTime uint            `json:"target_track_time"`
	Tracks          []PlaylistTrack `json:"tracks"`
	TracksJSON      []byte          `json:"-"`
}

func (ps PlaylistSlice) String() (res string) {
	for _, playlist := range ps {
		if playlist.Description != "" {
			res += fmt.Sprintf("%s (%d tracks) - %s", playlist.Name, len(playlist.Tracks), playlist.Description)
		} else {
			res += fmt.Sprintf("%s (%d tracks)", playlist.Name, len(playlist.Tracks))
		}
	}

	res = strings.TrimRight(res, "\n")
	return
}

func (p *Playlist) BeforeSave() (err error) {
	b, err := json.Marshal(p.Tracks)

	if err != nil {
		return
	}

	p.TracksJSON = b

	return
}

func (p *Playlist) BeforeUpdate() (err error) {
	b, err := json.Marshal(p.Tracks)

	if err != nil {
		return
	}

	p.TracksJSON = b

	return
}

func (p *Playlist) AfterFind() (err error) {

	p.Tracks = make([]PlaylistTrack, 0)
	if len(p.TracksJSON) != 0 {
		err = json.Unmarshal(p.TracksJSON, &p.Tracks)

		if err != nil {
			return
		}
	}

	return
}

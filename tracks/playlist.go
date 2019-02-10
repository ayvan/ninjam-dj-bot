package tracks

import (
	"fmt"
	"strings"
)

type PlaylistSlice []Playlist
type PlaylistTrackSlice []PlaylistTrack

type PlaylistTrack struct {
	Model
	PlaylistID uint `json:"playlist_id"`
	TrackID    uint `json:"track_id"`
	Repeats    uint `json:"repeats"` // число повторений зацикленной части трека
	Timeout    uint `json:"timeout"` // пауза после трека
	Order      int  `json:"order"`   // порядок трека в плейлисте
	Queue      bool `json:"queue"`   // действует ли очередь во время трека
}

type Playlist struct {
	Model
	Name        string `json:"name"`
	Description string `json:"description"`
	// TrackTime время трека в секундах, по-умолчанию для добавляемого трека, на его основе будет рассчитано число повторов трека
	TrackTime uint            `json:"track_time"`
	Tracks    []PlaylistTrack `json:"tracks"`
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

func (s PlaylistTrackSlice) Len() int {
	return len(s)
}

func (s PlaylistTrackSlice) Less(i, j int) bool {
	return s[i].Order < s[j].Order
}

func (s PlaylistTrackSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

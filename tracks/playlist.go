package tracks

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"strings"
)

type PlaylistSlice []Playlist

type PlaylistTrack struct {
	gorm.Model
	PlaylistID uint     `json:"playlist_id"`
	Track      Track    `json:"track"`
	Playlist   Playlist `json:"playlist"`
	Repeats    uint     `json:"repeats"` // число повторений зацикленной части трека
	Timeout    uint     `json:"timeout"` // пауза после трека
	Queue      bool     `json:"queue"`   // действует ли очередь во время трека
}

type Playlist struct {
	gorm.Model
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

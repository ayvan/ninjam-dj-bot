package tracks

import (
	"github.com/jinzhu/gorm"
)

type Track struct {
	gorm.Model
	FileName string `json:"file_name"`

	Title            string `json:"title"`
	Artist           string `json:"artist"`
	Album            string `json:"album"`
	AlbumTrackNumber uint   `json:"album_track_number"`
	Tags             []Tag  `gorm:"many2many:track_tags;"`
	Played           uint64 `json:"played"`

	// JamPlayer info
	LoopStart     uint64  `json:"loop_start"`
	LoopEnd       uint64  `json:"loop_end"`
	Loudness      float32 `json:"loudness"`
	LoudnessRange float32 `json:"loudness_range"`
	LoudnessPeak  float32 `json:"loudness_peak"`
	BPM           uint    `json:"bpm"`
	BPI           uint    `json:"bpi"`
	Key           string  `json:"key"`
}

type Tag struct {
	gorm.Model
	Name string `json:"name"`
}

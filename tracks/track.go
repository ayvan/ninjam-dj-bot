package tracks

import (
	"github.com/jinzhu/gorm"
)

const (
	KeyUnknown = iota
	KeyA
	KeyASharp
	KeyB
	KeyC
	KeyCSharp
	KeyD
	KeyDSharp
	KeyE
	KeyESharp
	KeyF
	KeyFSharp
	KeyG
	KeyGSharp
)

const (
	ModeUnknown = iota
	ModeMinor
	ModeMajor
)

const (
	KeyNameUnknown = "Unknown"
	KeyNameA       = "A"
	KeyNameASharp  = "A#"
	KeyNameB       = "B"
	KeyNameC       = "C"
	KeyNameCSharp  = "C#"
	KeyNameD       = "D"
	KeyNameDSharp  = "D#"
	KeyNameE       = "E"
	KeyNameESharp  = "E#"
	KeyNameF       = "F"
	KeyNameFSharp  = "F#"
	KeyNameG       = "G"
	KeyNameGSharp  = "G#"
)

var KeysMapping = map[uint]string{
	KeyUnknown: KeyNameUnknown,
	KeyA:       KeyNameA,
	KeyASharp:  KeyNameASharp,
	KeyB:       KeyNameB,
	KeyC:       KeyNameC,
	KeyCSharp:  KeyNameCSharp,
	KeyD:       KeyNameD,
	KeyDSharp:  KeyNameDSharp,
	KeyE:       KeyNameE,
	KeyESharp:  KeyNameESharp,
	KeyF:       KeyNameF,
	KeyFSharp:  KeyNameFSharp,
	KeyG:       KeyNameG,
	KeyGSharp:  KeyNameGSharp,
}

const (
	ModeNameMinor = "minor"
	ModeNameMajor = "major"
)

var ModesMapping = map[uint]string{
	ModeMinor: ModeNameMinor,
	ModeMajor: ModeNameMajor,
}

type Track struct {
	gorm.Model
	FilePath string `json:"file_path"`

	Title            string `json:"title"`
	Artist           string `json:"artist"`
	Album            string `json:"album"`
	AlbumTrackNumber uint   `json:"album_track_number"`
	Tags             []Tag  `gorm:"many2many:track_tags;"`
	Played           uint64 `json:"played"`

	AuthorInfo string `json:"author_info"`
	AlbumInfo  string `json:"album_info"`
	TrackInfo  string `json:"track_info"`

	// JamPlayer info
	Length        float32 `json:"length"`
	LoopStart     uint64  `json:"loop_start"`
	LoopEnd       uint64  `json:"loop_end"`
	Loudness      float32 `json:"loudness"`
	LoudnessRange float32 `json:"loudness_range"`
	LoudnessPeak  float32 `json:"loudness_peak"`
	BPM           uint    `json:"bpm"`
	BPI           uint    `json:"bpi"`
	Key           uint    `json:"key"`
	Mode          uint    `json:"mode"`
}

type Tag struct {
	gorm.Model
	Name string `json:"name"`
}

func (t Track) KeyString() string {
	return KeysMapping[t.Key] + " " + ModesMapping[t.Mode]
}

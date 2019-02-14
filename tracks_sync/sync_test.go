package tracks_sync

import (
	"fmt"
	"github.com/ayvan/ninjam-dj-bot/tracks"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_analyzeMP3Track(t *testing.T) {

	track, err := AnalyzeMP3Track("DrumLoop.mp3")
	assert.NoError(t, err)

	fmt.Println(track)
}

func TestUpdateMP3Track(t *testing.T) {
	t.Skip()
	track, err := AnalyzeMP3Track("Dynamic Drums.mp3")
	assert.NoError(t, err)

	fmt.Println(track.Key, track.Title, track.Artist, track.AlbumTrackNumber, track.BPI, track.BPM, track.LoopStart, track.LoopEnd)

	trackBackup := &tracks.Track{}

	*trackBackup = *track

	track.BPM = 100
	track.BPI = 16
	track.Key = 5
	track.Mode = 2
	track.AlbumTrackNumber = 6
	track.Title = "testing"
	track.Artist = "Tester"
	track.Album = "Testo"

	err = UpdateMP3Track(track)
	assert.NoError(t, err)

	trackUpdated, err := AnalyzeMP3Track("Dynamic Drums.mp3")
	assert.NoError(t, err)

	fmt.Println(trackUpdated.Key, trackUpdated.Title, trackUpdated.Artist, trackUpdated.AlbumTrackNumber, trackUpdated.BPI, trackUpdated.BPM)

	assert.EqualValues(t, track, trackUpdated)

	err = UpdateMP3Track(trackBackup)
	assert.NoError(t, err)
}

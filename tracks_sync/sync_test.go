package tracks_sync

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_analyzeMP3Track(t *testing.T) {

	track, err := AnalyzeMP3Track("DrumLoop.mp3")
	assert.NoError(t, err)

	fmt.Println(track)
}

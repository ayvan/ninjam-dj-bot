package helpers

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_IsMP3(t *testing.T) {
	assert.True(t, IsMP3("Some Good Music.mp3"))
	assert.True(t, IsMP3("Some Good Music 3.mp3"))
	assert.True(t, IsMP3("Track1.mp3"))

	assert.False(t, IsMP3("Some Good Music.mp4"))
	assert.False(t, IsMP3("Some Good Music 3.mp4"))
	assert.False(t, IsMP3("Track1.mp4"))

	assert.False(t, IsMP3("Some Good Music.ogg"))
	assert.False(t, IsMP3("Some Good Music 3.ogg"))
	assert.False(t, IsMP3("Track1.ogg"))

	assert.False(t, IsMP3("Track mp3"))
	assert.False(t, IsMP3("Trackmp3"))
}

func Test_NewFileName(t *testing.T) {
	f, err := NewFileName("Some Good Music.mp3")
	assert.Equal(t, "Some Good Music 2.mp3", f)
	assert.NoError(t, err)

	f, err = NewFileName("Some Good Music 2.mp3")
	assert.Equal(t, "Some Good Music 3.mp3", f)
	assert.NoError(t, err)

	f, err = NewFileName("Track1.mp3")
	assert.Equal(t, "Track1 2.mp3", f)
	assert.NoError(t, err)

	f, err = NewFileName("Track 1.mp3")
	assert.Equal(t, "Track 2.mp3", f)
	assert.NoError(t, err)

	f, err = NewFileName("Some Good Music.ogg")
	assert.Equal(t, "Some Good Music 2.ogg", f)
	assert.NoError(t, err)

	f, err = NewFileName("Some Good Music 2.ogg")
	assert.Equal(t, "Some Good Music 3.ogg", f)
	assert.NoError(t, err)

	f, err = NewFileName("Track1.ogg")
	assert.Equal(t, "Track1 2.ogg", f)
	assert.NoError(t, err)

	f, err = NewFileName("Track 1.ogg")
	assert.Equal(t, "Track 2.ogg", f)
	assert.NoError(t, err)

	f, err = NewFileName("")
	assert.Equal(t, "", f)
	if assert.Error(t, err) {
		assert.Equal(t, "bad file name", err.Error())
	}

	f, err = NewFileName("test track")
	assert.Equal(t, "", f)
	if assert.Error(t, err) {
		assert.Equal(t, "bad file name", err.Error())
	}
}

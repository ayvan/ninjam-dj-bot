package lib

import (
	"github.com/ayvan/ninjam-dj-bot/tracks"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestKeyModeByName(t *testing.T) {
	keyMode := KeyModeByName("G#m")
	assert.Equal(t, tracks.KeyGSharp, keyMode.Key)
	assert.Equal(t, tracks.ModeMinor, keyMode.Mode)

	keyMode = KeyModeByName("G#")
	assert.Equal(t, tracks.KeyGSharp, keyMode.Key)
	assert.Equal(t, tracks.ModeMajor, keyMode.Mode)

	keyMode = KeyModeByName("G")
	assert.Equal(t, tracks.KeyG, keyMode.Key)
	assert.Equal(t, tracks.ModeMajor, keyMode.Mode)

	keyMode = KeyModeByName("Gm")
	assert.Equal(t, tracks.KeyG, keyMode.Key)
	assert.Equal(t, tracks.ModeMinor, keyMode.Mode)
}

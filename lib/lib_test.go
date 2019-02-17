package lib

import (
	"github.com/stretchr/testify/assert"

	"testing"
	"time"
)

func Test_calcUserPlayDuration(t *testing.T) {

	dur := CalcUserPlayDuration(time.Minute*10 + time.Second*15)
	assert.Equal(t, time.Minute*2+time.Second*3, dur)

	dur = CalcUserPlayDuration(time.Minute*5 + time.Second*15)
	assert.Equal(t, time.Minute*1+time.Second*45, dur)

	dur = CalcUserPlayDuration(time.Minute*5 + time.Second*10)
	assert.Equal(t, time.Minute*2+time.Second*35, dur)

	dur = CalcUserPlayDuration(time.Minute*5 + time.Second*30)
	assert.Equal(t, time.Minute*1+time.Second*50, dur)
}

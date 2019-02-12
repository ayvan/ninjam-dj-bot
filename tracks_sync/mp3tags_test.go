package tracks_sync

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_private_ext_frame_data_v3(t *testing.T) {
	p := &private_ext_frame_data_v3{}

	p.bpi = 4
	p.bpm = 121
	p.key = 3
	p.ls = 1961235
	p.le = 2445884
	p.mode = 1

	d := p.Marshal()

	p2 := &private_ext_frame_data_v3{}

	p2.Unmarshal(d)

	assert.EqualValues(t, p, p2)
}

func Test_private_ext_frame_data_v2(t *testing.T) {
	p := &private_ext_frame_data_v2{}

	p.bpi = 4
	p.bpm = 121
	p.key = 3
	p.ls = 1961235
	p.le = 2445884
	p.mode = 1

	d := p.Marshal()

	p2 := &private_ext_frame_data_v2{}

	p2.Unmarshal(d)

	assert.EqualValues(t, p, p2)
}

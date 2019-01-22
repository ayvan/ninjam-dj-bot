package tracks

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_load(t *testing.T) {
	Init("./tracks.db")
	LoadCache()

	tags, err := Tags()
	assert.NoError(t, err)

	fmt.Println(tags)

	fmt.Println(Tracks())
}

package tracks

import (
	"fmt"
	"testing"
)

func Test_load(t *testing.T) {
	Init("./tracks.db")
	LoadCache()

	tags := GetTags()

	fmt.Println(tags)

	fmt.Println(GetTracks())
}

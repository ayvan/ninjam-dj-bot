package lib

import "time"

func CalcUserPlayDuration(trackDuration time.Duration) time.Duration {

	plays := trackDuration / (time.Second * 105)
	return trackDuration / plays
}

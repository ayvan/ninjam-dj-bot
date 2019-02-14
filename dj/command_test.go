package dj

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCommandParse(t *testing.T) {

	cases := map[string]JamChatCommand{
		"random": {Command: "random"},
		"  	 random  ": {Command: "random"},
		"random A":                     {Command: "random", Param: "A"},
		"random Dm":                    {Command: "random", Param: "Dm"},
		"random A [blues]":             {Command: "random", Param: "A", Tags: []string{"blues"}},
		"random Dm [metal,death]":      {Command: "random", Param: "Dm", Tags: []string{"metal", "death"}},
		"random [metal, death]":        {Command: "random", Param: "", Tags: []string{"metal", "death"}},
		" random   [metal,  death]":    {Command: "random", Param: "", Tags: []string{"metal", "death"}},
		"random C [metal,death] (10m)": {Command: "random", Param: "C", Tags: []string{"metal", "death"}, Duration: time.Minute * 10},
		"random [blues] (5m 30s)":      {Command: "random", Param: "", Tags: []string{"blues"}, Duration: time.Minute*5 + time.Second*30},
		" track  123":                  {Command: "track", ID: 123},
		" play 123   ":                 {Command: "play", ID: 123},
		"	list  54": {Command: "list", ID: 54},
		"	playlist  279": {Command: "playlist", ID: 279},
	}

	for commText, comm := range cases {
		command := CommandParse(commText)
		assert.EqualValues(t, comm, command, commText)
	}
}

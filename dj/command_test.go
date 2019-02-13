package dj

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCommandParse(t *testing.T) {

	cases := map[string]JamChatCommand{
		"dj random": {Command: "random"},
		"dj   	 random  ": {Command: "random"},
		"dj	random A": {Command: "random", Param: "A"},
		"dj random Dm":                {Command: "random", Param: "Dm"},
		"dj random A [blues]":         {Command: "random", Param: "A", Tags: []string{"blues"}},
		"dj random Dm [metal,death]":  {Command: "random", Param: "Dm", Tags: []string{"metal", "death"}},
		"dj random [metal, death]":    {Command: "random", Param: "", Tags: []string{"metal", "death"}},
		"dj random   [metal,  death]": {Command: "random", Param: "", Tags: []string{"metal", "death"}},
		"dj track  123":               {Command: "track", ID: 123},
		"dj play 123   ":              {Command: "play", ID: 123},
		"dj	list  54": {Command: "list", ID: 54},
		"dj	playlist  279": {Command: "playlist", ID: 279},
	}

	for commText, comm := range cases {
		command := CommandParse(commText)
		assert.EqualValues(t, comm, command, commText)
	}
}

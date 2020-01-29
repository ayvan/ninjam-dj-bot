package lib

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	CommandUnknown = iota
	CommandRandom
	CommandTrack
	CommandPlaylist
	CommandStop
	CommandPlay
	CommandNext
	CommandPrev
	CommandPlaying
	CommandHelp
	CommandQStart
	CommandQFinish
)

var commandAliases = map[uint][]string{
	CommandRandom:   {"random"},
	CommandTrack:    {"track"},
	CommandPlaylist: {"playlist", "list"},
	CommandStop:     {"stop"},
	CommandPlay:     {"play", "start"},
	CommandNext:     {"next"},
	CommandPrev:     {"prev", "previous"},
	CommandPlaying:  {"playing", "current", "now"},
	CommandHelp:     {"help"},
	CommandQStart:   {"qstart", "qs"},
	CommandQFinish:  {"qfinish", "qf"},
}

var commandMap = make(map[string]uint)

func init() {
	for command, aliases := range commandAliases {
		for _, alias := range aliases {
			commandMap[alias] = command
		}
	}
}

type JamChatCommand struct {
	Command  string
	Param    string
	Tags     []string
	ID       uint
	Duration time.Duration
}

type JamCommand struct {
	Command  uint
	Param    string
	Key      uint
	Mode     uint
	ID       uint
	Tags     []uint
	Duration time.Duration
}

func commandByName(name string) uint {
	return commandMap[strings.ToLower(name)]
}

var commandRegexp = regexp.MustCompile(`(\w+)[ \t]*([\w#]*)[ \t]*(?:\[([\w, ]+)\])*[\t ]*(?:\(([\w ]+)\))*`)

func CommandParse(command string) (jamCommand JamChatCommand) {
	commandStrings := commandRegexp.FindStringSubmatch(command)

	if len(commandStrings) == 0 {
		return
	}

	jamCommand = JamChatCommand{}

	jamCommand.Command = strings.Trim(commandStrings[1], " ")

	if len(commandStrings) > 2 {
		commParam := strings.Trim(commandStrings[2], " ")

		if id, err := strconv.Atoi(commParam); err == nil {
			jamCommand.ID = uint(id)
		} else {
			jamCommand.Param = commParam
		}
	}

	if len(commandStrings) > 3 {
		tagsString := strings.Trim(commandStrings[3], " []")
		if tagsString != "" {

			tags := strings.Split(tagsString, ",")
			for i, tag := range tags {
				tags[i] = strings.Trim(tag, " ")
			}
			jamCommand.Tags = tags
		}
	}
	if len(commandStrings) > 4 {
		commParam := strings.Trim(commandStrings[4], " ")
		commParam = strings.Replace(commParam, " ", "", -1)
		duration, err := time.ParseDuration(commParam)
		if err == nil {
			jamCommand.Duration = duration
		}
	}

	return
}

func Command(jamChatCommand JamChatCommand) (command JamCommand) {

	command.Command = commandByName(jamChatCommand.Command)
	command.Param = jamChatCommand.Param

	keyMode := KeyModeByName(jamChatCommand.Param)
	command.Key = keyMode.Key
	command.Mode = keyMode.Mode

	command.ID = jamChatCommand.ID
	command.Duration = jamChatCommand.Duration

	return
}

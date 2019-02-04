package dj

import (
	"regexp"
	"strconv"
	"strings"
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
}

var commandMap = make(map[string]uint)

func init() {
	for command, aliases := range commandAliases {
		for _, alias := range aliases {
			commandMap[alias] = command
		}
	}
}

func commandByName(name string) uint {
	return commandMap[strings.ToLower(name)]
}

var commandRegexp = regexp.MustCompile(`(\w+)[ \t]*(\w*)[ \t]*(?:\[([\w, ]+)\])*`)

func CommandParse(command string) (jamCommand JamChatCommand) {
	commandStrings := commandRegexp.FindStringSubmatch(command)

	if len(commandStrings) == 0 {
		return
	}

	jamCommand = JamChatCommand{}

	jamCommand.Command = strings.Trim(commandStrings[1], " ")

	if len(commandStrings) > 2 {
		commParam1 := strings.Trim(commandStrings[2], " ")

		if id, err := strconv.Atoi(commParam1); err == nil {
			jamCommand.ID = uint(id)
		} else {
			jamCommand.Param = commParam1
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

	return
}

func Command(jamChatCommand JamChatCommand) (command JamCommand) {

	command.Command = commandByName(jamChatCommand.Command)
	command.Param = jamChatCommand.Param

	keyMode := keyModeByName(jamChatCommand.Param)
	command.Key = keyMode.Key
	command.Mode = keyMode.Mode

	command.ID = jamChatCommand.ID

	return
}

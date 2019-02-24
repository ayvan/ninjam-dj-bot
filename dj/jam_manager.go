package dj

import (
	"github.com/ayvan/ninjam-chatbot/models"
	"github.com/ayvan/ninjam-dj-bot/config"
	"github.com/ayvan/ninjam-dj-bot/lib"
	"github.com/ayvan/ninjam-dj-bot/tracks"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"math/rand"
	"time"
)

const (
	messageAlreadyStarted           = "playing already started"
	messageCantStartRandomTrack     = "can't start random track"
	messageUnableToRecognizeCommand = `unable to recognize command, please use "dj help" to get the list and format of the available commands`
	messagePlayingTrack             = "playing track %s, playback duration %s"
	messagePlaylistStarted          = "playlist %s started"
	helpMessage                     = "DJ Bot commands: \n" +
		"%s random - start random track\n" +
		"%s random Am - start random track with key\n" +
		"%s stop - stop track\n" +
		"%s playlist 12 - start playlist by ID\n" +
		"%s next - next track (only if playlist playing)\n" +
		"%s playing - show current track/playlist info"

	errorGeneral          = "an error has occurred"
	errorTrackNotSelected = "track not selected, please select track"
	errorTrackNotFound    = "track %d not found"
	errorPlaylistNotFound = "playlist %d not found"
	errorPlaylistIsEmpty  = "playlist %d is empty"
)

var p *message.Printer

func init() {
	message.SetString(language.Russian, messageAlreadyStarted, "воспроизведение уже запущено")
	message.SetString(language.Russian, messageCantStartRandomTrack, "не удалось запустить случайный трек")
	message.SetString(language.Russian, messageUnableToRecognizeCommand, `невозможно распознать команду, используйте "dj help" для получения списка и формата доступных команд`)
	message.SetString(language.Russian, messagePlayingTrack, "запущен трек %s, длительность воспроизведения %s")
	message.SetString(language.Russian, messagePlaylistStarted, "запущен плейлист %s")
	message.SetString(language.Russian, errorTrackNotSelected, "трек не выбран, пожалуйста, выберите трек")
	message.SetString(language.Russian, errorGeneral, "произошла ошибка")
	message.SetString(language.Russian, errorTrackNotFound, "трек %d не найден")
	message.SetString(language.Russian, errorPlaylistNotFound, "плейлист %d не найден")
	message.SetString(language.Russian, errorPlaylistIsEmpty, "плейлист %d не содержит треков")
	message.SetString(language.Russian, helpMessage, "Команды DJ-бота : \n"+
		"%s random - зпаустить случайный трек\n"+
		"%s random Am - запустить случайный трек с заданной тональностью\n"+
		"%s stop - остановить трек\n"+
		"%s playlist 12 - запустить плейлист с заданным ID\n"+
		"%s next - следующий трек (только если играет плейлист)\n"+
		"%s playing - показать информацию о текущем треке/плейлисте")

	p = message.NewPrinter(config.Language)
}

type Manager interface {
	Playlists() []tracks.Playlist
	PlayRandom(command lib.JamCommand) string
	StartPlaylist(id uint) string
	StartTrack(id uint) string
	Stop() string
}

var _ Manager = &JamManager{} // check interface implementation

type playingMode uint

const (
	playingTrack playingMode = iota + 1
	playingPlaylist
)

type JamChatBot interface {
	SendMessage(string)
	UserName() string
	SetOnUserinfoChange(f func(user models.UserInfo))
}

type JamManager struct {
	playingMode playingMode // playing single track or playing list of tracks
	playlist    *tracks.Playlist
	track       *tracks.Track
	repeats     uint
	playing     bool // играем или нет в данный момент

	jamPlayer  *JamPlayer
	jamDB      tracks.JamTracksDB
	jamChatBot JamChatBot

	queueManager *QueueManager
}

func NewJamManager(jamDB tracks.JamTracksDB, player *JamPlayer, chatBot JamChatBot) *JamManager {
	jm := &JamManager{
		jamPlayer:    player,
		jamDB:        jamDB,
		jamChatBot:   chatBot,
		queueManager: NewQueueManager(chatBot.UserName(), chatBot.SendMessage),
	}
	chatBot.SetOnUserinfoChange(jm.queueManager.OnUserinfoChange)
	player.SetOnStop(jm.onStop)
	player.SetOnStart(jm.onStart)
	return jm
}

func (jm *JamManager) Playlists() (res []tracks.Playlist) {
	return
}

func (jm *JamManager) PlayRandom(command lib.JamCommand) (msg string) {
	count, err := jm.jamDB.CountTracks()

	if err != nil {
		logrus.Error(err)
		return p.Sprint(errorGeneral)
	}

	if count == 0 {
		msg = p.Sprint(messageCantStartRandomTrack)
		return
	}

	randSource := rand.NewSource(time.Now().UnixNano())
	randomizer := rand.New(randSource)

	i := 0
	var track *tracks.Track
	for {
		i++
		if i > 1000 {
			msg = p.Sprint(messageCantStartRandomTrack)
			return
		}
		id := uint(randomizer.Intn(int(count)))

		track, err = jm.jamDB.Track(id)
		if err == tracks.ErrorNotFound {
			continue
		} else if err != nil {
			logrus.Error(err)
			return p.Sprint(errorGeneral)
		}

		if command.Key != 0 {
			if track.Key != command.Key {
				continue
			}
		}
		if command.Mode != 0 {
			if track.Mode != command.Mode {
				continue
			}
		}
		if len(command.Tags) > 0 {
			found := false
		tags:
			for _, tag := range track.Tags {
				for _, tID := range command.Tags {
					if tag.ID == tID {
						found = true
						break tags
					}
				}
			}
			if !found {
				continue
			}
		}
		break
	}

	jm.track = track
	jm.LoadTrack(jm.track)
	var repeats uint

	if command.Duration != 0 {
		repeats = jm.countRepeats(track, command.Duration)
	}

	jm.SetRepeats(repeats)

	jm.track = track
	jm.playlist = nil
	jm.playingMode = playingTrack

	return jm.Start()
}

func (jm *JamManager) StartPlaylist(id uint) (msg string) {
	jm.Stop()

	playlist, err := jm.jamDB.Playlist(id)
	if err == tracks.ErrorNotFound {
		return p.Sprintf(errorPlaylistNotFound, id)
	} else if err != nil {
		logrus.Error(err)
		return p.Sprint(errorGeneral)
	}

	if len(playlist.Tracks) == 0 {
		return p.Sprintf(errorPlaylistIsEmpty, id)
	}

	jm.playlist = playlist

	trackID := playlist.Tracks[0].TrackID
	jm.track, err = jm.jamDB.Track(trackID)
	if err == tracks.ErrorNotFound {
		return p.Sprintf(errorTrackNotFound, trackID)
	} else if err != nil {
		logrus.Error(err)
		return p.Sprint(errorGeneral)
	}

	jm.LoadTrack(jm.track)
	jm.SetRepeats(playlist.Tracks[0].Repeats)
	jm.playingMode = playingPlaylist

	msg = p.Sprintf(messagePlaylistStarted, playlist.Name)
	msg += ", "
	msg += jm.Start()
	return msg
}

func (jm *JamManager) StartTrack(id uint) (msg string) {
	return
}

func (jm *JamManager) Stop() (msg string) {
	if jm.jamPlayer.Playing() {
		jm.jamPlayer.Stop()
		jm.playing = false
	}
	return // todo msg
}

func (jm *JamManager) Start() (msg string) {
	if jm.playing == true {
		return p.Sprint(messageAlreadyStarted)
	}
	if jm.track == nil {
		return p.Sprint(errorTrackNotSelected)
	}
	jm.playing = true
	err := jm.jamPlayer.Start()
	if err != nil {
		logrus.Error(err)
		return p.Sprint(errorGeneral)
	}

	t := time.Time{}.Add(jm.calcTrackTime(jm.track, jm.repeats))
	return p.Sprintf(messagePlayingTrack, jm.track, t.Format("04:05"))
}

func (jm *JamManager) Help() (msg string) {
	msg = p.Sprintf(helpMessage,
		jm.jamChatBot.UserName(),
		jm.jamChatBot.UserName(),
		jm.jamChatBot.UserName(),
		jm.jamChatBot.UserName(),
		jm.jamChatBot.UserName(),
		jm.jamChatBot.UserName())

	return
}

func (jm *JamManager) Command(chatCommand string) string {
	command := lib.Command(lib.CommandParse(chatCommand))

	switch command.Command {
	case lib.CommandRandom:
		return jm.PlayRandom(command)
	case lib.CommandTrack:
	case lib.CommandPlaylist:
		return jm.StartPlaylist(command.ID)
	case lib.CommandStop:
		return jm.Stop()
	case lib.CommandPlay:
		return jm.Start()
	case lib.CommandNext:
		return jm.Next()
	case lib.CommandPrev:
	case lib.CommandHelp:
		return jm.Help()
	case lib.CommandPlaying:
	default:
		return p.Sprint(messageUnableToRecognizeCommand)
	}

	return ""
}

func (jm *JamManager) Next() (msg string) {
	msg, _ = jm.next()
	return
}

func (jm *JamManager) next() (msg string, ok bool) {
	var listTrack tracks.PlaylistTrack
	found := true
	for _, lTrack := range jm.playlist.Tracks {
		if lTrack.TrackID == jm.track.ID {
			found = true
			continue
		}
		// на прошлой итерации нашли текущий трек - берём следующий
		if found {
			listTrack = lTrack
		}
	}

	var err error
	if listTrack.TrackID != 0 {
		jm.track, err = jm.jamDB.Track(listTrack.TrackID)
		if err == tracks.ErrorNotFound {
			msg = p.Sprintf(errorTrackNotFound, listTrack.TrackID)
		} else if err != nil {
			logrus.Error(err)
			msg = p.Sprint(errorGeneral)
			return
		}

		jm.LoadTrack(jm.track)
		jm.SetRepeats(listTrack.Repeats)
		jm.playingMode = playingPlaylist

		msg = jm.Start()
		ok = true

		return
	}

	// TODO msg playlist ended

	return
}

func (jm *JamManager) onStart() {
	if jm.track == nil {
		return
	}

	jm.queueManager.OnStart(jm.calcTrackTime(jm.track, jm.repeats), jm.calcTrackIntervalTime(jm.track))
}

func (jm *JamManager) onStop() {
	jm.queueManager.OnStop()

	if jm.playingMode == playingPlaylist {
		// если у нас jm.playing == false значит стоп пришёл т.к. мы сами дали команды на стоп - тогда ничего не делаем
		if !jm.playing {
			// todo msg
			return
		}

		msg, ok := jm.next()
		if ok {
			jm.jamChatBot.SendMessage(msg)
			return
		}

		// TODO сообщить что плейлист окончен
	}

	jm.playing = false
}

func (jm *JamManager) countRepeats(track *tracks.Track, duration time.Duration) uint {
	if duration == 0 || track.LoopEnd <= track.LoopStart {
		return 0
	}

	trackDuration := time.Duration(track.Length) * time.Microsecond

	if trackDuration > duration {
		return 0
	}

	durationMicroS := uint64(duration / time.Microsecond)

	loopDurationMicroS := track.LoopEnd - track.LoopStart

	outroDurationMicroS := track.Length - track.LoopEnd
	introDurationMicroS := track.LoopStart

	durationMicroS = durationMicroS - introDurationMicroS - outroDurationMicroS

	repeats := uint(durationMicroS / loopDurationMicroS)

	return repeats
}

func (mk *JamManager) calcTrackTime(track *tracks.Track, repeats uint) time.Duration {
	if repeats == 0 {
		return time.Duration(track.Length) * time.Microsecond
	}
	loopDurationMicroS := track.LoopEnd - track.LoopStart

	return time.Duration(loopDurationMicroS*uint64(repeats)+track.LoopStart+track.LoopEnd) * time.Microsecond
}

func (mk *JamManager) calcTrackIntervalTime(track *tracks.Track) time.Duration {
	intervalTime := (float64(time.Minute) / float64(track.BPM)) * float64(track.BPI)
	return time.Duration(intervalTime)
}

func (jm *JamManager) SetRepeats(repeats uint) {
	jm.jamPlayer.SetRepeats(repeats)
	jm.repeats = repeats
}

func (jm *JamManager) LoadTrack(track *tracks.Track) {
	jm.jamPlayer.LoadTrack(track)
}

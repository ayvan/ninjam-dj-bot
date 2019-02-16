package dj

import (
	"encoding/binary"
	"fmt"
	"github.com/ayvan/ninjam-dj-bot/tracks"
	"github.com/azul3d/engine/audio"
	"github.com/burillo-se/lv2host-go/lv2host"
	"github.com/burillo-se/lv2hostconfig"
	"github.com/burillo-se/ninjamencoder"
	"github.com/hajimehoshi/go-mp3"
	"github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"io"
	"math"
	"os"
	"path"
	"runtime/debug"
	"time"
)

// channels всегда 2 т.к. используемый MP3-декодер всегда отдаёт звук в стерео
const channels = 2

type JamBot interface {
	IntervalBegin(guid [16]byte, channelIndex uint8)
	IntervalWrite(guid [16]byte, data []byte, flags uint8)
	SendAdminMessage(string)
}

type JamPlayer struct {
	track      *tracks.Track
	tracksPath string
	source     io.Reader
	sampleRate int
	bpm        uint
	bpi        uint
	repeats    uint
	ninjamBot  JamBot
	stop       chan bool
	playing    bool
	host       *lv2host.CLV2Host
	hostConfig *lv2hostconfig.LV2HostConfig
	onStopFunc func()
}

type AudioInterval struct {
	GUID         [16]byte
	ChannelIndex uint8
	Flags        uint8
	Data         [][]byte
	index        int // index of current audio data block
}

// TODO получать сообщения о смене bpm/bpi и форсить их назад
func NewJamPlayer(tracksPath string, ninjamBot JamBot, lv2hostConfig *lv2hostconfig.LV2HostConfig) *JamPlayer {
	return &JamPlayer{ninjamBot: ninjamBot, tracksPath: tracksPath, stop: make(chan bool, 1), hostConfig: lv2hostConfig}
}

func (jp *JamPlayer) SetOnStop(f func()) {
	jp.onStopFunc = f
}

func (jp *JamPlayer) onStop() {
	if jp.onStopFunc != nil {
		jp.onStopFunc()
	}
}

func (jp *JamPlayer) Playing() bool {
	return jp.playing
}

func (jp *JamPlayer) Track() *tracks.Track {
	return jp.track
}

func (jp *JamPlayer) LoadTrack(track *tracks.Track) {
	jp.track = track
	filePath := track.FilePath

	if !path.IsAbs(filePath) {
		filePath = path.Join(jp.tracksPath, filePath)
	}

	err := jp.setMP3Source(filePath)
	if err != nil {
		logrus.Error(err)
	}

	jp.SetRepeats(0) // по-умолчанию повторы не заданы, их должны будут задать отдельно если запуск происходит из плейлиста

	jp.hostConfig.ValueMap["integrated"] = track.Integrated
	jp.hostConfig.ValueMap["range"] = track.Range
	jp.hostConfig.ValueMap["peak"] = track.Peak
	jp.hostConfig.ValueMap["shortterm"] = track.Shortterm
	jp.hostConfig.ValueMap["momentary"] = track.Momentary

	err = jp.hostConfig.Evaluate()
	if err != nil {
		logrus.Fatal(err)
	}

	// initialize LV2 plugins
	jp.host = lv2host.Alloc(float64(jp.sampleRate))

	for i, p := range jp.hostConfig.Plugins {
		if lv2host.AddPluginInstance(jp.host, p.PluginURI) != 0 {
			logrus.Errorf("Cannot add plugin: %v\n", p.PluginURI)
			return
		}
		for param, val := range p.Data {
			if lv2host.SetPluginParameter(jp.host, uint32(i), param, val) != 0 {
				logrus.Errorf("Cannot set plugin parameter: %v\n", param)
				lv2host.ListPluginParameters(jp.host, uint32(i))
				return
			}
			logrus.Debugf("Setting '%v' to '%v'\n", param, val)
		}
	}

	lv2host.Activate(jp.host)
}

func (jp *JamPlayer) SetRepeats(repeats uint) {
	jp.repeats = repeats
}

func (jp *JamPlayer) setMP3Source(source string) error {
	jp.Stop() // stop before set new source

	out, err := os.OpenFile(source, os.O_RDONLY, 0664)
	if err != nil {
		return fmt.Errorf("setMP3Source error: %s", err)
	}

	decoder, err := mp3.NewDecoder(out)
	if err != nil {
		return fmt.Errorf("NewDecoder error: %s", err)
	}

	jp.source = decoder

	jp.sampleRate = decoder.SampleRate()

	return nil
}

func (jp *JamPlayer) setBPM(bpm uint) {
	msg := fmt.Sprintf("bpm %d", bpm)
	jp.ninjamBot.SendAdminMessage(msg)
	jp.bpm = bpm
}

func (jp *JamPlayer) setBPI(bpi uint) {
	msg := fmt.Sprintf("bpi %d", bpi)
	jp.ninjamBot.SendAdminMessage(msg)
	jp.bpi = bpi
}

func (jp *JamPlayer) Start() error {
	if jp.playing {
		return nil
	}
	if jp.source == nil || jp.track == nil {
		return fmt.Errorf("no source detected")
	}

	// default values
	var bpm, bpi uint = 100, 16
	if jp.track.BPM > 0 {
		bpm = jp.track.BPM
	}
	if jp.track.BPI > 0 {
		bpi = jp.track.BPI
	}
	jp.setBPM(bpm)
	jp.setBPI(bpi)

	jp.stop = make(chan bool, 1)

	// посчитаем на каких сэмплах у нас начало, и на каких конец зацикливания
	startTime := time.Duration(jp.track.LoopStart) * time.Microsecond
	loopStartPos := timeToSamples(startTime, jp.sampleRate) - 1 // это позиция в слайсе, потому -1

	endTime := time.Duration(jp.track.LoopEnd) * time.Microsecond
	loopEndPos := timeToSamples(endTime, jp.sampleRate) - 1 // это позиция в слайсе, потому -1

	intervalTime := (float64(time.Minute) / float64(jp.bpm)) * float64(jp.bpi)
	intervalSamples := int(math.Ceil(float64(jp.sampleRate) * intervalTime / float64(time.Second)))
	intervalSamplesChannels := intervalSamples * channels

	logrus.Debugf("Loop start pos: %d | Loop End Pos: %d", loopStartPos, loopEndPos)

	// не позволяем повторы если нет метки конца цикла либо она меньше/равна метке начала цикла
	if loopEndPos <= loopStartPos {
		jp.repeats = 0
	}

	jp.playing = true

	samplesBuffer := make([][]float32, 2)

	// эта переменная будет установлена когда буфер будет заполнен всеми данными из MP3 файла
	//bufferFull := false

	waitData := make(chan bool, 1)
	// это фоновая загрузка и декодирование MP3 в буфер
	go func() {
		intervalsReady := 0

		defer lv2host.Free(jp.host)

		for {
			buf := audio.Float32{}.Make(intervalSamplesChannels, intervalSamplesChannels)
			rs, err := toReadSeeker(jp.source, intervalSamplesChannels)
			if err != nil && err != io.EOF && err.Error() != "end of stream" {
				logrus.Errorf("source.Read error: %s", err)
			}

			var bufLen int
			bufLen, err = rs.Read(buf)
			if err != nil && err != io.EOF && err.Error() != "end of stream" {
				logrus.Errorf("source.Read error: %s", err)
			}
			if bufLen == 0 {
				//bufferFull = true
				return
			}

			deinterleavedSamples, err := ninjamencoder.DeinterleaveSamples(buf.(audio.Float32)[:bufLen], channels)
			if err != nil {
				logrus.Errorf("DeinterleaveSamples error: %s", err)
				return
			}

			lv2host.ProcessBuffer(jp.host, deinterleavedSamples[0], deinterleavedSamples[1], uint32(len(deinterleavedSamples[0])))

			for i := 0; i < channels; i++ {
				samplesBuffer[i] = append(samplesBuffer[i], deinterleavedSamples[i]...)
			}

			intervalsReady++
			if intervalsReady == 3 {
				waitData <- true
			}
		}
	}()

	// ждём пока будут готовы интервалы
	<-waitData

	// TODO на выходе функции ловить ошибку и сообщать в чат что трек прерван из-за ошибки
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logrus.Errorf("panic in JamPlayer.Start: %s", r)
				logrus.Error(string(debug.Stack()))
			}
			// если закончили - значит до того, как поставим флаг что игра трека завершена, мы подождём до конца интервала
			timer := time.NewTimer(time.Duration(intervalTime))
			<-timer.C
			jp.playing = false
			jp.onStop()
		}()

		ticker := time.NewTicker(time.Duration(intervalTime))

		oggEncoder := ninjamencoder.NewEncoder()
		oggEncoder.SampleRate = jp.sampleRate

		play := true
		currentPos := 0

		for play {
			logrus.Debugf("Current pos: %d", currentPos)
			deinterleavedSamples := make([][]float32, 2)
			endPos := currentPos + intervalSamples - 1

			if endPos > len(samplesBuffer[0])-1 {
				endPos = len(samplesBuffer[0]) - 1 // это позиция в слайсе, потому -1
				play = false                       // дошли до конца - завершаем
			}

			if endPos >= loopEndPos && jp.repeats > 0 {
				play = true // если ранее получили флаг остановки - значит снимем его, мы ушли в очередной цикл
				var loops uint
				deinterleavedSamples, currentPos, loops = loop(samplesBuffer, currentPos, loopStartPos, loopEndPos, intervalSamples, channels)

				jp.repeats -= loops
				logrus.Debugf("repeats left: %d", jp.repeats)
			} else {
				for i := 0; i < channels; i++ {
					deinterleavedSamples[i] = samplesBuffer[i][currentPos : endPos+1]
				}

				currentPos = endPos + 1
			}

			data, err := oggEncoder.EncodeNinjamInterval(deinterleavedSamples)
			if err != nil {
				logrus.Errorf("EncodeNinjamInterval error: %s", err)
				return
			}

			guid, _ := uuid.NewV1()

			select {
			case <-ticker.C:
			case <-jp.stop:
				ticker.Stop()
				return
			}

			interval := AudioInterval{
				GUID:         guid,
				ChannelIndex: 0,
				Flags:        0,
				Data:         data,
			}

			jp.ninjamBot.IntervalBegin(interval.GUID, interval.ChannelIndex)

			hasNext := true
			for hasNext {
				var intervalData []byte

				intervalData, hasNext = interval.next()

				if !hasNext {
					interval.Flags = 1
				}

				jp.ninjamBot.IntervalWrite(interval.GUID, intervalData, interval.Flags)
			}
		}
	}()

	return nil
}

func (jp *JamPlayer) Stop() {
	if jp.stop != nil && len(jp.stop) == 0 {
		jp.stop <- true
	}
}

func (ai *AudioInterval) next() (data []byte, hasNext bool) {
	hasNext = true
	if len(ai.Data) > ai.index {
		data = ai.Data[ai.index]
		ai.index++
	}
	if len(ai.Data) < ai.index+1 {
		hasNext = false
	}

	return
}

func toReadSeeker(reader io.Reader, samples int) (res audio.ReadSeeker, err error) {
	buf := audio.NewBuffer(make(audio.Float32, 0, samples))
	res = buf

	for ; samples > 0; samples-- {
		data := make([]byte, 2, 2)
		var n int
		n, err = reader.Read(data)
		if err != nil && err != io.EOF {
			return
		}
		if n == 0 {
			err = nil // remove EOF error
			return
		}

		intData := int16(binary.LittleEndian.Uint16(data))
		buf.Write(audio.Float32{Int16ToFloat32(intData)})
	}

	return
}

func Int16ToFloat32(s int16) float32 {
	return float32(s) / float32(math.MaxInt16+1)
}

func timeToSamples(t time.Duration, sampleRate int) int {
	return int(math.Round(float64(sampleRate) * float64(t) / float64(time.Second)))

}

func loop(s [][]float32, cPos, sPos, ePos, length, channels int) (res [][]float32, ncPos int, loops uint) {
	ncPos = cPos
	l := length

	res = make([][]float32, channels, channels)

	fors := 0
	if ncPos+l >= ePos {
		for {
			sliceEnd := ncPos + l
			if sliceEnd > ePos+1 {
				sliceEnd = ePos + 1
				loops++
			}

			for i := 0; i < channels; i++ {
				res[i] = append(res[i], s[i][ncPos:sliceEnd]...)
				l = length - len(res[i])
			}

			ncPos = sliceEnd

			if l == 0 {
				return
			}

			ncPos = sPos

			if l < 0 || fors > 100 {
				logrus.Errorf("SHIT HAPPENED %d %d", l, fors)
				return
			}

			fors++
			continue
		}
	}

	res = s[ncPos : ncPos+length]

	return
}

func (jp *JamPlayer) OnServerConfigChange(bpm, bpi uint) {
	if jp.Playing() && jp.track != nil && jp.track.BPM != bpm && jp.track.BPI != bpi {
		jp.setBPM(jp.track.BPM)
		jp.setBPI(jp.track.BPI)
	}
}

package dj

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/ayvan/ninjam-dj-bot/tracks"
	"github.com/ayvan/ninjam-dj-bot/tts"
	"github.com/azul3d/engine/audio"
	"github.com/burillo-se/lv2host-go/lv2host"
	"github.com/burillo-se/lv2hostconfig"
	"github.com/burillo-se/ninjamencoder"
	"github.com/hajimehoshi/go-mp3"
	"github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"github.com/tosone/minimp3"
	"io"
	"math"
	"os"
	"path"
	"runtime/debug"
	"sync"
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
	track        *tracks.Track
	tracksPath   string
	source       io.Reader
	sampleRate   int
	bpm          uint
	bpi          uint
	repeats      uint
	ninjamBot    JamBot
	stop         chan bool
	playing      bool
	hostConfig   *lv2hostconfig.LV2HostConfig
	speechConfig *lv2hostconfig.LV2HostConfig
	onStopFunc   func()
	onStartFunc  func()
	bpmBPIOnSet  bool // set if bot called set bpm/bpi to ignore OnServerConfigChange callback
	voiceMtx     *sync.Mutex
}

type AudioInterval struct {
	GUID         [16]byte
	ChannelIndex uint8
	Flags        uint8
	Data         [][]byte
	index        int // index of current audio data block
}

// TODO получать сообщения о смене bpm/bpi и форсить их назад
func NewJamPlayer(tracksPath string, ninjamBot JamBot, lv2hostConfig, lv2speechConfig *lv2hostconfig.LV2HostConfig) *JamPlayer {
	return &JamPlayer{ninjamBot: ninjamBot, tracksPath: tracksPath, stop: make(chan bool, 1), hostConfig: lv2hostConfig, speechConfig: lv2speechConfig, voiceMtx: new(sync.Mutex)}
}

func (jp *JamPlayer) SetOnStart(f func()) {
	jp.onStartFunc = f
}

func (jp *JamPlayer) SetOnStop(f func()) {
	jp.onStopFunc = f
}

func (jp *JamPlayer) onStart() {
	if jp.onStartFunc != nil {
		jp.onStartFunc()
	}
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

func (jp *JamPlayer) LoadTrack(track *tracks.Track) error {
	jp.track = track
	filePath := track.FilePath

	if !path.IsAbs(filePath) {
		filePath = path.Join(jp.tracksPath, filePath)
	}
	logrus.Debugf("loading track %s", filePath)
	err := jp.setMP3Source(filePath)
	if err != nil {
		logrus.Error(err)
		return err
	}

	jp.SetRepeats(0) // по-умолчанию повторы не заданы, их должны будут задать отдельно если запуск происходит из плейлиста

	if jp.hostConfig == nil {
		err = fmt.Errorf("hostConfig not found")
		logrus.Error(err)
		return err
	}
	jp.hostConfig.ValueMap["integrated"] = track.Integrated
	jp.hostConfig.ValueMap["range"] = track.Range
	jp.hostConfig.ValueMap["peak"] = track.Peak
	jp.hostConfig.ValueMap["shortterm"] = track.Shortterm
	jp.hostConfig.ValueMap["momentary"] = track.Momentary

	return nil
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
		return fmt.Errorf("NewDecoder error in %s: %s", source, err)
	}

	jp.source = decoder

	jp.sampleRate = decoder.SampleRate()

	return nil
}

func (jp *JamPlayer) setBPM(bpm uint) {
	jp.bpmBPIOnSet = true
	defer func() {
		time.Sleep(time.Millisecond * 100)
		jp.bpmBPIOnSet = false
	}()
	logrus.Infof("setBPM %d", bpm)
	msg := fmt.Sprintf("bpm %d", bpm)
	jp.bpm = bpm
	jp.ninjamBot.SendAdminMessage(msg)
}

func (jp *JamPlayer) setBPI(bpi uint) {
	defer func() {
		time.Sleep(time.Millisecond * 100)
		jp.bpmBPIOnSet = false
	}()
	logrus.Infof("setBPI %d", bpi)
	msg := fmt.Sprintf("bpi %d", bpi)
	jp.bpi = bpi
	jp.ninjamBot.SendAdminMessage(msg)
}

func (jp *JamPlayer) Start() error {
	if jp.playing {
		return nil
	}
	if jp.source == nil || jp.track == nil {
		return fmt.Errorf("no source detected")
	}

	jp.playing = true

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

	samplesBuffer := make([][]float32, 2)

	waitData := make(chan bool, 1)
	errChan := make(chan error, 1)
	// это фоновая загрузка и декодирование MP3 в буфер
	go func() {
		defer func() {
			if r := recover(); r != nil {
				err := fmt.Errorf("panic: %s\n trace: %s", r, string(debug.Stack()))
				errChan <- err
			}
		}()
		intervalsReady := 0

		// initialize LV2 plugins
		host, err := jp.prepareLV2Host(float64(jp.sampleRate), jp.hostConfig)
		if err != nil {
			errChan <- err
			return
		}

		lv2host.Activate(host)
		defer func() {
			lv2host.Free(host)
		}()
		source := jp.source
		for {
			// на случай если кто-то уже остановил плеер или сменил источник
			if !jp.playing || jp.source != source {
				err := fmt.Errorf("no playing or source changed")
				errChan <- err
				return
			}
			buf := audio.Float32{}.Make(intervalSamplesChannels, intervalSamplesChannels)
			rs, err := toReadSeeker(source, intervalSamplesChannels)
			if err != nil && err != io.EOF && err.Error() != "end of stream" {
				err := fmt.Errorf("source.Read error: %s", err)
				errChan <- err
				return
			}

			var bufLen int
			bufLen, err = rs.Read(buf)
			if err != nil && err != io.EOF && err.Error() != "end of stream" {
				err := fmt.Errorf("source.Read error: %s", err)
				errChan <- err
				return
			}
			if bufLen == 0 {
				//bufferFull = true
				err := fmt.Errorf("error: bufLen == 0")
				errChan <- err
				return
			}

			deinterleavedSamples, err := ninjamencoder.DeinterleaveSamples(buf.(audio.Float32)[:bufLen], channels)
			if err != nil {
				err := fmt.Errorf("DeinterleaveSamples error: %s", err)
				errChan <- err
				return
			}

			lv2host.ProcessBuffer(host, deinterleavedSamples[0], deinterleavedSamples[1], uint32(len(deinterleavedSamples[0])))

			for i := 0; i < channels; i++ {
				samplesBuffer[i] = append(samplesBuffer[i], deinterleavedSamples[i]...)
			}

			intervalsReady++
			if intervalsReady == 3 {
				waitData <- true
			}
		}
	}()

	// ждём пока будут готовы интервалы или ошибку
	select {
	case <-waitData:
	case err := <-errChan:
		logrus.Error(err)
		jp.playing = false
		return err
	}
	jp.onStart()

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
	for jp.playing {
		time.Sleep(time.Millisecond * 500)
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
	logrus.Infof("Server change notify: BPM %d, BPI %d", bpm, bpi)
	if jp.Playing() && jp.track != nil && !jp.bpmBPIOnSet {
		if jp.bpm != bpm {
			jp.setBPM(jp.track.BPM)
		}
		if jp.bpi != bpi {
			jp.setBPI(jp.track.BPI)
		}
	}
}

func (jp *JamPlayer) PlayText(lang, text string) {
	go func() {
		defer recoverer()

		data, err := tts.Say(lang, text, false)
		if err != nil {
			logrus.Error(err)
			return
		}

		err = jp.playVoice(data)
		if err != nil {
			logrus.Error(err)
		}
	}()
}

func (jp *JamPlayer) playVoice(b []byte) (err error) {
	jp.voiceMtx.Lock()
	defer jp.voiceMtx.Unlock()

	dec, data, _ := minimp3.DecodeFull(b)

	rd := bytes.NewReader(data)

	buf := audio.Float32{}.Make(len(data), len(data))
	rs, err := toReadSeeker(rd, len(data))
	if err != nil && err != io.EOF && err.Error() != "end of stream" {
		err = fmt.Errorf("source.Read error: %s", err)
		return
	}

	var bufLen int
	bufLen, err = rs.Read(buf)
	if err != nil && err != io.EOF && err.Error() != "end of stream" {
		err = fmt.Errorf("source.Read error: %s", err)
		return
	}
	if bufLen == 0 {
		err = fmt.Errorf("error: bufLen == 0")
		return
	}

	if jp.speechConfig == nil {
		err = fmt.Errorf("speechConfig not found")
		logrus.Error(err)
		return err
	}

	// initialize LV2 plugins
	host, err := jp.prepareLV2Host(float64(dec.SampleRate), jp.speechConfig)
	if err != nil {
		return err
	}

	lv2host.Activate(host)
	defer lv2host.Free(host)

	deinterleavedSamples, err := ninjamencoder.DeinterleaveSamples(buf.(audio.Float32)[:bufLen], dec.Channels)
	if err != nil {
		err = fmt.Errorf("DeinterleaveSamples error: %s", err)
		return
	}

	// if only 1 channel - make it double for left and right channel
	if len(deinterleavedSamples) == 1 {
		deinterleavedSamples = append(deinterleavedSamples, deinterleavedSamples...)
	}

	lv2host.ProcessBuffer(host, deinterleavedSamples[0], deinterleavedSamples[1], uint32(len(deinterleavedSamples[0])))

	oggEncoder := ninjamencoder.NewEncoder()
	oggEncoder.SampleRate = dec.SampleRate

	oggData, err := oggEncoder.EncodeNinjamInterval(deinterleavedSamples)
	if err != nil {
		logrus.Errorf("EncodeNinjamInterval error: %s", err)
		return
	}

	guid, _ := uuid.NewV1()

	interval := AudioInterval{
		GUID:         guid,
		ChannelIndex: 1,
		Flags:        0,
		Data:         oggData,
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

	return nil
}

func (jp *JamPlayer) prepareLV2Host(sampleRate float64, config *lv2hostconfig.LV2HostConfig) (*lv2host.CLV2Host, error) {
	err := config.Evaluate()
	if err != nil {
		logrus.Fatal(err)
	}

	// initialize LV2 plugins
	host := lv2host.Alloc(sampleRate)

	for i, p := range config.Plugins {
		if lv2host.AddPluginInstance(host, p.PluginURI) != 0 {
			err := fmt.Errorf("Cannot add plugin: %v\n", p.PluginURI)
			return nil, err
		}
		logrus.Debugf("Add plugin %s'\n", p.PluginURI)

		for param, val := range p.Data {
			if lv2host.SetPluginParameter(host, uint32(i), param, val) != 0 {
				err := fmt.Errorf("Cannot set plugin parameter: %v\n", param)
				lv2host.ListPluginParameters(host, uint32(i))
				return nil, err
			}
			logrus.Debugf("Setting '%v' to '%v'\n", param, val)
		}
	}

	return host, nil
}

package dj

import (
	"encoding/binary"
	"fmt"
	"github.com/azul3d/engine/audio"
	"github.com/burillo-se/ninjamencoder"
	"github.com/hajimehoshi/go-mp3"
	"github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"io"
	"math"
	"os"
	"runtime/debug"
	"time"
)

// channels всегда 2 т.к. используемый MP3-декодер всегда отдаёт звук в стерео
const channels = 2

type IntervalBeginWriter interface {
	IntervalBegin(guid [16]byte, channelIndex uint8)
	IntervalWrite(guid [16]byte, data []byte, flags uint8)
}

type JamPlayer struct {
	source     io.Reader
	sampleRate int
	bpm        uint
	bpi        uint
	ninjamBot  IntervalBeginWriter
	stop       chan bool
	playing    bool
}

type AudioInterval struct {
	GUID         [16]byte
	ChannelIndex uint8
	Flags        uint8
	Data         [][]byte
	index        int // index of current audio data block
}

func NewJamPlayer(ninjamBot IntervalBeginWriter) *JamPlayer {
	return &JamPlayer{ninjamBot: ninjamBot, stop: make(chan bool, 1)}
}

func (jp *JamPlayer) Playing() bool {
	return jp.playing
}

func (jp *JamPlayer) SetBPM(bpm uint) {
	jp.bpm = bpm
}

func (jp *JamPlayer) SetBPI(bpi uint) {
	jp.bpi = bpi
}

func (jp *JamPlayer) SetMP3Source(source string) error {
	jp.Stop() // stop before set new source

	out, err := os.OpenFile(source, os.O_RDONLY, 0664)
	if err != nil {
		return fmt.Errorf("SetMP3Source error: %s", err)
	}

	decoder, err := mp3.NewDecoder(out)
	if err != nil {
		return fmt.Errorf("NewDecoder error: %s", err)
	}

	jp.source = decoder

	//jp.source, err = toReadSeeker(decoder)
	//if err != nil {
	//	err = fmt.Errorf("toReadSeeker error: %s", err)
	//	logrus.Error(err)
	//	fmt.Println(err)
	//}

	jp.sampleRate = decoder.SampleRate()

	return nil
}

func (jp *JamPlayer) Start() error {
	if jp.source == nil {
		fmt.Println("no source detected")
		return fmt.Errorf("no source detected")
	}

	jp.stop = make(chan bool, 1)

	repeats := 5
	_ = repeats

	// посчитаем на каких сэмплах у нас начало, и на каких конец зацикливания
	startTime := 1827878 * time.Microsecond
	loopStartPos := timeToSamples(startTime, jp.sampleRate)
	_ = loopStartPos

	endTime := 16373318 * time.Microsecond
	loopEndPos := timeToSamples(endTime, jp.sampleRate)
	_ = loopEndPos

	intervalTime := (float64(time.Minute) / float64(jp.bpm)) * float64(jp.bpi)
	intervalSamples := int(math.Ceil(float64(jp.sampleRate) * intervalTime / float64(time.Second)))
	intervalSamples2Channels := intervalSamples * channels

	jp.playing = true

	samplesBuffer := make([][]float32, 2)

	// эта переменная будет установлена когда буфер будет заполнен всеми данными из MP3 файла
	bufferFull := false

	waitData := make(chan bool, 1)
	// это фоновая загрузка и декодирование MP3 в буфер
	go func() {
		intervalsReady := 0

		for {
			buf := audio.Float32{}.Make(intervalSamples2Channels, intervalSamples2Channels)
			rs, err := toReadSeeker(jp.source, intervalSamples2Channels)
			if err != nil && err != io.EOF && err.Error() != "end of stream" {
				logrus.Errorf("source.Read error: %s", err)
			}

			var n int
			n, err = rs.Read(buf)
			if err != nil && err != io.EOF && err.Error() != "end of stream" {
				logrus.Errorf("source.Read error: %s", err)
			}
			if n == 0 {
				bufferFull = true
				return
			}

			deinterleavedSamples, err := ninjamencoder.DeinterleaveSamples(buf.(audio.Float32), channels)
			if err != nil {
				logrus.Errorf("DeinterleaveSamples error: %s", err)
				return
			}

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

			jp.playing = false
		}()

		ticker := time.NewTicker(time.Duration(intervalTime))

		oggEncoder := ninjamencoder.NewEncoder()
		oggEncoder.SampleRate = jp.sampleRate

		play := true
		currentPos := 0

		loops := 1000
		_ = loops

		for play {
			deinterleavedSamples := make([][]float32, 2)
			endPos := currentPos + intervalSamples
			fmt.Println(loopStartPos, loopEndPos, currentPos, endPos)
			if endPos > len(samplesBuffer[0]) {
				endPos = len(samplesBuffer[0])
				play = false // дошли до конца - завершаем
			}

			if currentPos >= loopStartPos && endPos >= loopEndPos && loops > 0 {
				play = true // если ранее получили флаг остановки - значит снимем его, мы ушли в очередной цикл
				samplesToIntervalEnd := endPos - loopEndPos
				endPos = loopEndPos

				for i := 0; i < channels; i++ {
					// создаём новый слайс и копируем в него, т.к. дальше нам нужно с ним работать отдельно от кэшированного слайса - мы будем его менять через append
					deinterleavedSamples[i] = make([]float32, endPos-currentPos, intervalSamples)
					copy(deinterleavedSamples[i], samplesBuffer[i][currentPos:endPos])
					deinterleavedSamples[i] = append(deinterleavedSamples[i], samplesBuffer[i][loopStartPos:loopStartPos+samplesToIntervalEnd]...)
				}

				currentPos = loopStartPos + samplesToIntervalEnd

				loops--
				logrus.Debugf("loops left: %d", loops)
			} else {
				for i := 0; i < channels; i++ {
					deinterleavedSamples[i] = samplesBuffer[i][currentPos:endPos]
				}

				currentPos = endPos
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

				intervalData, hasNext = interval.Next()

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

func (ai *AudioInterval) Next() (data []byte, hasNext bool) {
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
	buf := audio.NewBuffer(audio.Float32{})
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
	return int(math.Ceil(float64(sampleRate) * float64(t) / float64(time.Second)))

}

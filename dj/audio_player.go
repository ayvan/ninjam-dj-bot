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
	"time"
)

// channels всегда 2 т.к. используемый MP3-декодер всегда отдаёт звук в стерео
const channels = 2

type IntervalBeginWriter interface {
	IntervalBegin(guid [16]byte, channelIndex uint8)
	IntervalWrite(guid [16]byte, data []byte, flags uint8)
}

type JamPlayer struct {
	source     audio.ReadSeeker
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

	jp.source, err = toReadSeeker(decoder)
	if err != nil {
		err = fmt.Errorf("toReadSeeker error: %s", err)
		logrus.Error(err)
		fmt.Println(err)
	}

	jp.sampleRate = decoder.SampleRate()

	return nil
}

func (jp *JamPlayer) Start() error {
	if jp.source == nil {
		fmt.Println("no source detected")
		return fmt.Errorf("no source detected")
	}

	jp.stop = make(chan bool, 1)

	intervalTime := (float64(time.Minute) / float64(jp.bpm)) * float64(jp.bpi)
	samples := int(float64(jp.sampleRate)*intervalTime/float64(time.Second)) * channels

	jp.playing = true

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logrus.Errorf("panic in JamPlayer.Start: %s", r)
			}

			jp.playing = false
		}()

		ticker := time.NewTicker(time.Duration(intervalTime))

		oggEncoder := ninjamencoder.NewEncoder()
		oggEncoder.SampleRate = jp.sampleRate

		for {
			buf := audio.Float32{}.Make(samples, samples)
			var n int
			n, err := jp.source.Read(buf)
			if err != nil && err != io.EOF {
				logrus.Errorf("source.Read error: %s", err)
			}
			if n == 0 {
				return
			}

			deinterleavedSamples, err := ninjamencoder.DeinterleaveSamples(buf.(audio.Float32), channels)
			if err != nil {
				logrus.Errorf("DeinterleaveSamples error: %s", err)
			}

			data, err := oggEncoder.EncodeNinjamInterval(deinterleavedSamples)
			if err != nil {
				logrus.Errorf("EncodeNinjamInterval error: %s", err)
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

				fmt.Print("|", len(intervalData))

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
	if len(ai.Data) <= ai.index+1 {
		hasNext = false
	}

	return
}

func toReadSeeker(reader io.Reader) (res audio.ReadSeeker, err error) {
	buf := audio.NewBuffer(audio.Float32{})
	res = buf

	for {
		data := make([]byte, 2, 2)
		var n int
		n, err = reader.Read(data)
		if err != nil && err != io.EOF {
			return nil, err
		}
		if n == 0 {
			err = nil // remove EOF error
			return
		}

		uintData := binary.LittleEndian.Uint16(data)
		buf.Write(audio.Float32{Uint16ToFloat32(uintData)})
	}

	return
}

func Uint16ToFloat32(s uint16) float32 {
	return float32(s)/float32(math.MaxInt16) - 1
}

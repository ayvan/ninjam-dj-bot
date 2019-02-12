package tracks_sync

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type TrackDater interface {
	Unmarshal([]byte) error
	Marshal() []byte
	Key() uint
	Mode() uint
	BPM() uint
	BPI() uint
	LoopStart() uint64
	LoopEnd() uint64

	SetKey(uint)
	SetMode(uint)
	SetBPM(uint)
	SetBPI(uint)
	SetLoopStart(uint64)
	SetLoopEnd(uint64)
}

type private_ext_frame_data struct {
	magic   uint64 //8
	version uint16 //2
	len     uint16 //2
	data    TrackDater
}

type private_ext_frame_data_v2 struct {
	key  int32  //4
	mode int32  //4
	ls   int64  //8
	le   int64  //8
	bpm  uint32 //4
	bpi  uint32 //4
}

type private_ext_frame_data_v3 struct {
	ls   uint64 // loop start, in microseconds
	le   uint64 // loop end, in microseconds
	key  uint32 // enum Key
	mode uint32 // enum Mode
	bpm  uint32
	bpi  uint32
}

func getFrameNameAndData(raw []byte) (name, data []byte) {
	i := bytes.Index(raw, []byte{0})
	i++
	return raw[:i-1], raw[i:]
}

func (p *private_ext_frame_data) Unmarshal(data []byte) error {
	if len(data) < 12 {
		return fmt.Errorf("too short data frame")
	}

	p.magic = binary.LittleEndian.Uint64(data[:8])
	p.version = binary.LittleEndian.Uint16(data[8:10])
	p.len = binary.LittleEndian.Uint16(data[10:12])

	switch p.version {
	case 2:
		p.data = new(private_ext_frame_data_v2)
	case 3:
		p.data = new(private_ext_frame_data_v3)
	default:
		return fmt.Errorf("wrong version: %d", p.version)
	}

	p.data.Unmarshal(data[12:])

	return nil
}

func (p *private_ext_frame_data) Marshal() (data []byte, err error) {
	data = make([]byte, 12)

	switch p.data.(type) {
	case *private_ext_frame_data_v2:
		p.version = 2
	case *private_ext_frame_data_v3:
		p.version = 3
	default:
		return nil, fmt.Errorf("wrong p.data.(type): %T", p.data)
	}

	binary.LittleEndian.PutUint64(data[:8], p.magic)
	binary.LittleEndian.PutUint16(data[8:10], p.version)
	binary.LittleEndian.PutUint16(data[10:12], p.len)

	data = append(data, p.data.Marshal()...)

	return
}

func (p *private_ext_frame_data) checkVersion() error {
	if p.version != 2 && p.version != 3 {
		return fmt.Errorf("bad tag version: %d", p.version)
	}

	return nil
}

func (p *private_ext_frame_data_v2) Unmarshal(data []byte) error {
	if len(data) < 32 {
		return fmt.Errorf("too short data frame")
	}
	p.key = int32(binary.LittleEndian.Uint32(data[:4]))
	p.mode = int32(binary.LittleEndian.Uint32(data[4:8]))
	p.ls = int64(binary.LittleEndian.Uint64(data[8:16]))
	p.le = int64(binary.LittleEndian.Uint64(data[16:24]))
	p.bpm = binary.LittleEndian.Uint32(data[24:28])
	p.bpi = binary.LittleEndian.Uint32(data[28:32])

	return nil
}

func (p *private_ext_frame_data_v2) Marshal() (data []byte) {

	data = make([]byte, 32)

	binary.LittleEndian.PutUint32(data[:4], uint32(p.key))
	binary.LittleEndian.PutUint32(data[4:8], uint32(p.mode))

	binary.LittleEndian.PutUint64(data[8:16], uint64(p.ls))
	binary.LittleEndian.PutUint64(data[16:24], uint64(p.le))
	binary.LittleEndian.PutUint32(data[24:28], uint32(p.bpm))
	binary.LittleEndian.PutUint32(data[28:32], uint32(p.bpi))

	return
}

func (p *private_ext_frame_data_v3) Unmarshal(data []byte) error {
	if len(data) < 32 {
		return fmt.Errorf("too short data frame")
	}
	p.ls = binary.LittleEndian.Uint64(data[:8])
	p.le = binary.LittleEndian.Uint64(data[8:16])
	p.key = binary.LittleEndian.Uint32(data[16:20])
	p.mode = binary.LittleEndian.Uint32(data[20:24])
	p.bpm = binary.LittleEndian.Uint32(data[24:28])
	p.bpi = binary.LittleEndian.Uint32(data[28:32])

	return nil
}

func (p *private_ext_frame_data_v3) Marshal() (data []byte) {

	data = make([]byte, 32)

	binary.LittleEndian.PutUint64(data[:8], p.ls)
	binary.LittleEndian.PutUint64(data[8:16], p.le)

	binary.LittleEndian.PutUint32(data[16:20], p.key)
	binary.LittleEndian.PutUint32(data[20:24], p.mode)
	binary.LittleEndian.PutUint32(data[24:28], p.bpm)
	binary.LittleEndian.PutUint32(data[28:32], p.bpi)

	return
}

func (p *private_ext_frame_data_v2) Key() uint {
	return uint(p.key)
}
func (p *private_ext_frame_data_v2) Mode() uint {
	return uint(p.mode)
}
func (p *private_ext_frame_data_v2) BPM() uint {
	return uint(p.bpm)
}
func (p *private_ext_frame_data_v2) BPI() uint {
	return uint(p.bpi)
}
func (p *private_ext_frame_data_v2) LoopStart() uint64 {
	return uint64(p.ls)
}
func (p *private_ext_frame_data_v2) LoopEnd() uint64 {
	return uint64(p.le)
}

func (p *private_ext_frame_data_v2) SetKey(v uint) {
	p.key = int32(v)
}
func (p *private_ext_frame_data_v2) SetMode(v uint) {
	p.mode = int32(v)
}
func (p *private_ext_frame_data_v2) SetBPM(v uint) {
	p.bpm = uint32(v)
}
func (p *private_ext_frame_data_v2) SetBPI(v uint) {
	p.bpi = uint32(v)
}
func (p *private_ext_frame_data_v2) SetLoopStart(v uint64) {
	p.ls = int64(v)
}
func (p *private_ext_frame_data_v2) SetLoopEnd(v uint64) {
	p.le = int64(v)
}

func (p *private_ext_frame_data_v3) Key() uint {
	return uint(p.key)
}
func (p *private_ext_frame_data_v3) Mode() uint {
	return uint(p.mode)
}
func (p *private_ext_frame_data_v3) BPM() uint {
	return uint(p.bpm)
}
func (p *private_ext_frame_data_v3) BPI() uint {
	return uint(p.bpi)
}
func (p *private_ext_frame_data_v3) LoopStart() uint64 {
	return uint64(p.ls)
}
func (p *private_ext_frame_data_v3) LoopEnd() uint64 {
	return uint64(p.le)
}

func (p *private_ext_frame_data_v3) SetKey(v uint) {
	p.key = uint32(v)
}
func (p *private_ext_frame_data_v3) SetMode(v uint) {
	p.mode = uint32(v)
}
func (p *private_ext_frame_data_v3) SetBPM(v uint) {
	p.bpm = uint32(v)
}
func (p *private_ext_frame_data_v3) SetBPI(v uint) {
	p.bpi = uint32(v)
}
func (p *private_ext_frame_data_v3) SetLoopStart(v uint64) {
	p.ls = v
}
func (p *private_ext_frame_data_v3) SetLoopEnd(v uint64) {
	p.le = v
}

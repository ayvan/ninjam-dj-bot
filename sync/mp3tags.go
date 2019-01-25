package sync

import (
	"fmt"
	"bytes"
	"encoding/binary"
)

type private_ext_frame_data struct {
	magic   uint64 //8
	version int16  //2
	len     int16  //2
	data    private_ext_frame_data_v2
}

type private_ext_frame_data_v2 struct {
	key  int32  //4
	mode int32  //4
	ls   int64  //8
	le   int64  //8
	bpm  uint32 //4
	bpi  uint32 //4
}

func getFrameNameAndData(raw []byte) (name, data []byte) {
	i := bytes.Index(raw, []byte{0})
	i++
	return raw[:i], raw[i:]
}

func (p *private_ext_frame_data) Unmarshal(data []byte) error {
	if len(data) < 12 {
		return fmt.Errorf("too short data frame")
	}

	p.magic = binary.LittleEndian.Uint64(data[:8])
	p.version = int16(binary.LittleEndian.Uint16(data[8:10]))
	p.len = int16(binary.LittleEndian.Uint16(data[10:12]))

	p.data.Unmarshal(data[12:])

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

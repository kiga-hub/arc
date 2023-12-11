package protocols

import (
	"encoding/binary"
	"fmt"

	"github.com/kiga-hub/common/utils"
)

const (
	// AudioSampleRate4k -
	AudioSampleRate4k uint16 = 4000
	// AudioSampleRate8k -
	AudioSampleRate8k uint16 = 8000
	// AudioSampleRate16k -
	AudioSampleRate16k uint16 = 16000
	// AudioSampleRate32k -
	AudioSampleRate32k uint16 = 32000
	// AudioSampleRate48k -
	AudioSampleRate48k uint16 = 48000
)

// SegmentAudio  Audio
type SegmentAudio struct {
	SType      byte   //1 数据类型
	SampleRate uint16 //2 采样率
	Bits       byte   //1 位深度
	Channel    byte   //1 通道
	Data       []byte //n 数据
}

// NewDefaultSegmentAudio -
func NewDefaultSegmentAudio() *SegmentAudio {
	return &SegmentAudio{
		SType:      STypeAudio,
		SampleRate: AudioSampleRate32k,
		Bits:       16,
		Channel:    1,
	}
}

// AudioValidate -
func AudioValidate(srcdata []byte) error {

	data := make([]byte, len(srcdata))
	copy(data, srcdata)

	idx := 0
	// check stype(1)
	if data[idx] != STypeAudio {
		return fmt.Errorf("audio segment stype invalid(%d)", data[idx])
	}
	idx++

	// samplerate(2)
	SampleRate := binary.BigEndian.Uint16(data[idx : idx+2])
	if SampleRate != AudioSampleRate8k &&
		SampleRate != AudioSampleRate4k &&
		SampleRate != AudioSampleRate16k &&
		SampleRate != AudioSampleRate32k &&
		SampleRate != AudioSampleRate48k {
		return fmt.Errorf("audio segment smplerate invalid(%d)", SampleRate)
	}
	idx += 2

	// bits(1)
	Bits := data[idx]
	if Bits != 8 &&
		Bits != 16 &&
		Bits != 32 {
		return fmt.Errorf("audio segment bit invalid(%d)", Bits)
	}
	idx++

	// channel(1)
	Channel := data[idx]
	if Channel <= 0 || Channel > 8 {
		return fmt.Errorf("audio segment channel invalid(%d)", Channel)
	}
	return nil
}

// Decode -
func (s *SegmentAudio) Decode(srcdata []byte) error {

	data := make([]byte, len(srcdata))
	copy(data, srcdata)

	idx := 0
	// stype(1)
	s.SType = data[idx]
	idx++
	// samplerate(2)
	s.SampleRate = binary.BigEndian.Uint16(data[idx : idx+2])
	idx += 2
	// bits(1)
	s.Bits = data[idx]
	idx++
	// channel(1)
	s.Channel = data[idx]
	idx++
	// data
	s.Data = data[idx:]
	return s.Validate()
}

// Validate -
func (s *SegmentAudio) Validate() error {
	if s.SType != STypeAudio {
		return fmt.Errorf("audio type %d", s.SType)
	}
	if s.SampleRate != AudioSampleRate8k &&
		s.SampleRate != AudioSampleRate4k &&
		s.SampleRate != AudioSampleRate16k &&
		s.SampleRate != AudioSampleRate32k &&
		s.SampleRate != AudioSampleRate48k {
		return fmt.Errorf("audio smplerate %d", s.SampleRate)
	}
	if s.Bits != 8 &&
		s.Bits != 16 &&
		s.Bits != 32 {
		return fmt.Errorf("audio bit %d", s.Bits)
	}
	if s.Channel <= 0 || s.Channel > 8 {
		return fmt.Errorf("audio channel %d", s.Channel)
	}
	return nil
}

// Encode - 编码
func (s *SegmentAudio) Encode(buf []byte) (int, error) {
	if len(buf) < len(s.Data)+5 {
		return 0, fmt.Errorf("out of allocated memory")
	}
	idx := 0

	// stype(1)
	buf[idx] = s.SType
	idx++
	// samplerate(2)
	binary.BigEndian.PutUint16(buf[idx:idx+2], s.SampleRate)
	idx += 2
	// bits(1)
	buf[idx] = s.Bits
	idx++
	// channel(1)
	buf[idx] = s.Channel
	idx++

	copy(buf[idx:], s.Data)
	idx += len(s.Data)
	return idx, nil
}

// Type - 段类型
func (s *SegmentAudio) Type() byte {
	return s.SType
}

// Size - 编码大小
func (s *SegmentAudio) Size() uint32 {
	return uint32(len(s.Data) + 5)
}

// Dump -
func (s *SegmentAudio) Dump() {
	title := fmt.Sprintf("  Dump Audio Samplerate: %d,Bit: %d,Channel: %d\n  ", s.SampleRate, s.Bits, s.Channel)
	utils.Hexdump(title, s.Data)
}

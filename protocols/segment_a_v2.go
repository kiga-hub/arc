package protocols

import (
	"encoding/binary"
	"fmt"

	"common/utils"
)

const (
	// SegmentAudioV2OffsetBytes 获取Data，需要偏移字节数
	SegmentAudioV2OffsetBytes = 6
)

// SegmentAudioV2  音频新版本数据块
type SegmentAudioV2 struct {
	SType      byte   //1 数据类型
	SampleRate uint16 //2 采样率
	Bits       byte   //1 位深度
	Channel    byte   //1 通道
	Gain       uint8  //1 增益
	Data       []byte //n 数据
}

// NewDefaultSegmentAudioV2 -
func NewDefaultSegmentAudioV2() *SegmentAudioV2 {
	return &SegmentAudioV2{
		SType:      STypeAudioV2,
		SampleRate: AudioSampleRate32k,
		Bits:       16,
		Channel:    1,
	}
}

// AudioV2Validate -
func AudioV2Validate(srcdata []byte) error {

	data := make([]byte, len(srcdata))
	copy(data, srcdata)

	idx := 0
	// check stype(1)
	if data[idx] != STypeAudioV2 {
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
func (s *SegmentAudioV2) Decode(srcdata []byte) error {

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
	// gain(1)
	s.Gain = data[idx]
	idx++
	// data
	s.Data = data[idx:]
	return s.Validate()
}

// Validate -
func (s *SegmentAudioV2) Validate() error {
	if s.SType != STypeAudioV2 {
		return fmt.Errorf("audioV2 type %d", s.SType)
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
func (s *SegmentAudioV2) Encode(buf []byte) (int, error) {
	if len(buf) < len(s.Data)+SegmentAudioV2OffsetBytes {
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
	// gain(1)
	buf[idx] = s.Gain
	idx++

	copy(buf[idx:], s.Data)
	idx += len(s.Data)
	return idx, nil
}

// Type - 段类型
func (s *SegmentAudioV2) Type() byte {
	return s.SType
}

// Size - 编码大小
func (s *SegmentAudioV2) Size() uint32 {
	return uint32(len(s.Data) + SegmentAudioV2OffsetBytes)
}

// Dump -
func (s *SegmentAudioV2) Dump() {
	title := fmt.Sprintf("  Dump AudioV2 Samplerate: %d,Bit: %d,Channel: %d\n  ", s.SampleRate, s.Bits, s.Channel)
	utils.Hexdump(title, s.Data)
}

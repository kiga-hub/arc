package protocols

import (
	"encoding/binary"
	"fmt"

	"github.com/kiga-hub/arc/utils"
)

const (
	// VibrateSampleRate3200 -
	VibrateSampleRate3200 = iota
	// VibrateSampleRate1600 -
	VibrateSampleRate1600
	// VibrateSampleRate800 -
	VibrateSampleRate800
	// VibrateSampleRate400 -
	VibrateSampleRate400
	// VibrateSampleRate200 -
	VibrateSampleRate200
	// VibrateSampleRate100 -
	VibrateSampleRate100
	// VibrateSampleRate50 -
	VibrateSampleRate50
	// VibrateSampleRate25 -
	VibrateSampleRate25
	// VibrateSampleRate125 -
	VibrateSampleRate125
	// VibrateSampleRate625 -
	VibrateSampleRate625
	// VibrateSampleRate313 -
	VibrateSampleRate313
	// VibrateSampleRate156 -
	VibrateSampleRate156
	// VibrateSampleRate078 -
	VibrateSampleRate078
	// VibrateSampleRate039 -
	VibrateSampleRate039
	// VibrateSampleRate020 -
	VibrateSampleRate020
	// VibrateSampleRate01 -
	VibrateSampleRate01
)

const (
	// VibrateDynamicRange2 -
	VibrateDynamicRange2 byte = 2
	// VibrateDynamicRange4 -
	VibrateDynamicRange4 byte = 4
	// VibrateDynamicRange8 -
	VibrateDynamicRange8 byte = 8
	// VibrateDynamicRange16 -
	VibrateDynamicRange16 byte = 16
)

const (
	// VibrateResolution10 -
	VibrateResolution10 byte = 10
	// VibrateResolution11 -
	VibrateResolution11 byte = 11
	// VibrateResolution12 -
	VibrateResolution12 byte = 12
	// VibrateResolution13 -
	VibrateResolution13 byte = 13
)

// Vibrate -
type Vibrate struct {
	X int16
	Y int16
	Z int16
}

// SegmentVibrate Vibrate   X-Y-Z  三个值
type SegmentVibrate struct {
	SType        byte      //1 数据类型s
	SampleRate   uint16    //2 采样率  0.1HZ~3200HZ
	DynamicRange byte      //1 量程范围  +- 2 4,8,16
	Resolution   byte      //1 分辨率  10 11 12 13
	Offset       Vibrate   //6 偏移量
	Data         []Vibrate //n 数据
	Count        int       // 震动数量
	Bytes        []byte
}

// Validate -
func (s *SegmentVibrate) Validate() error {
	if s.SType != STypeVibrate {
		return fmt.Errorf("vibrate type %d", s.SType)
	}
	if s.SampleRate > 15 {
		return fmt.Errorf("vibrate samplerate %d", s.SampleRate)
	}
	if s.DynamicRange != VibrateDynamicRange2 &&
		s.DynamicRange != VibrateDynamicRange4 &&
		s.DynamicRange != VibrateDynamicRange8 &&
		s.DynamicRange != VibrateDynamicRange16 {
		return fmt.Errorf("vibrate Dynamic range %d", s.DynamicRange)
	}

	if s.Resolution != VibrateResolution10 &&
		s.Resolution != VibrateResolution11 &&
		s.Resolution != VibrateResolution12 &&
		s.Resolution != VibrateResolution13 {
		return fmt.Errorf("vibrate resolution %d", s.Resolution)
	}
	return nil
}

// NewDefaultSegmentVibrate -
func NewDefaultSegmentVibrate() *SegmentVibrate {
	return &SegmentVibrate{
		SType:        STypeVibrate,
		SampleRate:   VibrateSampleRate3200,
		DynamicRange: VibrateDynamicRange2,
		Resolution:   VibrateResolution10,
	}
}

// SetOffset - 偏移
func (s *SegmentVibrate) SetOffset(x, y, z int16) {
	s.Offset.X = x
	s.Offset.Y = y
	s.Offset.Z = z
}

// AppendData - 追加数据
func (s *SegmentVibrate) AppendData(x, y, z int16) {
	s.Data = append(s.Data, Vibrate{
		X: x,
		Y: y,
		Z: z,
	})
	if len(s.Bytes) < (s.Count+1)*6 {
		buf := make([]byte, (s.Count+1)*6-len(s.Bytes))
		s.Bytes = append(s.Bytes, buf...)
	}
	idx := s.Count * 6
	binary.BigEndian.PutUint16(s.Bytes[idx:idx+2], uint16(x))
	binary.BigEndian.PutUint16(s.Bytes[idx+2:idx+4], uint16(y))
	binary.BigEndian.PutUint16(s.Bytes[idx+4:idx+6], uint16(z))
	s.Count++
}

// Encode - 编码
func (s *SegmentVibrate) Encode(buf []byte) (int, error) {
	if len(buf) < len(s.Data)*6+11 {
		return 0, fmt.Errorf("out of allocated memory")
	}
	idx := 0

	// stype(1)
	buf[idx] = s.SType
	idx++
	// samplerate(2)
	binary.BigEndian.PutUint16(buf[idx:idx+2], s.SampleRate)
	idx += 2
	// dynamicrange(1)
	buf[idx] = s.DynamicRange
	idx++
	// resolution(1)
	buf[idx] = s.Resolution
	idx++
	// offset(6)
	binary.BigEndian.PutUint16(buf[idx:idx+2], uint16(s.Offset.X))
	idx += 2
	binary.BigEndian.PutUint16(buf[idx:idx+2], uint16(s.Offset.Y))
	idx += 2
	binary.BigEndian.PutUint16(buf[idx:idx+2], uint16(s.Offset.Z))
	idx += 2

	copy(buf[idx:], s.Bytes[:s.Count*6])
	idx += s.Count * 6
	return idx, nil
}

// Type - 段类型
func (s *SegmentVibrate) Type() byte {
	return s.SType
}

// Size - 编码大小
func (s *SegmentVibrate) Size() uint32 {
	return uint32(s.Count*6 + 11)
}

// GetData - 获取数据
func (s *SegmentVibrate) GetData() []VibrateValues {
	revalue := []VibrateValues{}
	for i := 0; i < s.Count; i++ {
		v := VibrateValues{Ts: 0, X: s.Data[i].X, Y: s.Data[i].Y, Z: s.Data[i].Z}
		revalue = append(revalue, v)
	}
	return revalue
}

// VibrateValidate - 解码
func VibrateValidate(srcdata []byte) error {

	data := make([]byte, len(srcdata))
	copy(data, srcdata)

	idx := 0
	// stype(1)
	if data[idx] != STypeVibrate {
		return fmt.Errorf("vibrate segment stype invalid(%d)", data[idx])
	}
	idx++

	// samplerate(2)
	SampleRate := binary.BigEndian.Uint16(data[idx : idx+2])
	if SampleRate > 15 {
		return fmt.Errorf("vibrate segment samplerate invalid(%d)", SampleRate)
	}
	idx += 2

	// dynamicrange(1)
	if data[idx] != VibrateDynamicRange2 &&
		data[idx] != VibrateDynamicRange4 &&
		data[idx] != VibrateDynamicRange8 &&
		data[idx] != VibrateDynamicRange16 {
		return fmt.Errorf("vibrate segment Dynamic range invalid(%d)", data[idx])
	}
	idx++

	// resolution(1)
	if data[idx] != VibrateResolution10 &&
		data[idx] != VibrateResolution11 &&
		data[idx] != VibrateResolution12 &&
		data[idx] != VibrateResolution13 {
		return fmt.Errorf("vibrate segment resolution invalid(%d)", data[idx])
	}
	idx++

	// offset_x(2)
	// s.Offset.X = int16(binary.BigEndian.Uint16(data[idx : idx+2]))
	// TODO: check offset
	idx += 2

	// offset_y(2)
	// s.Offset.Y = int16(binary.BigEndian.Uint16(data[idx : idx+2]))
	// TODO: check offset
	idx += 2

	// offset_Z(2)
	// s.Offset.Z = int16(binary.BigEndian.Uint16(data[idx : idx+2]))
	// TODO: check offset
	idx += 2

	if (len(data)-idx)%6 != 0 {
		return fmt.Errorf("vibrate segment data invalid(%d)", len(data)-idx)
	}
	return nil
}

// Decode - 解码
func (s *SegmentVibrate) Decode(srcdata []byte) error {

	data := make([]byte, len(srcdata))
	copy(data, srcdata)

	idx := 0
	// stype(1)
	s.SType = data[idx]
	idx++
	// samplerate(2)
	s.SampleRate = binary.BigEndian.Uint16(data[idx : idx+2])
	idx += 2
	// dynamicrange(1)
	s.DynamicRange = data[idx]
	idx++
	// resolution(1)
	s.Resolution = data[idx]
	idx++
	// offset_x(2)
	s.Offset.X = int16(binary.BigEndian.Uint16(data[idx : idx+2]))
	idx += 2
	// offset_y(2)
	s.Offset.Y = int16(binary.BigEndian.Uint16(data[idx : idx+2]))
	idx += 2
	// offset_Z(2)
	s.Offset.Z = int16(binary.BigEndian.Uint16(data[idx : idx+2]))
	idx += 2

	s.Bytes = data[idx:]
	s.Count = (len(data) - idx) / 6
	for i := 0; i < s.Count; i++ {
		if i >= len(s.Data) {
			s.Data = append(s.Data, Vibrate{})
		}
		s.Data[i].X = int16(binary.BigEndian.Uint16(data[idx : idx+2]))
		data[idx], data[idx+1] = data[idx+1], data[idx]
		idx += 2

		s.Data[i].Y = int16(binary.BigEndian.Uint16(data[idx : idx+2]))
		data[idx], data[idx+1] = data[idx+1], data[idx]
		idx += 2

		s.Data[i].Z = int16(binary.BigEndian.Uint16(data[idx : idx+2]))
		data[idx], data[idx+1] = data[idx+1], data[idx]
		idx += 2
	}
	return s.Validate()
}

// Dump -
func (s *SegmentVibrate) Dump() {
	title := fmt.Sprintf(" Dump Vibrate Samplerate: %d, Range: %d, Resolution: %d, Offset: %d/%d/%d\n  ",
		s.SampleRate, s.DynamicRange, s.Resolution,
		s.Offset.X, s.Offset.Y, s.Offset.Z,
	)
	utils.Hexdump(title, s.Bytes)
}

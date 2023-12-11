package protocols

import (
	"encoding/binary"
	"fmt"

	"github.com/kiga-hub/common/utils"
)

// SegmentTemperature  Temperature
type SegmentTemperature struct {
	SType byte  //1 数据类型
	Data  int16 //2 数据
	Bytes []byte
}

// Validate -
func (s *SegmentTemperature) Validate() error {
	if s.SType != STypeTemperature {
		return fmt.Errorf("temperature type %d", s.SType)
	}
	return nil
}

// NewDefaultSegmentTemperature -
func NewDefaultSegmentTemperature() *SegmentTemperature {
	return &SegmentTemperature{
		SType: STypeTemperature,
	}
}

// TemperatureValidate - 解码
func TemperatureValidate(srcdata []byte) error {
	data := make([]byte, len(srcdata))
	copy(data, srcdata)

	idx := 0

	// stype(1)
	if data[idx] != STypeTemperature {
		return fmt.Errorf("temperature segment stype invalid(%d)", data[idx])
	}

	// data(2)
	if len(data) != 3 {
		return fmt.Errorf("temperature segment data invalid(%d)", len(data)-1)
	}

	return nil
}

// Decode - 解码
func (s *SegmentTemperature) Decode(srcdata []byte) error {

	data := make([]byte, len(srcdata))
	copy(data, srcdata)

	idx := 0
	// stype(1)
	s.SType = data[idx]
	idx++
	// data(2)
	s.Data = int16(binary.BigEndian.Uint16(data[idx:]))
	s.Bytes = data[idx:]
	return s.Validate()
}

// SetData - 获取数据
func (s *SegmentTemperature) SetData(t float32) {
	t = t * 16.0
	s.Data = int16(t)
	if len(s.Bytes) < 2 {
		s.Bytes = make([]byte, 2)
	}
	binary.BigEndian.PutUint16(s.Bytes[:2], uint16(s.Data))
}

// GetData - 设置数据
func (s *SegmentTemperature) GetData() float32 {
	return float32(s.Data) / 16.0
}

// Encode - 编码
func (s *SegmentTemperature) Encode(buf []byte) (int, error) {
	if len(buf) < 3 {
		return 0, fmt.Errorf("out of allocated memory")
	}
	idx := 0

	// stype(1)
	buf[idx] = s.SType
	idx++

	// data(2)
	copy(buf[idx:idx+2], s.Bytes[:2])
	idx += 2

	return idx, nil
}

// Type - 段类型
func (s *SegmentTemperature) Type() byte {
	return s.SType
}

// Size - 编码大小
func (s *SegmentTemperature) Size() uint32 {
	return 2 + 1
}

// Dump -
func (s *SegmentTemperature) Dump() {
	title := fmt.Sprintf("  Dump Temperature: %.3f\n  ", s.GetData())
	utils.Hexdump(title, s.Bytes)
}

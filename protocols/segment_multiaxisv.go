package protocols

import (
	"encoding/binary"
	"fmt"

	"github.com/kiga-hub/arc/utils"
)

// SegmentMultiAxisVibrate Vibrate
type SegmentMultiAxisVibrate struct {
	SType        byte    //1 数据类型s
	AxleNum      byte    //1 轴数
	SampleRate   uint16  //2 采样率  单位HZ， 默认20000HZ
	DynamicRange uint16  //2 量程范围  单位g +-200g
	Resolution   uint16  //2 分辨率  mv/g
	Data         []int16 //n 数据
	Count        int     // 震动数量
	Bytes        []byte
}

// Validate -
func (s *SegmentMultiAxisVibrate) Validate() error {
	if s.SType != STypeMultiAxisVibrate {
		return fmt.Errorf("vibrate type %d", s.SType)
	}
	return nil
}

// NewDefaultSegmentMultiAxisVibrate -
func NewDefaultSegmentMultiAxisVibrate() *SegmentMultiAxisVibrate {
	return &SegmentMultiAxisVibrate{
		SType:        STypeMultiAxisVibrate,
		AxleNum:      1,
		SampleRate:   20000,
		DynamicRange: 200,
		Resolution:   0,
	}
}

// Type - 段类型
func (s *SegmentMultiAxisVibrate) Type() byte {
	return s.SType
}

// Size - 编码大小
func (s *SegmentMultiAxisVibrate) Size() uint32 {
	return uint32(s.Count*int(s.AxleNum)*2 + 8)
}

// Encode - 编码
func (s *SegmentMultiAxisVibrate) Encode(srcbuf []byte) (int, error) {

	buf := make([]byte, len(srcbuf))
	copy(buf, srcbuf)

	if len(buf) < len(s.Data)*2+8 {
		return 0, fmt.Errorf("out of allocated memory")
	}
	idx := 0

	// stype(1)
	buf[idx] = s.SType
	idx++
	// axle(1)
	buf[idx] = s.AxleNum
	idx++
	// samplerate(2)
	binary.BigEndian.PutUint16(buf[idx:idx+2], s.SampleRate)
	idx += 2
	// dynamicrange(1)
	binary.BigEndian.PutUint16(buf[idx:idx+2], s.DynamicRange)
	idx += 2
	// resolution(1)
	binary.BigEndian.PutUint16(buf[idx:idx+2], s.Resolution)
	idx += 2

	copy(buf[idx:], s.Bytes[:s.Count*int(s.AxleNum)*2])

	idx += s.Count * int(s.AxleNum) * 2
	return idx, nil
}

// Decode - 解码
func (s *SegmentMultiAxisVibrate) Decode(srcdata []byte) error {

	data := make([]byte, len(srcdata))
	copy(data, srcdata)

	idx := 0
	// stype(1)
	s.SType = data[idx]
	idx++
	// axle(1)
	s.AxleNum = data[idx]
	idx++
	// samplerate(2)
	s.SampleRate = binary.BigEndian.Uint16(data[idx : idx+2])
	idx += 2
	// dynamicrange(1)
	s.DynamicRange = binary.BigEndian.Uint16(data[idx : idx+2])
	idx += 2
	// resolution(1)
	s.Resolution = binary.BigEndian.Uint16(data[idx : idx+2])
	idx += 2

	s.Bytes = data[idx:]
	s.Count = (len(data) - idx) / int(s.AxleNum) / 2
	for i := 0; i < s.Count*int(s.AxleNum); i++ {
		axle := int16(binary.BigEndian.Uint16(data[idx : idx+2]))
		data[idx], data[idx+1] = data[idx+1], data[idx]
		if i >= len(s.Data) {
			s.Data = append(s.Data, axle)
		} else {
			s.Data[i] = axle
		}
		idx += 2
	}
	return s.Validate()
}

// Dump -
func (s *SegmentMultiAxisVibrate) Dump() {
	title := fmt.Sprintf(" Dump Vibrate Samplerate: %d, Range: %d, Resolution: %d, Axle: %d\n  ",
		s.SampleRate, s.DynamicRange, s.Resolution, s.AxleNum,
	)
	utils.Hexdump(title, s.Bytes)
}

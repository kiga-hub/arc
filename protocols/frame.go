package protocols

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"common/utils"
)

// DefaultHeadLength head + version + size
const DefaultHeadLength = 9 // head + version + size
// LengthWithoutData ...
const LengthWithoutData uint32 = 32

// FrameVersion -
const FrameVersion byte = 0x01

var (
	//Head ...
	Head = []byte{0xFC, 0xFC, 0xFC, 0xFC}

	//End ...
	End uint8 = 0xFD

	// MaxSize -
	MaxSize uint32 = 1024 * 5
)

// Frame 全包大小 41+n
type Frame struct {
	// this part of data is used to validate
	Head      [4]byte //4 头标志 0xFC 0xFC 0xFC 0xFC
	Version   byte    //1 包格式版本 = 1
	Size      uint32  //4 包大小 [Timestamp, End] = 32+n BigEndian
	Timestamp int64   //8 时间戳/序号 精确到毫秒 BigEndian
	BasicInfo [6]byte //6 (2-客户 1-设备 1-年份 1-月份 1-日期)
	ID        [6]byte //6 设备编号 (Mac Address)
	// this part of data will be stored
	Firmware  [3]byte   //3 固件版本3位
	Hardware  byte      //1 硬件版本1位
	Protocol  uint16    //2 协议版本 = 1 BigEndian
	Flag      [3]byte   //3 标志位 前8位表示数据形式 AVT_____ 第9位表示 有线/无线 其他预留
	DataGroup DataGroup //n 数据
	// this part of data is used to validate
	Crc uint16 //2 校验位 [Timestamp, Data], CRC-16 BigEndian
	End byte   //1 结束标志 0xFD
}

// ConfigFrame -
func ConfigFrame(maxSize uint32) error {
	MaxSize = maxSize
	return nil
}

// NewDefaultFrame -
func NewDefaultFrame() *Frame {
	return &Frame{
		Head:      [4]byte{0xFC, 0xFC, 0xFC, 0xFC},
		Version:   FrameVersion,
		Size:      32,
		BasicInfo: [6]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		ID:        [6]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		Firmware:  [3]byte{0x00, 0x00, 0x00},
		Hardware:  0x00,
		Protocol:  2,
		Flag:      [3]byte{0x00, 0x00, 0x00},
		End:       0xFD,
	}
}

// SetID - 设置传感器ID， 模拟客户端使用
func (f *Frame) SetID(id uint64) *Frame {
	f.ID[5] = byte(id & 0xFF)
	f.ID[4] = byte((id >> 8) & 0xFF)
	f.ID[3] = byte((id >> 16) & 0xFF)
	f.ID[2] = byte((id >> 24) & 0xFF)
	f.ID[1] = byte((id >> 32) & 0xFF)
	f.ID[0] = byte((id >> 40) & 0xFF)
	return f
}

// GetID - 获取传感器ID
func (f *Frame) GetID() uint64 {
	var sensorid uint64
	for _, b := range f.ID {
		sensorid <<= 8
		sensorid += uint64(b)
	}
	return sensorid
}

// SetProto -
func (f *Frame) SetProto(proto uint16) *Frame {
	f.Protocol = proto
	return f
}

// SetDataGroup -
func (f *Frame) SetDataGroup(dg *DataGroup) *Frame {
	f.DataGroup = *dg
	f.Size = LengthWithoutData + 1
	for _, s := range dg.Sizes {
		f.Size += 4
		f.Size += s
	}
	return f
}

// IProto -
type IProto interface {
	Encode([]byte) (int, error)
}

// Dump -
func (f *Frame) Dump() {
	fmt.Printf("ID: ")
	for i := range f.ID {
		fmt.Printf("%02X", f.ID[i])
	}
	fmt.Printf(", Timestamp: %d, Size: %d\n", f.Timestamp, f.Size)

	model := ""
	if f.Flag[0]&0x80 == 1 {
		model += "A"
	}
	if f.Flag[0]&0x40 == 1 {
		model += "V"
	}
	if f.Flag[0]&0x20 == 1 {
		model += "T"
	}
	if model != "" {
		model += "-"
	}
	if f.Flag[1]&0x40 == 1 {
		model += "W"
	} else {
		model += "C"
	}
	model += fmt.Sprintf("%02d", f.Flag[1]&0x3F)
	fmt.Printf("Fw: V%d.%d.%d, Hw: V%d, Flag: %s, Base: ", f.Firmware[0], f.Firmware[1], f.Firmware[2], f.Hardware, model)
	for i := range f.BasicInfo {
		fmt.Printf("%02X", f.BasicInfo[i])
	}
	fmt.Printf("\n")
	f.DataGroup.Dump()
}

// IsAck - 应答包号标志位
func (f *Frame) IsAck() bool {
	return f.Flag[2]&0x20 != 0
}

// SetAckFlag - 设置应答包号标志位
func (f *Frame) SetAckFlag() {
	f.Flag[2] |= 0x20
}

// IsTimeAlign - 数据包时间对齐标志位
func (f *Frame) IsTimeAlign() bool {
	return f.Flag[2]&0x80 != 0
}

// SetTimeAlignFlag -
func (f *Frame) SetTimeAlignFlag() {
	f.Flag[2] |= 0x80
}

// IsEnd - 数据结束标志位
func (f *Frame) IsEnd() bool {
	return f.Flag[2]&0x40 != 0
}

// SetEndFlag - 数据结束标志位
func (f *Frame) SetEndFlag() {
	f.Flag[2] |= 0x40
}

// Encode 字节数组写入当前流, 直接对buf操作
func (f *Frame) Encode(buf []byte) (int, error) {
	if len(buf) < int(f.Size)+DefaultHeadLength {
		return 0, fmt.Errorf("frame out of allocated memory")
	}
	idx := 0

	// head(4)
	for i := 0; i < 4; i++ {
		buf[idx] = f.Head[i]
		idx++
	}
	// version(1)
	buf[idx] = f.Version
	idx++
	// size(4)
	binary.BigEndian.PutUint32(buf[idx:idx+4], f.Size)
	idx += 4
	// timestamp(8)
	binary.BigEndian.PutUint64(buf[idx:idx+8], uint64(f.Timestamp))
	idx += 8
	// basicinfo(6)
	for i := 0; i < 6; i++ {
		buf[idx] = f.BasicInfo[i]
		idx++
	}
	// ID(6)
	for i := 0; i < 6; i++ {
		buf[idx] = f.ID[i]
		idx++
	}
	// firmware(3)
	for i := 0; i < 3; i++ {
		buf[idx] = f.Firmware[i]
		idx++
	}
	// hardware(1)
	buf[idx] = f.Hardware
	idx++
	// protocols(2)
	binary.BigEndian.PutUint16(buf[idx:idx+2], f.Protocol)
	idx += 2
	// flag(3)
	for i := 0; i < 3; i++ {
		buf[idx] = f.Flag[i]
		idx++
	}
	// datagroup
	n, err := f.DataGroup.Encode(buf[idx:])
	idx += n
	if err != nil {
		return idx, err
	}

	// crc(2)
	f.Crc = utils.CheckSum(buf[DefaultHeadLength:idx])
	binary.BigEndian.PutUint16(buf[idx:idx+2], f.Crc)
	idx += 2
	// end(1)
	buf[idx] = f.End
	idx++
	return idx, nil
}

// Decode 当前流中读取字节
func (f *Frame) Decode(srcdata []byte) error {
	idx := 0

	data := make([]byte, len(srcdata))
	copy(data, srcdata)

	// head(4)
	for i := 0; i < 4; i++ {
		f.Head[i] = data[idx]
		idx++
	}
	// version(1)
	f.Version = data[idx]
	idx++
	// size(4)
	f.Size = binary.BigEndian.Uint32(data[idx : idx+4])
	idx += 4
	// timestamp(8)
	f.Timestamp = int64(binary.BigEndian.Uint64(data[idx : idx+8]))
	idx += 8
	// basicinfo(6)
	for i := 0; i < 6; i++ {
		f.BasicInfo[i] = data[idx]
		idx++
	}
	// ID(6)
	for i := 0; i < 6; i++ {
		f.ID[i] = data[idx]
		idx++
	}
	// Firmware(3)
	for i := 0; i < 3; i++ {
		f.Firmware[i] = data[idx]
		idx++
	}
	// Hardware(1)
	f.Hardware = data[idx]
	idx++
	// Protocol(2)
	f.Protocol = binary.BigEndian.Uint16(data[idx : idx+2])
	idx += 2
	// Flag(3)
	for i := 0; i < 3; i++ {
		f.Flag[i] = data[idx]
		idx++
	}

	// Crc(2)
	f.Crc = binary.BigEndian.Uint16(data[len(data)-3 : len(data)-1])
	// End(1)
	f.End = data[len(data)-1]

	// DataGroup
	return f.DataGroup.Decode(data[idx : len(data)-3])
}

// FrameValidate -  包格式检查
func FrameValidate(srcdata []byte) error {

	data := make([]byte, len(srcdata))
	copy(data, srcdata)

	idx := 0

	// check head(4)
	if !bytes.Equal(data[idx:idx+4], Head[:]) {
		return fmt.Errorf("invalid frame header(%02x%02x%02x%02x)", data[idx], data[idx+1], data[idx+2], data[idx+3])
	}
	idx += 4

	// check version(1)
	if data[idx] != 1 {
		return fmt.Errorf("invalid frame version(%d)", data[idx])
	}
	idx++

	// size(4)
	fSize := binary.BigEndian.Uint32(data[idx : idx+4])
	if fSize > MaxSize || fSize < LengthWithoutData || fSize != uint32(len(data)-DefaultHeadLength) {
		return fmt.Errorf("invalid frame size(%d)", fSize)
	}
	idx += 4

	// timestamp(8)
	seq := int64(binary.BigEndian.Uint64(data[idx : idx+8]))
	// TODO: check timestamp
	idx += 8

	// basicinfo(6)
	// TODO: check basicinfo
	idx += 6

	// ID(6)
	// TODO: check id
	clientid := hex.EncodeToString(data[idx : idx+6])
	idx += 6

	// Firmware(3)
	// TODO: check firmware
	idx += 3

	// Hardware(1)
	// TODO: check hardware
	idx++

	// Protocol(2)
	protocol := binary.BigEndian.Uint16(data[idx : idx+2])
	if protocol != 2 {
		return fmt.Errorf("[%s][%d]invalid protocol(%d)", clientid, seq, protocol)
	}
	idx += 2

	// Flag(3)
	// TODO: check flag
	idx += 3

	// DataGroup.Count, 新增SegmentAudioV2 和SegmentNumericalTable
	SegmentCount := data[idx]
	if SegmentCount > DataGroupMaxSegmentCount {
		return fmt.Errorf("[%s][%d]invalid datagroup count(%d)", clientid, seq, SegmentCount)
	}
	idx++
	// DataGroup.Segments
	SegmentSize := 0
	for i := 0; i < int(SegmentCount); i++ {
		// Sizes
		Size := binary.BigEndian.Uint32(data[idx : idx+4]) // Sizes
		// Segments
		SegmentIdx := idx + ((int(SegmentCount) - i) * 4) + SegmentSize
		if SegmentIdx+int(Size) > len(data) {
			return fmt.Errorf("[%s][%d]datagroup valid size(%d:%d)", clientid, seq, i+1, Size)
		}
		switch data[SegmentIdx] {
		case STypeAudio:
			if err := AudioValidate(data[SegmentIdx : SegmentIdx+int(Size)]); err != nil {
				return fmt.Errorf("[%s][%d]%s", clientid, seq, err.Error())
			}
		case STypeVibrate:
			if err := VibrateValidate(data[SegmentIdx : SegmentIdx+int(Size)]); err != nil {
				return fmt.Errorf("[%s][%d]%s", clientid, seq, err.Error())
			}
		case STypeTemperature:
			if err := TemperatureValidate(data[SegmentIdx : SegmentIdx+int(Size)]); err != nil {
				return fmt.Errorf("[%s][%d]%s", clientid, seq, err.Error())
			}
		case STypeAudioV2:
			if err := AudioV2Validate(data[SegmentIdx : SegmentIdx+int(Size)]); err != nil {
				return fmt.Errorf("[%s][%d]%s", clientid, seq, err.Error())
			}
		case STypeNumericalTable:
			if err := NumericalTableValidate(data[SegmentIdx : SegmentIdx+int(Size)]); err != nil {
				return fmt.Errorf("[%s][%d]%s", clientid, seq, err.Error())
			}
		default:
			return fmt.Errorf("[%s][%d]datagroup invalid segment stype(%d:%d)", clientid, seq, i+1, data[SegmentIdx])
		}
		idx += 4
		SegmentSize += int(Size)
	}
	idx += SegmentSize

	// Crc(2)
	Crc := binary.BigEndian.Uint16(data[idx : idx+2])
	if Crc != utils.CheckSum(data[DefaultHeadLength:idx]) {
		return fmt.Errorf("[%s][%d]invalid frame crc(%d)", clientid, seq, Crc)
	}
	idx += 2

	// End(1)
	if data[idx] != End {
		return fmt.Errorf("[%s][%d]invalid frame end(%02x)", clientid, seq, data[idx])
	}

	return nil
}

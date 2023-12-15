package protocols

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/kiga-hub/arc/utils"
)

// DefaultHeadLength head  + size
const DefaultHeadLength = 8

// LengthWithoutData ... Timestamp + ID + Crc + End
const LengthWithoutData uint32 = 17

var (
	//Head ...
	Head = []byte{0xFC, 0xFC, 0xFC, 0xFC}

	//End ...
	End uint8 = 0xFD

	// MaxSize -
	MaxSize uint32 = 1024 * 6
)

// Frame package size 25+n
type Frame struct {
	Head      [4]byte   //4 Head 0xFC 0xFC 0xFC 0xFC
	Size      uint32    //4 Package size  [Timestamp, End] = 17+n BigEndian
	Timestamp int64     //8 timestamp ms BigEndian
	ID        [6]byte   //6 machin id (Mac Address)
	DataGroup DataGroup //n data
	Crc       uint16    //2 crc [Timestamp, Data], CRC-16 BigEndian
	End       byte      //1 End 0xFD
}

// ConfigFrame -
//
//goland:noinspection GoUnusedExportedFunction
func ConfigFrame(maxSize uint32) error {
	MaxSize = maxSize
	return nil
}

// NewDefaultFrame -
func NewDefaultFrame() *Frame {
	return &Frame{
		Head: [4]byte{0xFC, 0xFC, 0xFC, 0xFC},
		Size: LengthWithoutData,
		ID:   [6]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		End:  0xFD,
	}
}

// SetID - set machin id for simulate client
func (f *Frame) SetID(id uint64) *Frame {
	f.ID[5] = byte(id & 0xFF)
	f.ID[4] = byte((id >> 8) & 0xFF)
	f.ID[3] = byte((id >> 16) & 0xFF)
	f.ID[2] = byte((id >> 24) & 0xFF)
	f.ID[1] = byte((id >> 32) & 0xFF)
	f.ID[0] = byte((id >> 40) & 0xFF)
	return f
}

// GetID - get ID
func (f *Frame) GetID() uint64 {
	var sensorID uint64
	for _, b := range f.ID {
		sensorID <<= 8
		sensorID += uint64(b)
	}
	return sensorID
}

// SetDataGroup -
func (f *Frame) SetDataGroup(dg *DataGroup) *Frame {
	f.DataGroup = *dg
	f.Size = LengthWithoutData + 1 // add count: 1byte
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
	for i := range f.ID {
		fmt.Printf("%02X", f.ID[i])
	}
	fmt.Printf(", Timestamp: %d, Size: %d\n", f.Timestamp, f.Size)

	fmt.Printf("\n")
	f.DataGroup.Dump()
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

	// size(4)
	binary.BigEndian.PutUint32(buf[idx:idx+4], f.Size)
	idx += 4

	// timestamp(8)
	binary.BigEndian.PutUint64(buf[idx:idx+8], uint64(f.Timestamp))
	idx += 8

	// ID(6)
	for i := 0; i < 6; i++ {
		buf[idx] = f.ID[i]
		idx++
	}

	// dataGroup TODO: check size
	n, err := f.DataGroup.Encode(buf[idx:])
	idx += n - 1
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
func (f *Frame) Decode(srcData []byte) error {
	idx := 0
	data := make([]byte, len(srcData))
	copy(data, srcData)

	// head(4)
	for i := 0; i < 4; i++ {
		f.Head[i] = data[idx]
		idx++
	}

	// size(4)
	f.Size = binary.BigEndian.Uint32(data[idx : idx+4])
	idx += 4
	// timestamp(8)
	f.Timestamp = int64(binary.BigEndian.Uint64(data[idx : idx+8]))
	idx += 8

	// ID(6)
	for i := 0; i < 6; i++ {
		f.ID[i] = data[idx]
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
func FrameValidate(srcData []byte) error {
	data := make([]byte, len(srcData))
	copy(data, srcData)

	idx := 0

	// check head(4)
	if !bytes.Equal(data[idx:idx+4], Head[:]) {
		return fmt.Errorf("invalid frame header(%02x%02x%02x%02x)", data[idx], data[idx+1], data[idx+2], data[idx+3])
	}
	idx += 4

	// size(4)
	// fSize := binary.BigEndian.Uint32(data[idx : idx+4])
	// if fSize > MaxSize || fSize < LengthWithoutData || fSize != uint32(len(data)-DefaultHeadLength) {
	// 	return fmt.Errorf("invalid frame size(%d)", fSize)
	// }
	_ = binary.BigEndian.Uint32(data[idx : idx+4])
	idx += 4

	// timestamp(8)
	seq := int64(binary.BigEndian.Uint64(data[idx : idx+8]))
	// TODO: check timestamp
	idx += 8

	// ID(6)
	// TODO: check id
	clientID := hex.EncodeToString(data[idx : idx+6])
	idx += 6

	// DataGroup.Count
	SegmentCount := data[idx]
	if SegmentCount > 255 {
		return fmt.Errorf("[%s][%d]invalid datagroup count(%d)", clientID, seq, SegmentCount)
	}
	idx++
	// DataGroup.Segments
	SegmentSize := 0
	for i := 0; i < int(SegmentCount); i++ {
		// Sizes
		Size := binary.BigEndian.Uint32(data[idx : idx+4]) // Sizes
		// Segments * size
		SegmentIdx := idx + ((int(SegmentCount) - i) * 4) + SegmentSize
		if SegmentIdx+int(Size) > len(data) {
			return fmt.Errorf("[%s][%d]datagroup valid size(%d:%d)", clientID, seq, i+1, Size)
		}
		switch data[SegmentIdx] {
		case STypeArc:
			if err := ArcSegmentValidate(data[SegmentIdx : SegmentIdx+int(Size)]); err != nil {
				return fmt.Errorf("[%s][%d]%s", clientID, seq, err.Error())
			}
		default:
			return fmt.Errorf("[%s][%d]datagroup invalid segment stype(%d:%d)", clientID, seq, i+1, data[SegmentIdx])
		}
		idx += 4
		SegmentSize += int(Size)
	}
	idx += SegmentSize

	// Crc(2)
	Crc := binary.BigEndian.Uint16(data[idx : idx+2])
	if Crc != utils.CheckSum(data[DefaultHeadLength:idx]) {
		return fmt.Errorf("[%s][%d]invalid frame crc(%d)", clientID, seq, Crc)
	}
	idx += 2

	// End(1)
	if data[idx] != End {
		return fmt.Errorf("[%s][%d]invalid frame end(%02x)", clientID, seq, data[idx])
	}
	idx++
	return nil
}

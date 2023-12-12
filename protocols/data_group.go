package protocols

import (
	"encoding/binary"
	"fmt"
)

const (
	// STypeTemperature 温度Type
	STypeTemperature byte = 10

	// ItemSampleRate  -
	ItemSampleRate = 1
)

// ISegment -
type ISegment interface {
	Encode([]byte) (int, error)
	Type() byte
	Size() uint32
	Dump()
}

// DataGroup Protocol=2   CType 1:GetCollector2Structure.Size()
type DataGroup struct {
	Count    byte       //1 数据类型个数
	Sizes    []uint32   //4 每个类型数据大小
	STypes   []byte     // 解析段类型记录
	Segments []ISegment //n 数据
}

// NewDefaultDataGroup -
func NewDefaultDataGroup() *DataGroup {
	return &DataGroup{}
}

// Validate - 校验
func (d *DataGroup) Validate() error {
	if d.Count <= 0 {
		return fmt.Errorf("count out of range %d", d.Count)
	}
	if uint8(len(d.Sizes)) != d.Count {
		return fmt.Errorf("size count don't match %d != %d", len(d.Sizes), d.Count)
	}
	for _, size := range d.Sizes {
		if size > MaxSize {
			return fmt.Errorf("size Over the maximum %d", size)
		}
	}
	return nil
}

// AppendSegment - 添加数据段
func (d *DataGroup) AppendSegment(segment ISegment) {
	d.Count++
	d.Sizes = append(d.Sizes, segment.Size())
	d.Segments = append(d.Segments, segment)
}

// GetSegment - 获取数据段
func (d *DataGroup) GetSegment(SType byte) (ISegment, error) {
	for _, s := range d.Segments {
		if s.Type() == SType {
			return s, nil
		}
	}
	return nil, fmt.Errorf("not find type %d", SType)
}

// GetTSegment - 获取温度数据段
func (d *DataGroup) GetTSegment() (*SegmentTemperature, error) {
	s, err := d.GetSegment(STypeTemperature)
	if err != nil {
		return nil, fmt.Errorf("temperature %s", err)
	}
	return s.(*SegmentTemperature), nil
}

// Decode - 解码
func (d *DataGroup) Decode(data []byte) error {
	idx := 0

	// 获取数据段个数
	d.Count = data[idx]
	idx++
	if d.Count <= 0 {
		return fmt.Errorf("data group count %d", d.Count)
	}

	d.Sizes = make([]uint32, d.Count)
	for i := 0; i < int(d.Count); i++ {
		d.Sizes[i] = binary.BigEndian.Uint32(data[idx : idx+4])
		idx += 4
	}

	// 校验
	if err := d.Validate(); err != nil {
		return err
	}

	// 解析数据段
	for i := 0; i < int(d.Count); i++ {
		if i >= len(d.STypes) {
			d.STypes = append(d.STypes, 0)
		}
		switch data[idx] {
		case STypeTemperature:
			st, err := d.GetTSegment()
			if err != nil {
				st = NewDefaultSegmentTemperature()
				d.Segments = append(d.Segments, st)
			}
			if err := st.Decode(data[idx : idx+int(d.Sizes[i])]); err != nil {
				return err
			}
			d.STypes[i] = STypeTemperature
		default:
			return fmt.Errorf("no match stype")
		}
		idx += int(d.Sizes[i])
	}
	return nil
}

// Encode - 编码
func (d *DataGroup) Encode(buf []byte) (int, error) {
	// 检查缓存大小
	minSize := 1
	for _, size := range d.Sizes {
		minSize += int(size)
		minSize += 4
	}
	if len(buf) < minSize {
		return 0, fmt.Errorf("datagroup out of allocated memory")
	}

	idx := 0

	// count(1)
	buf[idx] = d.Count
	idx++
	// size
	for _, size := range d.Sizes {
		binary.BigEndian.PutUint32(buf[idx:idx+4], size)
		idx += 4
	}
	// segments
	for i, s := range d.Segments {
		n, err := s.Encode(buf[idx:])
		idx += n
		if err != nil {
			return idx, err
		}
		if n != int(d.Sizes[i]) {
			return idx, fmt.Errorf("data segment sizes do not match")
		}
	}
	return idx, nil
}

// Dump -
func (d *DataGroup) Dump() {
	fmt.Printf("DataGroup Count: %d\n", d.Count)
	for _, tp := range d.STypes {
		s, err := d.GetSegment(tp)
		if err != nil {
			fmt.Printf("|Not Found Segment Type: %d\n", tp)
		}
		s.Dump()
	}
}

package protocols

import (
	"encoding/binary"
	"fmt"
)

const (
	// STypeArc arc type
	STypeArc byte = 10

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
	Count    byte       //1 segment count
	Sizes    []uint32   //4 each segment size
	STypes   []byte     // segment type
	Segments []ISegment //n data
}

// NewDefaultDataGroup -
func NewDefaultDataGroup() *DataGroup {
	return &DataGroup{}
}

// Validate - Validate
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

// AppendSegment - AppendSegment
func (d *DataGroup) AppendSegment(segment ISegment) {
	d.Count++
	d.Sizes = append(d.Sizes, segment.Size())
	d.STypes = append(d.STypes, segment.Type())
	d.Segments = append(d.Segments, segment)
}

// GetSegment - GetSegment
func (d *DataGroup) GetSegment(SType byte) (ISegment, error) {
	for _, s := range d.Segments {
		if s.Type() == SType {
			return s, nil
		}
	}
	return nil, fmt.Errorf("not find type %d", SType)
}

// GetArcSegment - GetArcSegment
func (d *DataGroup) GetArcSegment() (*SegmentArc, error) {
	s, err := d.GetSegment(STypeArc)
	if err != nil {
		return nil, fmt.Errorf("arc %s", err)
	}
	return s.(*SegmentArc), nil
}

// Decode - Decode
func (d *DataGroup) Decode(data []byte) error {
	idx := 0

	// get segment count
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

	// validate
	if err := d.Validate(); err != nil {
		return err
	}

	// decode
	for i := 0; i < int(d.Count); i++ {
		if i >= len(d.STypes) {
			d.STypes = append(d.STypes, 0)
		}

		switch data[idx] {
		case STypeArc:
			st, err := d.GetArcSegment()
			if err != nil {
				st = NewDefaultSegmentArc()
				d.Segments = append(d.Segments, st)
			}
			if err := st.Decode(data[idx : idx+int(d.Sizes[i])]); err != nil {
				return err
			}
			d.STypes[i] = STypeArc
		default:
			return fmt.Errorf("no match stype")
		}
		idx += int(d.Sizes[i])
	}
	return nil
}

// Encode - encode
func (d *DataGroup) Encode(buf []byte) (int, error) {
	// check buf size
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

	// //type 1
	for _, tp := range d.STypes {
		buf[idx] = tp
		idx++
	}

	// segments
	for i, s := range d.Segments {
		n, err := s.Encode(buf[idx:])
		idx += n
		if err != nil {
			return idx, err
		}
		if n != int(d.Sizes[i]) {
			return idx, fmt.Errorf("data segment sizes do not match:%d != %d", n, d.Sizes[i])
		}
	}
	return idx, nil
}

// Dump -
func (d *DataGroup) Dump() {
	for _, tp := range d.STypes {
		s, err := d.GetSegment(tp)
		if err != nil {
			fmt.Printf("|Not Found Segment Type: %d\n", tp)
		}
		s.Dump()
	}
}

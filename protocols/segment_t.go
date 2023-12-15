package protocols

import (
	"fmt"

	"github.com/kiga-hub/arc/utils"
)

// SegmentArc arc
type SegmentArc struct {
	SType byte //1 segment type
	Data  []byte
}

// Validate -
func (s *SegmentArc) Validate() error {
	if s.SType != STypeArc {
		return fmt.Errorf("arc segment type %d", s.SType)
	}
	return nil
}

// NewDefaultSegmentArc -
func NewDefaultSegmentArc() *SegmentArc {
	return &SegmentArc{
		SType: STypeArc,
	}
}

// ArcSegmentValidate - 校验
func ArcSegmentValidate(srcData []byte) error {
	data := make([]byte, len(srcData))
	copy(data, srcData)

	idx := 0

	if data[idx] != STypeArc {
		return fmt.Errorf("arc segment stype invalid(%d)", data[idx])
	}

	return nil
}

// Decode - 解码
func (s *SegmentArc) Decode(srcData []byte) error {

	data := make([]byte, len(srcData))
	copy(data, srcData)

	idx := 0
	// sType(1)
	s.SType = data[idx]
	idx++

	s.Data = data[idx:]
	return s.Validate()
}

// SetData - 获取数据
func (s *SegmentArc) SetData(data []byte) {
	s.Data = make([]byte, len(data))
	copy(s.Data, data)
}

// GetData - 设置数据
func (s *SegmentArc) GetData() []byte {
	return s.Data
}

// GetData - 设置数据
func (s *SegmentArc) GetType() byte {
	return s.SType
}

// Encode - 编码
func (s *SegmentArc) Encode(buf []byte) (int, error) {
	idx := 0

	// sType(1)
	buf[idx] = s.SType
	idx++

	copy(buf[idx:], s.Data)
	idx += len(s.Data)

	return idx, nil
}

// Type - 段类型
func (s *SegmentArc) Type() byte {
	return s.SType
}

// Size - 编码大小
func (s *SegmentArc) Size() uint32 {
	return 1 + uint32(len(s.Data))
}

// Dump -
func (s *SegmentArc) Dump() {
	title := fmt.Sprintf("Dump len: %d\n  ", len(s.GetData()))
	utils.Hexdump(title, s.Data)
}

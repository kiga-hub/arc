package protocols

import (
	"encoding/binary"
	"fmt"

	"github.com/kiga-hub/common/utils"
)

const (
	// ResultScoreTableType 金州水务数值表类型
	ResultScoreTableType = 100
	// segmentNumericalTableoffsetBytes 获取Data，需要偏移字节数
	segmentNumericalTableOffsetBytes = 3
)

// INumerical -
type INumerical interface {
	SetData([]byte)
	GetData() []byte
}

// SegmentNumericalTable 数值表数据块
type SegmentNumericalTable struct {
	SType byte       //1
	TType int16      //2
	Data  INumerical //n 数据
}

// Validate -
func (s *SegmentNumericalTable) Validate() error {
	if s.SType != STypeNumericalTable {
		return fmt.Errorf("numerical type %d", s.SType)
	}
	return nil
}

// NewDefaultSegmentNumericalTable -
func NewDefaultSegmentNumericalTable() *SegmentNumericalTable {
	return &SegmentNumericalTable{
		SType: STypeNumericalTable,
	}
}

// NumericalTableValidate -
func NumericalTableValidate(srcData []byte) error {
	data := make([]byte, len(srcData))
	copy(data, srcData)
	idx := 0

	// stype(1)
	if data[idx] != STypeNumericalTable {
		return fmt.Errorf("numerical segment stype invalid(%d)", data[idx])
	}

	return nil
}

// Encode -
func (s *SegmentNumericalTable) Encode(buf []byte) (int, error) {
	if len(buf) < 1 {
		return 0, fmt.Errorf("out of allocated memory")
	}
	idx := 0
	// SType
	buf[idx] = s.SType
	idx++

	// TType
	binary.BigEndian.PutUint16(buf[idx:idx+2], uint16(s.TType))
	idx += 2
	data := s.GetData()
	copy(buf[idx:], data)
	idx += len(data)
	return idx, nil
}

// Decode -
func (s *SegmentNumericalTable) Decode(srcData []byte) error {
	data := make([]byte, len(srcData))
	copy(data, srcData)

	idx := 0
	// stype(1)
	s.SType = data[idx]
	idx++

	// TType(2)
	s.TType = int16(binary.BigEndian.Uint16(data[idx : idx+2]))
	idx += 2

	s.SetData(data[idx:])
	return s.Validate()
}

// SetData - 设置数据
func (s *SegmentNumericalTable) SetData(data []byte) {
	s.Data.SetData(data)
}

// GetData - 获取数据
func (s *SegmentNumericalTable) GetData() []byte {
	return s.Data.GetData()
}

// Type -
func (s *SegmentNumericalTable) Type() byte {
	return s.SType
}

// TableType -
func (s *SegmentNumericalTable) TableType() int16 {
	return s.TType
}

// Size -
func (s *SegmentNumericalTable) Size() uint32 {
	return uint32(len(s.Data.GetData()) + segmentNumericalTableOffsetBytes)
}

// Dump -
func (s *SegmentNumericalTable) Dump() {
	title := fmt.Sprintf("Dump Numerical TTYpe: %d", s.TType)
	utils.Hexdump(title, s.GetData())
}

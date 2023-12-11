package protocols

import "encoding/binary"

const (
	// ResultAndScoreDataSize -
	ResultAndScoreDataSize = 3
)

// ResultAndScore 数值表
type ResultAndScore struct {
	Result uint8  //1 结果
	Score  uint16 //2 评分
}

// SetData -
func (jz *ResultAndScore) SetData(buf []byte) {
	jz.Result = buf[0]
	jz.Score = uint16(binary.BigEndian.Uint16(buf[1:3]))
}

// GetData -
func (jz *ResultAndScore) GetData() []byte {
	data := make([]byte, ResultAndScoreDataSize)
	data[0] = jz.Result
	binary.BigEndian.PutUint16(data[1:3], jz.Score)

	return data
}

// GetDataValue -
func (jz *ResultAndScore) GetDataValue() *ResultAndScore {
	data := make([]byte, ResultAndScoreDataSize)
	data[0] = jz.Result
	binary.BigEndian.PutUint16(data[1:3], jz.Score)

	return &ResultAndScore{
		Result: jz.Result,
		Score:  jz.Score,
	}
}

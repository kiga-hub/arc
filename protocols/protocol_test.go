package protocols

import (
	"fmt"
	"testing"
	"time"
)

func TestProto2(t *testing.T) {
	// 1. 准备数据段
	audio := make([]byte, 2048)
	for i := 0; i < 2048; i++ {
		audio[i] = byte(i)
	}

	st := NewDefaultSegmentTemperature()
	st.SetData(25.4)
	if err := st.Validate(); err != nil {
		fmt.Println(err)
	}

	// 2. 添加数据段到组
	g := NewDefaultDataGroup()
	g.AppendSegment(st)
	if err := g.Validate(); err != nil {
		fmt.Println(err)
	}

	// 3. 添加组到包
	p := NewDefaultFrame()
	p.SetProto(2)
	p.SetID(15)
	p.Timestamp = time.Now().UnixNano() / 1e3
	p.Hardware = 0x01
	p.Firmware[0] = 0x0F
	p.SetDataGroup(g)

	// 4. 打包二进制
	buf := make([]byte, p.Size+9)
	n, err := p.Encode(buf)
	if err != nil {
		fmt.Println("%w", err)
	}
	fmt.Println(n)

	// 数据包有效判断
	if err := FrameValidate(buf); err != nil {
		fmt.Println("%w", err)
	}

	// 解析Frame
	p2 := NewDefaultFrame()
	if err := p2.Decode(buf); err != nil {
		fmt.Println("%w", err)
	}
	p2.Dump()
}

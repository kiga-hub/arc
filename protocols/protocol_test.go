package protocols

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestProto2(t *testing.T) {
	// 1. 准备数据段
	audio := make([]byte, 2048)
	for i := 0; i < 2048; i++ {
		audio[i] = byte(i)
	}
	sa := NewDefaultSegmentAudio()
	sa.Data = append(sa.Data, audio...)
	if err := sa.Validate(); err != nil {
		fmt.Println(err)
	}

	st := NewDefaultSegmentTemperature()
	st.SetData(25.4)
	if err := st.Validate(); err != nil {
		fmt.Println(err)
	}

	sv := NewDefaultSegmentVibrate()
	sv.SetOffset(0, 0, 0)
	sv.SampleRate = 0

	for i := 0; i < 102; i++ {
		sv.AppendData(1, 10, 200)
	}

	if err := sv.Validate(); err != nil {
		fmt.Println(err)
	}

	sa2 := NewDefaultSegmentAudioV2()
	sa2.Data = append(sa2.Data, audio...)
	if err := sa2.Validate(); err != nil {
		fmt.Println(err)
	}
	sa2.Gain = 110

	sm := NewDefaultSegmentNumericalTable()
	sm.Data = &ResultAndScore{}

	result := uint8(rand.Intn(255))
	score := uint16(rand.Intn(65535))

	smBuf := make([]byte, ResultAndScoreDataSize)
	smBuf[0] = result
	binary.BigEndian.PutUint16(smBuf[1:3], score)
	sm.SetData(smBuf)

	// 2. 添加数据段到组
	g := NewDefaultDataGroup()
	g.AppendSegment(sa)
	g.AppendSegment(st)
	g.AppendSegment(sv)
	g.AppendSegment(sa2)
	g.AppendSegment(sm)
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

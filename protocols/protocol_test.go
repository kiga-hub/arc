package protocols

import (
	"fmt"
	"testing"
	"time"
)

func TestProto2(t *testing.T) {

	dataCount := 10
	// 1. prepare data
	audio := make([]byte, dataCount)
	for i := 0; i < dataCount; i++ {
		audio[i] = byte(i)
	}

	st := NewDefaultSegmentArc()
	st.SetData(audio)
	if err := st.Validate(); err != nil {
		panic(err)
	}

	// 2. add segment to group
	g := NewDefaultDataGroup()
	g.AppendSegment(st)
	if err := g.Validate(); err != nil {
		panic(err)
	}

	_, _ = g.GetArcSegment()

	// 3. add group to frame
	p := NewDefaultFrame()
	p.SetID(15)
	p.Timestamp = time.Now().UnixNano() / 1e3
	p.SetDataGroup(g)

	// 4. encode frame
	buf := make([]byte, p.Size+DefaultHeadLength)
	_, err := p.Encode(buf)
	if err != nil {
		panic(err)
	}

	// 5. validate frame
	if err := FrameValidate(buf); err != nil {
		panic(err)
	}

	// 6. decode frame
	p2 := NewDefaultFrame()
	if err := p2.Decode(buf); err != nil {
		fmt.Println("%w", err)
	}
	p2.Dump()
}

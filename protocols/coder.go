package protocols

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"common/utils"

	"github.com/panjf2000/gnet"
)

// Coder -
type Coder struct {
	IsCrcCheck bool
}

// Decode decodes frames from TCP stream via specific implementation.
func (coder *Coder) Decode(c gnet.Conn) ([]byte, error) {
	// find package head
	idx := 0
	var size int
	var header []byte
	var find bool
	for {
		if idx == 0 || idx+4 > DefaultHeadLength || find {
			idx = 0
			size, header = c.ReadN(DefaultHeadLength)
			if size != DefaultHeadLength {
				return nil, fmt.Errorf("not enough header data")
			}
			if find {
				break
			}
		}
		if !bytes.Equal(header[idx:idx+4], Head[:]) {
			c.ShiftN(1)
			idx++
			continue
		}
		find = true
		if idx == 0 {
			break
		}
	}

	// check version
	if header[4] != 1 {
		c.ShiftN(1)
		return nil, fmt.Errorf("version=!1 ignore this package")
	}

	// check size
	fSize := binary.BigEndian.Uint32(header[5:9])
	if fSize > MaxSize || fSize < LengthWithoutData {
		c.ShiftN(DefaultHeadLength - 4)
		return nil, fmt.Errorf("bad size [%d]", fSize)
	}

	// parse payload
	protocolLen := DefaultHeadLength + int(fSize)
	dataSize, data := c.ReadN(protocolLen)
	if dataSize != protocolLen {
		return nil, fmt.Errorf("not enough payload data")
	}

	// check packet end
	if data[protocolLen-1] != End {
		c.ShiftN(DefaultHeadLength - 1)
		return nil, fmt.Errorf("bad end [%02X]", data[protocolLen-1])
	}

	// check packet crc
	if coder.IsCrcCheck {
		fCrc := binary.BigEndian.Uint16(data[protocolLen-3 : protocolLen-1])
		crc := utils.CheckSum(data[DefaultHeadLength : dataSize-3])
		if crc != fCrc {
			c.ShiftN(protocolLen)
			fmt.Println(FrameValidate(data))
			return nil, fmt.Errorf("bad crc check sum %v != %v", crc, fCrc)
		}
	}

	output := make([]byte, len(data))
	copy(output, data)

	c.ShiftN(protocolLen)

	return output, nil
}

// Encode encodes frames upon server responses into TCP stream.
func (coder *Coder) Encode(c gnet.Conn, buf []byte) ([]byte, error) {
	return buf, nil
}

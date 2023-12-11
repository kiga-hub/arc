package utils

import (
	"testing"
)

func TestHexdump(t *testing.T) {
	data := make([]byte, 256)
	for i := 0; i < 256; i++ {
		data[i] = byte(i)
	}
	Hexdump("", data)
}

package utils

import (
	"fmt"
	"unsafe"
)

// Str2byte -
//
//goland:noinspection GoUnusedExportedFunction
func Str2byte(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

// Hexdump -
func Hexdump(title string, data []byte) {
	size := len(data)
	line := ""
	i := 0
	if title == "" {
		title = "  Dump data"
	}
	fmt.Printf("%s at [%p], len=%d\n", title, data, size)
	for ofs := 0; ofs < size; {
		line = fmt.Sprintf("%08X:", ofs)
		for i = 0; ((ofs + i) < size) && (i < 16); i++ {
			line += fmt.Sprintf(" %02X", data[ofs+i]&0xff)
		}
		for ; i <= 16; i++ {
			line += " | "
		}
		for i = 0; (ofs < size) && (i < 16); i++ {
			c := data[ofs]
			if (c < byte(' ')) || (c > byte('~')) {
				c = byte('.')
			}
			line += fmt.Sprintf("%c", c)
			ofs++
		}
		fmt.Println(line)
	}
}

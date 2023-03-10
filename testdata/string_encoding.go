// Code generated by "gobinenc string.go"; DO NOT EDIT.

package main

import (
	"io"
	"unsafe"
)

func (s *String) WriteTo(w io.Writer) (n int, err error) {
	size := 2
	size += len(s.S)
	buf := make([]byte, size)
	offset := 0
	buf[offset] = byte(len(s.S))
	buf[offset+1] = byte(len(s.S) >> 8)
	offset += 2
	copy(buf[offset:], s.S)
	offset += len(s.S)
	return w.Write(buf)
}

func (s *String) ReadFrom(r io.Reader) error {
	buf := make([]byte, 8)
	var size uint16
	var tmp []byte
	m := 0
	c := 64
	strBuf := make([]byte, c)
	r.Read(buf[:2])
	size = uint16(buf[0]) | (uint16(buf[1]) << 8)
	if c-m < int(size) {
		c = int(size)
		if c < 2*cap(strBuf) {
			c = 2 * cap(strBuf)
		}
		strBuf = append([]byte(nil), make([]byte, c)...)
		m = 0
	}
	r.Read(strBuf[m : m+int(size)])
	tmp = strBuf[m : m+int(size)]
	s.S = *(*string)(unsafe.Pointer(&tmp))
	m += int(size)
	return nil
}

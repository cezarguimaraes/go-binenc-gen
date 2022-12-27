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
	r.Read(buf[:2])
	size = uint16(buf[0]) | (uint16(buf[1]) << 8)
	strBuf_0 := make([]byte, size)
	r.Read(strBuf_0)
	s.S = *(*string)(unsafe.Pointer(&strBuf_0))
	return nil
}

// Code generated by "gobinenc slice.go"; DO NOT EDIT.

package main

import (
	"io"
)

func (s *Slice) WriteTo(w io.Writer) (n int, err error) {
	size := 2
	size += 1 * len(s.Int8Slice)
	buf := make([]byte, size)
	offset := 0
	buf[offset] = byte(len(s.Int8Slice))
	buf[offset+1] = byte(len(s.Int8Slice) >> 8)
	offset += 2
	for _, v := range s.Int8Slice {
		buf[offset] = byte(uint8(v))
		offset += 1
	}
	return w.Write(buf)
}

func (s *Slice) ReadFrom(r io.Reader) error {
	buf := make([]byte, 8)
	var size uint16
	r.Read(buf[:2])
	size = uint16(buf[0]) | (uint16(buf[1]) << 8)
	s.Int8Slice = make([]int8, size)
	si := int(size)
	for i := 0; i < si; i++ {
		r.Read(buf[:1])
		s.Int8Slice[i] = int8(uint8(buf[0]))
	}
	return nil
}

// Code generated by "gobinenc soa.go"; DO NOT EDIT.

package main

import (
	"io"
	"unsafe"
)

func (s *Inner) WriteTo(w io.Writer) (n int, err error) {
	size := 6
	size += len(s.Str) + 1*len(s.Arr3)
	for _, v := range s.Arr4 {
		size += 2
		size += len(v)
	}
	buf := make([]byte, size)
	offset := 0
	buf[offset] = byte(len(s.Str))
	buf[offset+1] = byte(len(s.Str) >> 8)
	offset += 2
	copy(buf[offset:], s.Str)
	offset += len(s.Str)
	buf[offset] = byte(len(s.Arr3))
	buf[offset+1] = byte(len(s.Arr3) >> 8)
	offset += 2
	for _, v := range s.Arr3 {
		buf[offset] = byte(v)
		offset += 1
	}
	buf[offset] = byte(len(s.Arr4))
	buf[offset+1] = byte(len(s.Arr4) >> 8)
	offset += 2
	for _, v := range s.Arr4 {
		buf[offset] = byte(len(v))
		buf[offset+1] = byte(len(v) >> 8)
		offset += 2
		copy(buf[offset:], v)
		offset += len(v)
	}
	return w.Write(buf)
}

func (s *Inner) ReadFrom(r io.Reader) error {
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
	s.Str = *(*string)(unsafe.Pointer(&tmp))
	m += int(size)
	r.Read(buf[:2])
	size = uint16(buf[0]) | (uint16(buf[1]) << 8)
	s.Arr3 = make([]uint8, size)
	si := int(size)
	for i := 0; i < si; i++ {
		r.Read(buf[:1])
		s.Arr3[i] = uint8(buf[0])
	}
	r.Read(buf[:2])
	size = uint16(buf[0]) | (uint16(buf[1]) << 8)
	s.Arr4 = make([]string, size)
	si1 := int(size)
	for i1 := 0; i1 < si1; i1++ {
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
		s.Arr4[i1] = *(*string)(unsafe.Pointer(&tmp))
		m += int(size)
	}
	return nil
}

func (s *Outer) WriteTo(w io.Writer) (n int, err error) {
	size := 6
	size += 1*len(s.Arr1) + 1*len(s.Arr2)
	for _, v := range s.Inners {
		size += 6
		size += len(v.Str) + 1*len(v.Arr3)
		for _, v1 := range v.Arr4 {
			size += 2
			size += len(v1)
		}
	}
	buf := make([]byte, size)
	offset := 0
	buf[offset] = byte(len(s.Arr1))
	buf[offset+1] = byte(len(s.Arr1) >> 8)
	offset += 2
	for _, v := range s.Arr1 {
		buf[offset] = byte(v)
		offset += 1
	}
	buf[offset] = byte(len(s.Inners))
	buf[offset+1] = byte(len(s.Inners) >> 8)
	offset += 2
	for _, v := range s.Inners {
		buf[offset] = byte(len(v.Str))
		buf[offset+1] = byte(len(v.Str) >> 8)
		offset += 2
		copy(buf[offset:], v.Str)
		offset += len(v.Str)
		buf[offset] = byte(len(v.Arr3))
		buf[offset+1] = byte(len(v.Arr3) >> 8)
		offset += 2
		for _, v1 := range v.Arr3 {
			buf[offset] = byte(v1)
			offset += 1
		}
		buf[offset] = byte(len(v.Arr4))
		buf[offset+1] = byte(len(v.Arr4) >> 8)
		offset += 2
		for _, v1 := range v.Arr4 {
			buf[offset] = byte(len(v1))
			buf[offset+1] = byte(len(v1) >> 8)
			offset += 2
			copy(buf[offset:], v1)
			offset += len(v1)
		}
	}
	buf[offset] = byte(len(s.Arr2))
	buf[offset+1] = byte(len(s.Arr2) >> 8)
	offset += 2
	for _, v := range s.Arr2 {
		buf[offset] = byte(uint8(v))
		offset += 1
	}
	return w.Write(buf)
}

func (s *Outer) ReadFrom(r io.Reader) error {
	buf := make([]byte, 8)
	var size uint16
	var tmp []byte
	m := 0
	c := 64
	strBuf := make([]byte, c)
	r.Read(buf[:2])
	size = uint16(buf[0]) | (uint16(buf[1]) << 8)
	s.Arr1 = make([]uint8, size)
	si := int(size)
	for i := 0; i < si; i++ {
		r.Read(buf[:1])
		s.Arr1[i] = uint8(buf[0])
	}
	r.Read(buf[:2])
	size = uint16(buf[0]) | (uint16(buf[1]) << 8)
	s.Inners = make([]Inner, size)
	si1 := int(size)
	for i1 := 0; i1 < si1; i1++ {
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
		s.Inners[i1].Str = *(*string)(unsafe.Pointer(&tmp))
		m += int(size)
		r.Read(buf[:2])
		size = uint16(buf[0]) | (uint16(buf[1]) << 8)
		s.Inners[i1].Arr3 = make([]uint8, size)
		si2 := int(size)
		for i2 := 0; i2 < si2; i2++ {
			r.Read(buf[:1])
			s.Inners[i1].Arr3[i2] = uint8(buf[0])
		}
		r.Read(buf[:2])
		size = uint16(buf[0]) | (uint16(buf[1]) << 8)
		s.Inners[i1].Arr4 = make([]string, size)
		si3 := int(size)
		for i3 := 0; i3 < si3; i3++ {
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
			s.Inners[i1].Arr4[i3] = *(*string)(unsafe.Pointer(&tmp))
			m += int(size)
		}
	}
	r.Read(buf[:2])
	size = uint16(buf[0]) | (uint16(buf[1]) << 8)
	s.Arr2 = make([]int8, size)
	si4 := int(size)
	for i4 := 0; i4 < si4; i4++ {
		r.Read(buf[:1])
		s.Arr2[i4] = int8(uint8(buf[0]))
	}
	return nil
}

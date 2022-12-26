package main_test

import (
	"bytes"
	"encoding/gob"
	"io"
	"math/rand"
	"testing"
	"unsafe"
)

type Person struct {
	Name    string
	Kids    []string
	Touches int
	MinT    int64
	MaxT    int64
	MeanT   int64
}

var s = &Person{
	Name:    "john doe",
	Kids:    []string{"jane", "bob", "santa"},
	Touches: 42,
	MinT:    1,
	MaxT:    11,
	MeanT:   6,
}

var big = &Person{
	Name:    "idk",
	Kids:    []string{},
	Touches: 42,
	MinT:    1,
	MaxT:    11,
	MeanT:   6,
}

func init() {
	nKids := 1024
	for i := 0; i < nKids; i++ {
		big.Kids = append(big.Kids, randomString(50))
	}
}

func randomString(l int) string {
	bytes := make([]byte, l)
	for i := 0; i < l; i++ {
		bytes[i] = byte(randInt(65, 90))
	}
	return string(bytes)
}

func randInt(min int, max int) int {
	// using unseeded math/rand for deterministic generation
	return min + rand.Intn(max-min)
}

func BenchmarkReadSmallPtr(b *testing.B) {
	b.SetBytes(58)
	ser := s.Bytes()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := io.Reader(bytes.NewBuffer(ser))
		out := &Person{}
		size := 0
		intBuf := make([]byte, 8)
		// Name
		r.Read(intBuf[:2])
		size = int(intBuf[0]) + (int(intBuf[1]) << 8)
		nameBuf := make([]byte, size)
		r.Read(nameBuf[:size])
		out.Name = *(*string)(unsafe.Pointer(&nameBuf))
		// Kids
		r.Read(intBuf[:2])
		size = int(intBuf[0]) + (int(intBuf[1]) << 8)
		out.Kids = make([]string, size)
		iEnd := size
		for i := 0; i < iEnd; i++ {
			r.Read(intBuf[:2])
			size = int(intBuf[0]) + (int(intBuf[1]) << 8)
			iBuf := make([]byte, size)
			r.Read(iBuf[:size])
			out.Kids[i] = *(*string)(unsafe.Pointer(&iBuf))
		}
		// Touches
		r.Read(intBuf[:4])
		out.Touches = int(intBuf[0]) + (int(intBuf[1]) << 8) + (int(intBuf[2]) << 16) + (int(intBuf[3]) << 24)

		// MinT
		r.Read(intBuf[:8])
		out.MinT = int64(intBuf[0]) + (int64(intBuf[1]) << 8) + (int64(intBuf[2]) << 16) + (int64(intBuf[3]) << 24) + (int64(intBuf[4]) << 32) + (int64(intBuf[5]) << 40) + (int64(intBuf[6]) << 48) + (int64(intBuf[7]) << 56)

		// MaxT
		r.Read(intBuf[:8])
		out.MaxT = int64(intBuf[0]) + (int64(intBuf[1]) << 8) + (int64(intBuf[2]) << 16) + (int64(intBuf[3]) << 24) + (int64(intBuf[4]) << 32) + (int64(intBuf[5]) << 40) + (int64(intBuf[6]) << 48) + (int64(intBuf[7]) << 56)

		// MeanT
		r.Read(intBuf[:8])
		out.MeanT = int64(intBuf[0]) + (int64(intBuf[1]) << 8) + (int64(intBuf[2]) << 16) + (int64(intBuf[3]) << 24) + (int64(intBuf[4]) << 32) + (int64(intBuf[5]) << 40) + (int64(intBuf[6]) << 48) + (int64(intBuf[7]) << 56)

		/*
			// sanity check
			if diff := cmp.Diff(s, out); diff != "" {
				b.Error(diff)
			}
		*/
	}

}

const initialSize = 64

func BenchmarkSmallManualAppend(b *testing.B) {
	b.SetBytes(58)
	for i := 0; i < b.N; i++ {
		out := &bytes.Buffer{}
		buf := make([]byte, 0, initialSize)
		// Name
		buf = append(buf,
			byte(len(s.Name)),
			byte(len(s.Name)>>8))
		buf = append(buf, []byte(s.Name)...)
		// Kids
		buf = append(buf,
			byte(len(s.Kids)),
			byte(len(s.Kids)>>8))
		for _, v := range s.Kids {
			buf = append(buf,
				byte(len(v)),
				byte(len(v)>>8))
			buf = append(buf, []byte(v)...)
		}
		// Touches
		buf = append(buf,
			byte(uint32(s.Touches)),
			byte(uint32(s.Touches)>>8),
			byte(uint32(s.Touches)>>16),
			byte(uint32(s.Touches)>>24))

		// MinT
		buf = append(buf,
			byte(uint64(s.MinT)),
			byte(uint64(s.MinT)>>8),
			byte(uint64(s.MinT)>>16),
			byte(uint64(s.MinT)>>24),
			byte(uint64(s.MinT)>>32),
			byte(uint64(s.MinT)>>40),
			byte(uint64(s.MinT)>>48),
			byte(uint64(s.MinT)>>56))

		// MaxT
		buf = append(buf,
			byte(uint64(s.MaxT)),
			byte(uint64(s.MaxT)>>8),
			byte(uint64(s.MaxT)>>16),
			byte(uint64(s.MaxT)>>24),
			byte(uint64(s.MaxT)>>32),
			byte(uint64(s.MaxT)>>40),
			byte(uint64(s.MaxT)>>48),
			byte(uint64(s.MaxT)>>56))
		// MeanT
		buf = append(buf,
			byte(uint64(s.MeanT)),
			byte(uint64(s.MeanT)>>8),
			byte(uint64(s.MeanT)>>16),
			byte(uint64(s.MeanT)>>24),
			byte(uint64(s.MeanT)>>32),
			byte(uint64(s.MeanT)>>40),
			byte(uint64(s.MeanT)>>48),
			byte(uint64(s.MeanT)>>56))

		out.Write(buf)
	}
}

func (s *Person) Bytes() []byte {
	size := 32 + len(s.Name)
	for _, v := range s.Kids {
		size += len(v) + 2
	}
	buf := make([]byte, size)
	offset := 0
	// Name
	buf[offset] = byte(len(s.Name))
	buf[offset+1] = byte(len(s.Name) >> 8)
	offset += 2
	copy(buf[offset:], s.Name)
	offset += len(s.Name)
	// Kids
	buf[offset] = byte(len(s.Kids))
	buf[offset+1] = byte(len(s.Kids) >> 8)
	offset += 2
	for _, v := range s.Kids {
		buf[offset] = byte(len(v))
		buf[offset+1] = byte(len(v) >> 8)
		offset += 2
		copy(buf[offset:], v)
		offset += len(v)
	}
	// Touches
	buf[offset] = byte(uint32(s.Touches))
	buf[offset+1] = byte(uint32(s.Touches) >> 8)
	buf[offset+2] = byte(uint32(s.Touches) >> 16)
	buf[offset+3] = byte(uint32(s.Touches) >> 24)
	offset += 4
	// MinT
	buf[offset] = byte(uint64(s.MinT))
	buf[offset+1] = byte(uint64(s.MinT) >> 8)
	buf[offset+2] = byte(uint64(s.MinT) >> 16)
	buf[offset+3] = byte(uint64(s.MinT) >> 24)
	buf[offset+4] = byte(uint64(s.MinT) >> 32)
	buf[offset+5] = byte(uint64(s.MinT) >> 40)
	buf[offset+6] = byte(uint64(s.MinT) >> 48)
	buf[offset+7] = byte(uint64(s.MinT) >> 56)
	offset += 8
	// MaxT
	buf[offset] = byte(uint64(s.MaxT))
	buf[offset+1] = byte(uint64(s.MaxT) >> 8)
	buf[offset+2] = byte(uint64(s.MaxT) >> 16)
	buf[offset+3] = byte(uint64(s.MaxT) >> 24)
	buf[offset+4] = byte(uint64(s.MaxT) >> 32)
	buf[offset+5] = byte(uint64(s.MaxT) >> 40)
	buf[offset+6] = byte(uint64(s.MaxT) >> 48)
	buf[offset+7] = byte(uint64(s.MaxT) >> 56)
	offset += 8
	// MeanT
	buf[offset] = byte(uint64(s.MeanT))
	buf[offset+1] = byte(uint64(s.MeanT) >> 8)
	buf[offset+2] = byte(uint64(s.MeanT) >> 16)
	buf[offset+3] = byte(uint64(s.MeanT) >> 24)
	buf[offset+4] = byte(uint64(s.MeanT) >> 32)
	buf[offset+5] = byte(uint64(s.MeanT) >> 40)
	buf[offset+6] = byte(uint64(s.MeanT) >> 48)
	buf[offset+7] = byte(uint64(s.MeanT) >> 56)
	offset += 8
	return buf
}

func BenchmarkSmallManualPreAlloc(b *testing.B) {
	b.SetBytes(58)
	for i := 0; i < b.N; i++ {
		s.Bytes()
	}
}

func BenchmarkBigManualAppend(b *testing.B) {
	b.SetBytes(53283)
	for i := 0; i < b.N; i++ {
		buf := make([]byte, 0, initialSize)
		// Name
		buf = append(buf,
			byte(len(big.Name)),
			byte(len(big.Name)>>8))
		buf = append(buf, []byte(big.Name)...)
		// Kids
		buf = append(buf,
			byte(len(big.Kids)),
			byte(len(big.Kids)>>8))
		for _, v := range big.Kids {
			buf = append(buf,
				byte(len(v)),
				byte(len(v)>>8))
			buf = append(buf, []byte(v)...)
		}
		// Touches
		buf = append(buf,
			byte(uint32(big.Touches)),
			byte(uint32(big.Touches)>>8),
			byte(uint32(big.Touches)>>16),
			byte(uint32(big.Touches)>>24))

		// MinT
		buf = append(buf,
			byte(uint64(big.MinT)),
			byte(uint64(big.MinT)>>8),
			byte(uint64(big.MinT)>>16),
			byte(uint64(big.MinT)>>24),
			byte(uint64(big.MinT)>>32),
			byte(uint64(big.MinT)>>40),
			byte(uint64(big.MinT)>>48),
			byte(uint64(big.MinT)>>56))

		// MaxT
		buf = append(buf,
			byte(uint64(big.MaxT)),
			byte(uint64(big.MaxT)>>8),
			byte(uint64(big.MaxT)>>16),
			byte(uint64(big.MaxT)>>24),
			byte(uint64(big.MaxT)>>32),
			byte(uint64(big.MaxT)>>40),
			byte(uint64(big.MaxT)>>48),
			byte(uint64(big.MaxT)>>56))
		// MeanT
		buf = append(buf,
			byte(uint64(big.MeanT)),
			byte(uint64(big.MeanT)>>8),
			byte(uint64(big.MeanT)>>16),
			byte(uint64(big.MeanT)>>24),
			byte(uint64(big.MeanT)>>32),
			byte(uint64(big.MeanT)>>40),
			byte(uint64(big.MeanT)>>48),
			byte(uint64(big.MeanT)>>56))
	}
}

func BenchmarkBigManualPreAlloc(b *testing.B) {
	b.SetBytes(53283)
	for i := 0; i < b.N; i++ {
		size := 32 + len(big.Name)
		for _, v := range big.Kids {
			size += len(v) + 2
		}
		buf := make([]byte, size)
		offset := 0
		// Name
		buf[offset] = byte(len(big.Name))
		buf[offset+1] = byte(len(big.Name) >> 8)
		offset += 2
		copy(buf[offset:], big.Name)
		offset += len(big.Name)
		// Kids
		buf[offset] = byte(len(big.Kids))
		buf[offset+1] = byte(len(big.Kids) >> 8)
		offset += 2
		for _, v := range big.Kids {
			buf[offset] = byte(len(v))
			buf[offset+1] = byte(len(v) >> 8)
			offset += 2
			copy(buf[offset:], v)
			offset += len(v)
		}
		// Touches
		buf[offset] = byte(uint32(big.Touches))
		buf[offset+1] = byte(uint32(big.Touches) >> 8)
		buf[offset+2] = byte(uint32(big.Touches) >> 16)
		buf[offset+3] = byte(uint32(big.Touches) >> 24)
		offset += 4
		// MinT
		buf[offset] = byte(uint64(big.MinT))
		buf[offset+1] = byte(uint64(big.MinT) >> 8)
		buf[offset+2] = byte(uint64(big.MinT) >> 16)
		buf[offset+3] = byte(uint64(big.MinT) >> 24)
		buf[offset+4] = byte(uint64(big.MinT) >> 32)
		buf[offset+5] = byte(uint64(big.MinT) >> 40)
		buf[offset+6] = byte(uint64(big.MinT) >> 48)
		buf[offset+7] = byte(uint64(big.MinT) >> 56)
		offset += 8
		// MaxT
		buf[offset] = byte(uint64(big.MaxT))
		buf[offset+1] = byte(uint64(big.MaxT) >> 8)
		buf[offset+2] = byte(uint64(big.MaxT) >> 16)
		buf[offset+3] = byte(uint64(big.MaxT) >> 24)
		buf[offset+4] = byte(uint64(big.MaxT) >> 32)
		buf[offset+5] = byte(uint64(big.MaxT) >> 40)
		buf[offset+6] = byte(uint64(big.MaxT) >> 48)
		buf[offset+7] = byte(uint64(big.MaxT) >> 56)
		offset += 8
		// MeanT
		buf[offset] = byte(uint64(big.MeanT))
		buf[offset+1] = byte(uint64(big.MeanT) >> 8)
		buf[offset+2] = byte(uint64(big.MeanT) >> 16)
		buf[offset+3] = byte(uint64(big.MeanT) >> 24)
		buf[offset+4] = byte(uint64(big.MeanT) >> 32)
		buf[offset+5] = byte(uint64(big.MeanT) >> 40)
		buf[offset+6] = byte(uint64(big.MeanT) >> 48)
		buf[offset+7] = byte(uint64(big.MeanT) >> 56)
		offset += 8
	}
}

func BenchmarkSmallGob(b *testing.B) {
	b.SetBytes(58)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		enc := gob.NewEncoder(io.Discard)
		enc.Encode(s)
	}
}

func BenchmarkBigGob(b *testing.B) {
	b.SetBytes(53283)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		enc := gob.NewEncoder(io.Discard)
		enc.Encode(big)
	}
}

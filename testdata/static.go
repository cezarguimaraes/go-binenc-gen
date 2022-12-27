package main

import (
	"bytes"

	"github.com/google/go-cmp/cmp"
)

//go:generate go-binenc-gen static.go
type Static struct {
	Uint8  uint8
	Uint16 uint16
	Uint32 uint32
	Uint64 uint64
	Int8   int8
	Int16  int16
	Int32  int32
	Int64  int64
	Arr    [4]uint8
}

func main() {
	s := &Static{
		Uint8:  1,
		Uint16: 2,
		Uint32: 3,
		Uint64: 4,
		Int8:   -1,
		Int16:  -2,
		Int32:  -3,
		Int64:  -4,
		Arr:    [4]uint8{1, 2, 3, 4},
	}

	var buf bytes.Buffer
	s.WriteTo(&buf)

	o := new(Static)
	o.ReadFrom(&buf)

	if diff := cmp.Diff(s, o); diff != "" {
		panic("static.go: \n" + diff)
	}
}

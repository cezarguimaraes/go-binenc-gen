package main

import (
	"bytes"

	"github.com/google/go-cmp/cmp"
)

//go:generate go-binenc-gen float.go
type Float struct {
	Float32 float32
	Float64 float64
}

func main() {
	s := &Float{
		Float32: 4.2,
		Float64: -136.3737,
	}

	var buf bytes.Buffer
	s.Write(&buf)

	o := new(Float)
	o.Read(&buf)

	if diff := cmp.Diff(s, o); diff != "" {
		panic("float.go: \n" + diff)
	}
}

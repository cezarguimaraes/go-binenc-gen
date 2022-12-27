package main

import (
	"bytes"

	"github.com/google/go-cmp/cmp"
)

//go:generate go-binenc-gen slice.go
type Slice struct {
	Int8Slice []int8
}

func main() {
	s := &Slice{
		Int8Slice: []int8{1, 2, 3, 4},
	}

	var buf bytes.Buffer
	s.Write(&buf)

	// fmt.Println("% x", buf.Bytes())

	o := new(Slice)
	o.Read(&buf)

	// fmt.Printf("%+v", o)

	if diff := cmp.Diff(s, o); diff != "" {
		panic("slice.go: \n" + diff)
	}
}

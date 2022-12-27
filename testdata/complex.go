package main

import (
	"bytes"

	"github.com/google/go-cmp/cmp"
)

//go:generate go-binenc-gen complex.go
type Complex struct {
	Complex64  complex64
	Complex128 complex128
}

func main() {
	s := &Complex{
		Complex32: 4.2 - 28i,
		Complex64: -136.3737 + 30e2i,
	}

	var buf bytes.Buffer
	s.Write(&buf)

	o := new(Complex)
	o.Read(&buf)

	if diff := cmp.Diff(s, o); diff != "" {
		panic("complex.go: \n" + diff)
	}
}

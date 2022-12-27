package main

import (
	"bytes"

	"github.com/google/go-cmp/cmp"
)

//go:generate go-binenc-gen embed_blank.go
type Innermost struct {
	Foo, Bar uint8
	_        uint8
}

type Inner struct {
	Num  uint8
	Arr4 []string
	Innermost
	_ string
}

type Outer struct {
	Foo uint8

	Inner

	_ uint64
}

func (o Outer) Method() error {
	return nil
}

func main() {
	s := &Outer{
		Foo: 42,
		Inner: Inner{
			Num:  1,
			Arr4: []string{"a", "b", "c"},
			Innermost: Innermost{
				Foo: 5,
				Bar: 8,
			},
		},
	}

	var buf bytes.Buffer
	s.WriteTo(&buf)

	o := new(Outer)
	o.ReadFrom(&buf)

	if diff := cmp.Diff(s, o); diff != "" {
		panic("embed_blank.go: \n" + diff)
	}
}

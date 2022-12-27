package main

import (
	"bytes"

	"github.com/google/go-cmp/cmp"
)

//go:generate go-binenc-gen soa.go
type Inner struct {
	Str  string
	Arr3 []uint8
	Arr4 []string
}

type Outer struct {
	Arr1   []uint8
	Inners []Inner
	Arr2   []int8
}

func main() {
	s := &Outer{
		Arr1: []uint8{1, 2, 3},
		Inners: []Inner{
			{
				Str:  "str1",
				Arr3: []uint8{3, 4, 5},
				Arr4: []string{"is1", "is2"},
			},
			{
				Str:  "str2",
				Arr3: []uint8{6, 7, 8, 9, 10},
				Arr4: []string{"is3", "is4", "is5"},
			},
		},
		Arr2: []int8{-1, -2, -3},
	}

	var buf bytes.Buffer
	s.Write(&buf)

	// fmt.Println("% x", buf.Bytes())

	o := new(Outer)
	o.Read(&buf)

	// fmt.Printf("%+v", o)

	if diff := cmp.Diff(s, o); diff != "" {
		panic("soa.go: \n" + diff)
	}
}

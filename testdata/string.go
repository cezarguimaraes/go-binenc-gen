package main

import (
	"bytes"

	"github.com/google/go-cmp/cmp"
)

//go:generate go-binenc-gen string.go
type String struct {
	S string
}

func main() {
	s := &String{
		S: "foo bar",
	}

	var buf bytes.Buffer
	s.WriteTo(&buf)

	o := new(String)
	o.ReadFrom(&buf)

	if diff := cmp.Diff(s, o); diff != "" {
		panic("string.go: \n" + diff)
	}
}

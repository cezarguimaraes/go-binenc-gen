package slice

//go:generate go-binenc-gen slice.go
type Slice struct {
	Int8Slice []int8
}

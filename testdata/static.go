package static

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
}

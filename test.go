package main

type SChannelEvent struct {
	OPCode       uint8
	ChannelId    uint16
	PlayerName   string
	ChannelEvent uint8
	TestInt16    int16
	TestUint32   uint32
	TestInt32    int32
	TestUint64   uint64
	TestInt64    int64
	TestPointer  *int8
	TestPInt16   *int16
}

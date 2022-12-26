package string

//go:generate go-binenc-gen string.go
type String struct {
	S string
}

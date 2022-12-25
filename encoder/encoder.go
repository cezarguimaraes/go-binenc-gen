package encoder

import (
	"bytes"
	"fmt"
	"go/types"
	"log"
)

const (
	byteFmt          = "\tbuf[%s] = byte(%s)\n"
	staticIndex      = "offset"
	initOffset       = "\t" + staticIndex + " := 0\n"
	incrOffsetPrefix = "\t" + staticIndex + " += "
	incrOffsetFmt    = incrOffsetPrefix + "%d\n"
	dynOffsetFmt     = incrOffsetPrefix + "%s\n"
	indexFmt         = staticIndex + " + %d"
	unsignedCastFmt  = "uint%d(%s)"
	rshiftFmt        = "%s >> %d"
	lenFmt           = "len(%s)"
	copyFmt          = "\tcopy(buf[" + staticIndex + ":], %s)\n"
	booleanFmt       = "\tif %s {\n\t" + byteFmt + "\telse {\n\t" + byteFmt + "\t}\n"
)

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func length(name string) string {
	return fmt.Sprintf(lenFmt, name)
}

func index(offset int) string {
	if offset == 0 {
		return staticIndex
	}
	return fmt.Sprintf(indexFmt, offset)
}

func rshift(name string, n int) string {
	if n == 0 {
		return name
	}
	return fmt.Sprintf(rshiftFmt, name, n*8)
}

type Writer struct {
	buf *bytes.Buffer

	stdSizes     *types.StdSizes
	dynamicSizes []string
	size         int
	bigEndian    bool
}

func NewWriter() *Writer {
	enc := &Writer{
		buf: &bytes.Buffer{},
		stdSizes: &types.StdSizes{
			WordSize: 4,
			MaxAlign: 4,
		},
	}
	return enc
}

func (w *Writer) Printf(format string, args ...interface{}) {
	fmt.Fprintf(w.buf, format, args...)
}

func (w *Writer) addOffset(size int) {
	w.Printf(incrOffsetFmt, size)
	w.size += size
}

func (w *Writer) addDynamicOffset(name string) {
	w.Printf(dynOffsetFmt, name)
	w.dynamicSizes = append(w.dynamicSizes, name)
}

func (w *Writer) writeByte(offset int, name string, incrOffset bool) {
	w.Printf(byteFmt, index(offset), name)
	if incrOffset {
		w.addOffset(1)
	}
}

func (w *Writer) writeString(name string) {
	strLength := length(name)
	w.writeNumberN(strLength, 2, true)
	w.Printf(copyFmt, name)
}

func (w *Writer) writeNumberN(name string, nbytes int, unsigned bool) {
	if !unsigned {
		name = fmt.Sprintf(unsignedCastFmt, 8*nbytes, name)
	}
	start, end, incr := 0, nbytes, 1
	if w.bigEndian {
		start, end, incr = nbytes-1, -1, -1
	}
	for i := start; i != end; i += incr {
		w.writeByte(abs(i-start), rshift(name, i), false)
	}
	w.addOffset(nbytes)
}

func (w *Writer) writeBoolean(name string) {

}

func (w *Writer) WriteField(name string, t types.Type) {
	switch f := t.Underlying().(type) {
	case *types.Basic:
		info := f.Info()
		if info&types.IsInteger != 0 {
			unsigned := info&types.IsUnsigned != 0
			size := w.stdSizes.Sizeof(f)
			w.writeNumberN(name, int(size), unsigned)
		} else if info&types.IsBoolean != 0 {

		} else {
			log.Printf("unknown type: %s\n", f.Name())
		}
	default:
		log.Printf("unknown type: %T\n", f)
	}

}

func (w *Writer) Bytes() []byte {
	return w.buf.Bytes()
}

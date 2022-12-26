package encoder

import (
	"bytes"
	"fmt"
	"go/types"
	"io"
	"log"
	"strings"
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
	booleanFmt       = "\tif %s {\n\t" + byteFmt + "\t} else {\n\t" + byteFmt + "\t}\n"
	forStartFmt      = "\tfor _, %s := range %s {\n"
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
}

func (w *Writer) writeString(name string) {
	strLen := length(name)
	w.writeNumberN(strLen, 2, true)
	w.Printf(copyFmt, name)
	w.addDynamicOffset(strLen)
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
	w.Printf(booleanFmt, name, staticIndex, "0x01", staticIndex, "0x00")
}

func (w *Writer) WriteField(name string, t types.Type) {
	w.writeField(name, t, 0)
}

func rangeForVar(lvl int) string {
	if lvl == 0 {
		return "v"
	}
	return fmt.Sprintf("v%d", lvl)
}

func (w *Writer) writeField(name string, t types.Type, forLvl int) {
	t = t.Underlying()
	if ptr, ok := t.(*types.Pointer); ok {
		w.WriteField("*"+name, ptr.Elem())
		return
	}
	if slc, ok := t.(*types.Slice); ok {
		// TODO: specially handle []byte
		slcLen := length(name)
		w.writeNumberN(slcLen, 2, true)
		w.Printf(forStartFmt, rangeForVar(forLvl), name)
		nxtForLvl := forLvl + 1
		w.writeField(rangeForVar(forLvl), slc.Elem(), nxtForLvl)
		w.Printf("\t}\n")
	}
	// TODO: handle structs
	switch f := t.(type) {
	case *types.Basic:
		info := f.Info()
		if info&types.IsInteger != 0 {
			unsigned := info&types.IsUnsigned != 0
			size := w.stdSizes.Sizeof(f)
			w.writeNumberN(name, int(size), unsigned)
		} else if info&types.IsBoolean != 0 {
			w.writeBoolean(name)
		} else if info&types.IsString != 0 {
			w.writeString(name)
		} else {
			log.Printf("unknown type: %s\n", f.Name())
		}
	default:
		log.Printf("unknown type: %T\n", f)
	}
}

func (w *Writer) SizeExpr() string {
	expr := fmt.Sprintf("%d", w.size)
	if len(w.dynamicSizes) > 0 {
		expr += "+" + strings.Join(w.dynamicSizes, "+")
	}
	return expr
}

func (w *Writer) Bytes() []byte {
	return w.buf.Bytes()
}

func (w *Writer) WriteTo(writer io.Writer) (n int64, err error) {
	return w.buf.WriteTo(writer)
}

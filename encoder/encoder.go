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

	stdSizes *types.StdSizes

	sizeExprs    []string
	dynamicSizes [][]string
	sizes        []int
	forLvl       int

	bigEndian bool
}

func NewWriter() *Writer {
	enc := &Writer{
		buf: &bytes.Buffer{},
		stdSizes: &types.StdSizes{
			WordSize: 4,
			MaxAlign: 4,
		},
		forLvl: -1,
	}
	enc.pushForLvl()
	return enc
}

func (w *Writer) Printf(format string, args ...interface{}) {
	fmt.Fprintf(w.buf, format, args...)
}

func (w *Writer) addOffset(size int) {
	w.Printf(incrOffsetFmt, size)
	w.sizes[w.forLvl] += size
}

func (w *Writer) addDynamicOffset(name string) {
	w.Printf(dynOffsetFmt, name)
	w.dynamicSizes[w.forLvl] = append(w.dynamicSizes[w.forLvl], name)
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

func (w *Writer) pushForLvl() {
	w.sizes = append(w.sizes, 0)
	w.dynamicSizes = append(w.dynamicSizes, []string{})
	w.forLvl += 1
	if w.forLvl > 0 {
		w.sizeExprs = append(w.sizeExprs, "}\n")
	}
}

func (w *Writer) popForLvl(name string) {
	// exprs are inserted in reverse order
	defer func() {
		w.dynamicSizes = w.dynamicSizes[:w.forLvl]
		w.sizes = w.sizes[:w.forLvl]
		w.forLvl -= 1
	}()

	// avoid adding a for that only sums a constant
	// also avoids compiler error for unused variable
	if len(w.dynamicSizes[w.forLvl]) == 0 && w.forLvl > 0 {
		// remove closing bracket, not sure if this is sound
		w.sizeExprs = w.sizeExprs[:len(w.sizeExprs)-1]

		w.dynamicSizes[w.forLvl-1] = append(w.dynamicSizes[w.forLvl-1],
			fmt.Sprintf("%d * len(%s)", w.sizes[w.forLvl], name),
		)
		return
	}

	// add dynamic sizes
	if len(w.dynamicSizes[w.forLvl]) > 0 {
		w.sizeExprs = append(w.sizeExprs,
			fmt.Sprintf(
				"\tsize += %s\n",
				strings.Join(w.dynamicSizes[w.forLvl], " + "),
			),
		)
	}

	// add static sizes
	sizeOp := "+="
	if w.forLvl == 0 {
		sizeOp = ":="
	}
	w.sizeExprs = append(w.sizeExprs,
		fmt.Sprintf(
			"\tsize %s %d\n",
			sizeOp,
			w.sizes[w.forLvl],
		),
	)
	if w.forLvl > 0 {
		w.sizeExprs = append(w.sizeExprs,
			fmt.Sprintf(forStartFmt, rangeForVar(w.forLvl), name),
		)
	}
}

func (w *Writer) WriteField(name string, t types.Type) {
	w.writeField(name, t)
}

func rangeForVar(lvl int) string {
	if lvl == 1 {
		return "v"
	}
	return fmt.Sprintf("v%d", lvl-1)
}

func (w *Writer) writeField(name string, t types.Type) {
	t = t.Underlying()
	if ptr, ok := t.(*types.Pointer); ok {
		w.WriteField("*"+name, ptr.Elem())
		return
	}
	if slc, ok := t.(*types.Slice); ok {
		// TODO: specially handle []byte
		slcLen := length(name)
		w.writeNumberN(slcLen, 2, true)
		w.pushForLvl()
		w.Printf(forStartFmt, rangeForVar(w.forLvl), name)
		w.writeField(rangeForVar(w.forLvl), slc.Elem())
		w.popForLvl(name)
		w.Printf("\t}\n")
		return
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
	if w.forLvl == 0 {
		w.popForLvl("ignored")
	}
	res := make([]string, len(w.sizeExprs))
	for i := 0; i < len(w.sizeExprs); i++ {
		res[i] = w.sizeExprs[len(w.sizeExprs)-i-1]
	}
	return strings.Join(res, "")
}

func (w *Writer) Bytes() []byte {
	return w.buf.Bytes()
}

func (w *Writer) WriteTo(writer io.Writer) (n int64, err error) {
	return w.buf.WriteTo(writer)
}

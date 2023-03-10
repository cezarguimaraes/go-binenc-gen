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
	lshiftFmt        = "(%s << %d)"
	lenFmt           = "len(%s)"
	copyFmt          = "\tcopy(buf[" + staticIndex + ":], %s)\n"
	booleanFmt       = "\tif %s {\n\t" + byteFmt + "\t} else {\n\t" + byteFmt + "\t}\n"
	forStartFmt      = "\tfor _, %s := range %s {\n"
	readBytesFmt     = "\tr.Read(buf[:%d])\n"
	bytesCastFmt     = "(*(*[%d]byte)(unsafe.Pointer(&(%s))))[:]"
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

func lshift(name string, n int) string {
	if n == 0 {
		return name
	}
	return fmt.Sprintf(lshiftFmt, name, n*8)
}

type Writer struct {
	buf *bytes.Buffer

	stdSizes *types.StdSizes
	pkg      *types.Package

	sizeExprs    []string
	dynamicSizes [][]string
	sizes        []int
	forLvl       int

	bigEndian bool

	strBufCount int
	usedSize    bool
	usedBuffer  bool
	needUnsafe  bool
}

func NewWriter(pkg *types.Package) *Writer {
	enc := &Writer{
		buf: &bytes.Buffer{},
		stdSizes: &types.StdSizes{
			WordSize: 4,
			MaxAlign: 4,
		},
		pkg:    pkg,
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

func (w *Writer) popForLvl(name string, static bool, staticLen int) {
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

		if !static {
			w.dynamicSizes[w.forLvl-1] = append(w.dynamicSizes[w.forLvl-1],
				fmt.Sprintf("%d * len(%s)", w.sizes[w.forLvl], name),
			)
		} else {
			w.sizes[w.forLvl-1] += w.sizes[w.forLvl] * staticLen
		}
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
		if !static {
			w.sizeExprs = append(w.sizeExprs,
				fmt.Sprintf(forStartFmt, rangeForVar(w.forLvl), name),
			)
		} else {
			w.sizeExprs = append(w.sizeExprs,
				fmt.Sprintf("\tfor %s := 0; %s < %d; %s++ {\n", indexForVar(w.forLvl), indexForVar(w.forLvl), staticLen, indexForVar(w.forLvl)),
			)
		}
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

func (w *Writer) writeBytes(name string, nbytes int) {
	w.needUnsafe = true
	w.Printf(copyFmt, fmt.Sprintf(bytesCastFmt, nbytes, name))
	w.addOffset(nbytes)
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
		w.popForLvl(name, false, 0)
		w.Printf("\t}\n")
		return
	}
	if s, ok := t.(*types.Struct); ok {
		for i := 0; i < s.NumFields(); i++ {
			f := s.Field(i)
			if f.Name() == "_" {
				continue
			}
			selector := fmt.Sprintf("%s.%s", name, f.Name())
			w.writeField(selector, f.Type())
		}
		return
	}
	if arr, ok := t.(*types.Array); ok {
		w.pushForLvl()
		w.Printf("\tfor %s := 0; %s < %d; %s++ {\n", indexForVar(w.forLvl), indexForVar(w.forLvl), arr.Len(), indexForVar(w.forLvl))
		w.writeField(fmt.Sprintf("%s[%s]", name, indexForVar(w.forLvl)), arr.Elem())
		w.popForLvl(name, true, int(arr.Len()))
		w.Printf("\t}\n")
		return
	}
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
		} else if info&types.IsFloat != 0 || info&types.IsComplex != 0 {
			size := w.stdSizes.Sizeof(f)
			w.writeBytes(name, int(size))
		} else {
			log.Printf("unknown type: %s\n", f.Name())
		}
	default:
		log.Printf("unknown type: %T\n", f)
	}
}

func (w *Writer) readNumberN(name string, nbytes int, unsigned bool) {
	w.usedBuffer = true
	w.Printf(readBytesFmt, nbytes)
	exprParts := []string{}
	start, end, incr := 0, nbytes, 1
	if w.bigEndian {
		start, end, incr = nbytes-1, -1, -1
	}
	for i := start; i != end; i += incr {
		b := fmt.Sprintf("buf[%d]", abs(i-start))
		b = fmt.Sprintf(unsignedCastFmt, 8*nbytes, b)
		exprParts = append(exprParts, lshift(b, i))
	}
	expr := strings.Join(exprParts, " | ")
	if !unsigned {
		expr = fmt.Sprintf("int%d(%s)", 8*nbytes, expr)
	}
	w.Printf("\t%s = %s\n", name, expr)
}

func (w *Writer) readBoolean(name string) {
	w.Printf(readBytesFmt, 1)
	w.Printf("\tif buf[0] == byte(0x01) {\n")
	w.Printf("\t%s = true\n", name)
	w.Printf("} else {\n")
	w.Printf("\t%s = false\n", name)
	w.Printf("}\n")
}

func (w *Writer) readString(name string) {
	w.needUnsafe = true
	w.usedSize = true
	w.readNumberN("size", 2, true)
	w.Printf("\tif c - m < int(size) {\n")
	w.Printf("\tc = int(size)\n")
	w.Printf("\tif c < 2*cap(strBuf) {\n")
	w.Printf("\tc = 2*cap(strBuf)\n")
	w.Printf("\t}\n")
	// based on bytes.Buffer.growSlice:
	// https://cs.opensource.google/go/go/+/refs/tags/go1.19.4:src/bytes/buffer.go;l=240
	w.Printf("\tstrBuf = append([]byte(nil), make([]byte, c)...)\n")
	// do we need to copy previous bytes here??
	w.Printf("\tm = 0\n")
	w.Printf("\t}\n")
	w.Printf("\tr.Read(strBuf[m:m+int(size)])\n")
	w.Printf("\ttmp = strBuf[m:m+int(size)]\n")
	// based on strings.Builder.Strings
	// https://cs.opensource.google/go/go/+/refs/tags/go1.19.4:src/strings/builder.go;l=48
	w.Printf("\t%s = *(*string)(unsafe.Pointer(&tmp))\n", name)
	w.Printf("\tm += int(size)\n")
	w.strBufCount += 1
}

func indexForVar(lvl int) string {
	if lvl == 0 {
		return "i"
	}
	return fmt.Sprintf("i%d", lvl)
}

func indexForSize(lvl int) string {
	if lvl == 0 {
		return "si"
	}
	return fmt.Sprintf("si%d", lvl)
}

func (w *Writer) HeaderExpr() string {
	var lines []string
	if w.usedBuffer {
		// TODO: use smallest buffer possible
		lines = append(lines, "buf := make([]byte, 8)\n")
	}
	if w.usedSize {
		lines = append(lines, "var size uint16\n")
	}
	if w.strBufCount > 0 {
		lines = append(lines, "var tmp []byte\n", "m := 0\n", "c := 64\n", "strBuf := make([]byte, c)\n")
	}
	return strings.Join(lines, "")
}

func (w *Writer) typeName(t types.Type) string {
	return types.TypeString(t, types.RelativeTo(w.pkg))
}

func (w *Writer) readBytes(name string, nbytes int) {
	w.needUnsafe = true
	w.Printf("\tr.Read(%s)\n", fmt.Sprintf(bytesCastFmt, nbytes, name))
}

func (w *Writer) ReadField(name string, t types.Type) {
	t = t.Underlying()
	if ptr, ok := t.(*types.Pointer); ok {
		w.Printf("\t%s = new(%s)\n", name, w.typeName(ptr.Elem()))
		w.ReadField("*"+name, ptr.Elem())
		return
	}
	if slc, ok := t.(*types.Slice); ok {
		// TODO: specially handle []byte
		w.usedSize = true
		slcLen := "size"
		w.readNumberN(slcLen, 2, true)
		w.Printf("\t%s = make(%s, %s)\n", name, w.typeName(slc), slcLen)
		// intentionally never decrease forLvl
		// to never reuse index variables
		w.Printf("\t%s := int(%s)\n", indexForSize(w.forLvl), slcLen)
		w.Printf("\tfor %s := 0; %s < %s; %s++ {\n", indexForVar(w.forLvl), indexForVar(w.forLvl), indexForSize(w.forLvl), indexForVar(w.forLvl))
		w.forLvl += 1
		w.ReadField(fmt.Sprintf("%s[%s]", name, indexForVar(w.forLvl-1)), slc.Elem())
		w.Printf("\t}\n")
		return
	}
	// TODO: add tests for read
	if s, ok := t.(*types.Struct); ok {
		for i := 0; i < s.NumFields(); i++ {
			f := s.Field(i)
			if f.Name() == "_" {
				continue
			}
			selector := fmt.Sprintf("%s.%s", name, f.Name())
			w.ReadField(selector, f.Type())
		}
		return
	}
	if arr, ok := t.(*types.Array); ok {
		w.Printf("\tfor %s := 0; %s < %d; %s++ {\n", indexForVar(w.forLvl), indexForVar(w.forLvl), arr.Len(), indexForVar(w.forLvl))
		w.forLvl += 1
		w.ReadField(fmt.Sprintf("%s[%s]", name, indexForVar(w.forLvl-1)), arr.Elem())
		w.Printf("\t}\n")
		return
	}
	switch f := t.(type) {
	case *types.Basic:
		info := f.Info()
		if info&types.IsInteger != 0 {
			unsigned := info&types.IsUnsigned != 0
			size := w.stdSizes.Sizeof(f)
			w.readNumberN(name, int(size), unsigned)
		} else if info&types.IsBoolean != 0 {
			w.readBoolean(name)
		} else if info&types.IsString != 0 {
			w.readString(name)
		} else if info&types.IsFloat != 0 || info&types.IsComplex != 0 {
			size := w.stdSizes.Sizeof(f)
			w.readBytes(name, int(size))
		} else {
			log.Printf("unknown type: %s\n", f.Name())
		}
	default:
		log.Printf("unknown type: %T\n", f)
	}
}

func (w *Writer) SizeExpr() string {
	if w.forLvl == 0 {
		w.popForLvl("ignored", false, -1)
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

func (w *Writer) NeedUnsafe() bool {
	return w.needUnsafe
}

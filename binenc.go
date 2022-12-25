package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"go/types"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"golang.org/x/tools/go/packages"
)

func main() {
	mInt8 := int8(12)
	mInt16 := int16(255 | (4 << 8))
	sce := &SChannelEvent{
		OPCode:       0x68,
		ChannelId:    0x01,
		PlayerName:   "cezar",
		ChannelEvent: 0x00,
		TestPointer:  &mInt8,
		TestPInt16:   &mInt16,
	}
	var ignore bytes.Buffer
	sce.Write(&ignore)
	// 68 01 00 05 00 63 65 7a 61 72 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 0c ff 04

	log.SetFlags(0)
	log.SetPrefix("binenc: ")

	flag.Parse()
	tags := []string{}
	args := flag.Args()
	if len(args) == 0 {
		args = []string{"test.go"}
	}

	g := &Generator{}
	g.parsePackage(args, tags)

	g.Printf("// Code generated by \"gobinenc %s\"; DO NOT EDIT.\n", strings.Join(os.Args[1:], " "))
	g.Printf("\n")
	g.Printf("package %s", g.pkg.name)
	g.Printf("\n")

	g.Printf("import (\n")
	g.Printf("\t\"encoding/binary\"\n")
	g.Printf("\t\"io\"\n")
	g.Printf("\t\"fmt\"\n")
	g.Printf(")")
	g.Printf("\n")

	g.generate()

	src := g.format()
	err := ioutil.WriteFile(fmt.Sprintf("%s_encoding.go", g.pkg.name), src, 0644)
	if err != nil {
		log.Fatalf("writing output: %s", err)
	}
}

type Struct struct {
	name   string
	fields []string
	types  []types.Type
}

type File struct {
	pkg      *Package
	file     *ast.File
	typeName string
	// values   []Value
	structs []*Struct
}

type Package struct {
	name     string
	typeInfo *types.Info
	files    []*File
}

type Generator struct {
	buf bytes.Buffer
	pkg *Package
}

func (g *Generator) parsePackage(patterns, tags []string) {
	cfg := &packages.Config{
		Mode:       packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax,
		Tests:      false,
		BuildFlags: []string{fmt.Sprintf("-tags=%s", strings.Join(tags, " "))},
	}

	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		log.Fatal(err)
	}
	if len(patterns) != 1 {
		log.Fatalf("error: %d packages found", len(pkgs))
	}

	g.addPackage(pkgs[0])
}

func (g *Generator) addPackage(pkg *packages.Package) {
	g.pkg = &Package{
		name:     pkg.Name,
		typeInfo: pkg.TypesInfo,
		files:    make([]*File, len(pkg.Syntax)),
	}

	for i, file := range pkg.Syntax {
		g.pkg.files[i] = &File{
			file: file,
			pkg:  g.pkg,
		}
	}
}

func (g *Generator) generate() {
	for _, file := range g.pkg.files {
		log.Printf("generating file %s\n", file.file.Name)
		if file.file != nil {
			ast.Inspect(file.file, file.inspectNode)
		}
	}

	for _, file := range g.pkg.files {
		for _, s := range file.structs {
			g.generateWrite(s)
			g.generateRead(s)
		}
	}
}

func (g *Generator) Printf(format string, args ...interface{}) {
	fmt.Fprintf(&g.buf, format, args...)
}

func (g *Generator) format() []byte {
	src, err := format.Source(g.buf.Bytes())
	if err != nil {
		log.Printf("warning: internal error: invalid Go generated: %s", err)
		log.Printf("warning: compile the package to analyze the error")
		return g.buf.Bytes()
	}
	return src
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (g *Generator) generateWrite(s *Struct) {
	g.Printf("func (s *%s) Write(w io.Writer) error {\n", s.name)
	// TODO: initialize buffer with required capacity
	//g.Printf("\tbuf := make([]byte, 256)\n")
	g.Printf("\toffset := 0\n")
	var buf bytes.Buffer
	staticSize := 0
	dynamicSizes := []string{}
	bigEndian := false
	for i, name := range s.fields {
		selector := fmt.Sprintf("s.%s", name)
		fmt.Fprintf(&buf, "\t// %s\n", name)
		t := s.types[i]
		if ptr, ok := t.(*types.Pointer); ok {
			selector = "*" + selector
			t = ptr.Elem()
		}
		switch t {
		case types.Typ[types.Bool]:
		case types.Typ[types.Uint8]:
			writeNumberN(&buf, selector, 1, true, &staticSize, bigEndian)
			break
		case types.Typ[types.Int8]:
			writeNumberN(&buf, selector, 1, false, &staticSize, bigEndian)
			break
		case types.Typ[types.Uint16]:
			writeNumberN(&buf, selector, 2, true, &staticSize, bigEndian)
			break
		case types.Typ[types.Int16]:
			writeNumberN(&buf, selector, 2, false, &staticSize, bigEndian)
			break
		case types.Typ[types.Uint32]:
			writeNumberN(&buf, selector, 4, true, &staticSize, bigEndian)
			break
		case types.Typ[types.Int32]:
			writeNumberN(&buf, selector, 4, false, &staticSize, bigEndian)
			break
		case types.Typ[types.Uint64]:
			writeNumberN(&buf, selector, 8, true, &staticSize, bigEndian)
			break
		case types.Typ[types.Int64]:
			writeNumberN(&buf, selector, 8, false, &staticSize, bigEndian)
			break
		case types.Typ[types.String]:
			dynSize := fmt.Sprintf("len(%s)", selector)
			writeNumberN(&buf, dynSize, 2, true, &staticSize, bigEndian)
			fmt.Fprintf(&buf, "\tcopy(buf[offset:], %s)\n", selector)
			fmt.Fprintf(&buf, "\toffset += %s\n", dynSize)
			dynamicSizes = append(dynamicSizes, dynSize)
			break
		default:
			//g.Printf("\tbinary.Write(&buf, binary.LittleEndian, s.%s)\n", name)
			fmt.Fprintf(&buf, "\t// unsupported type\n")
			log.Printf("warning: type not supported %T\n", t)
		}
	}
	fmt.Fprintf(&buf, "\tfmt.Printf(\"%% x\\n\", buf)\n")
	fmt.Fprintf(&buf, "\t_, err := w.Write(buf)\n")
	fmt.Fprintf(&buf, "\treturn err\n")
	fmt.Fprintf(&buf, "}\n\n")

	dynSizeExpr := ""
	if len(dynamicSizes) > 0 {
		dynSizeExpr = "+" + strings.Join(dynamicSizes, "+")
	}
	g.Printf("\tbuf := make([]byte, %d%s)\n", staticSize, dynSizeExpr)
	buf.WriteTo(&g.buf)
}

func (g *Generator) generateRead(s *Struct) {
	g.Printf("func (s *%s) Read(r io.Reader) error {\n", s.name)
	for _, name := range s.fields {
		g.Printf("\tbinary.Read(r, binary.LittleEndian, s.%s)\n", name)
	}
	g.Printf("\treturn nil\n")
	g.Printf("}\n\n")
}

func (f *File) inspectNode(node ast.Node) bool {
	decl, ok := node.(*ast.GenDecl)
	if !ok || decl.Tok != token.TYPE {
		return true
	}
	for _, spec := range decl.Specs {
		tspec, ok := spec.(*ast.TypeSpec)
		if !ok {
			log.Printf("not type spec")
			break
		}
		if tspec.Type == nil {
			continue
		}
		st, ok := tspec.Type.(*ast.StructType)
		if !ok || st.Fields == nil || st.Fields.List == nil {
			log.Printf("not struct type or missing field list")
			continue
		}
		s := &Struct{name: tspec.Name.Name}
		for _, field := range st.Fields.List {
			if len(field.Names) != 1 {
				log.Printf("warning: ignoring field because len(field.Names) != 1: %s\n", field.Names)
				continue
			}
			s.fields = append(s.fields, field.Names[0].Name)
			s.types = append(s.types, f.pkg.typeInfo.TypeOf(field.Type))
		}
		f.structs = append(f.structs, s)
	}
	return false
}

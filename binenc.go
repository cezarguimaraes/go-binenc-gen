// Copyright (c) 2009 The Go Authors. All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//    * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//    * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//    * Neither the name of Google Inc. nor the names of its
// contributors may be used to endorse or promote products derived from
// this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
//
// Source adapted from https://cs.opensource.google/go/x/tools/+/refs/tags/v0.4.0:cmd/stringer/stringer.go

// go-binenc-gen is a tool to automate the creation of binary serialization methods.
// Given the name of a Go source file containing structs definitions, go-binenc-gen
// will create a new self-contained Go source file implementing
//
//	func (s *T) Write(w io.Writer) (n int, err error)
//	func (s *T) Read(r io.Reader) (err error)
//
// The file is created in the same package and directory as the package that defines
// T. It has helpful defaults designed for use with go generate.
//
// The goal of the tool is to generate extremely fast binary serialization code by
// forfeiting the convenience of runtime reflection enabled libraries, such as
// binary or gob.
//
// For example, given this snippet,
//
//	package example
//
//	type Header struct {
//		Name string
//		Value string
//	}
//
//	type Request struct {
//		Headers []Header
//		ResponseTime uin64
//	}
//
// running this command
//
//	go-binenc-gen example.go
//
// in the same directory will create the file example_encoding.go, in package example,
// containing definitions of
//
//	func (s *Header) Write(w io.Writer) (n int, err error)
//	func (s *Header) Read(r io.Reader) (err error)
//	func (s *Request) Write(w io.Writer) (n int, err error)
//	func (s *Request) Read(r io.Reader) (err error)
//
// These methods will serialize Header and Request objects, using a single allocation
// per Write. For Reads, there will be as many allocations as pointers and slices in
// the struct:
//
//	func (s *Request) Write(w io.Writer) (n int, err error) {
//	        size := 10
//	        for _, v := range s.Headers {
//	                size += 4
//	                size += len(v.Name) + len(v.Value)
//	        }
//	        buf := make([]byte, size)
//	        offset := 0
//	        // Headers
//	        buf[offset] = byte(len(s.Headers))
//	        buf[offset+1] = byte(len(s.Headers) >> 8)
//	        offset += 2
//	        for _, v := range s.Headers {
//	                buf[offset] = byte(len(v.Name))
//	                buf[offset+1] = byte(len(v.Name) >> 8)
//	                offset += 2
//	                copy(buf[offset:], v.Name)
//	                offset += len(v.Name)
//	                buf[offset] = byte(len(v.Value))
//	                buf[offset+1] = byte(len(v.Value) >> 8)
//	                offset += 2
//	                copy(buf[offset:], v.Value)
//	                offset += len(v.Value)
//	        }
//	        // ResponseTime
//	        buf[offset] = byte(s.ResponseTime)
//	        buf[offset+1] = byte(s.ResponseTime >> 8)
//	        buf[offset+2] = byte(s.ResponseTime >> 16)
//	        buf[offset+3] = byte(s.ResponseTime >> 24)
//	        buf[offset+4] = byte(s.ResponseTime >> 32)
//	        buf[offset+5] = byte(s.ResponseTime >> 40)
//	        buf[offset+6] = byte(s.ResponseTime >> 48)
//	        buf[offset+7] = byte(s.ResponseTime >> 56)
//	        offset += 8
//	        return w.Write(buf)
//	}
//
//	func (s *Request) Read(r io.Reader) error {
//	        buf := make([]byte, 8)
//	        var size uint16
//	        // Headers
//	        r.Read(buf[:2])
//	        size = uint16(buf[0]) | (uint16(buf[1]) << 8)
//	        s.Headers = make([]Header, size)
//	        si := int(size)
//	        for i := 0; i < si; i++ {
//	                r.Read(buf[:2])
//	                size = uint16(buf[0]) | (uint16(buf[1]) << 8)
//	                strBuf_0 := make([]byte, size)
//	                r.Read(strBuf_0)
//	                s.Headers[i].Name = *(*string)(unsafe.Pointer(&strBuf_0))
//	                r.Read(buf[:2])
//	                size = uint16(buf[0]) | (uint16(buf[1]) << 8)
//	                strBuf_1 := make([]byte, size)
//	                r.Read(strBuf_1)
//	                s.Headers[i].Value = *(*string)(unsafe.Pointer(&strBuf_1))
//	        }
//	        // ResponseTime
//	        r.Read(buf[:8])
//	        s.ResponseTime = uint64(buf[0]) | (uint64(buf[1]) << 8) | (uint64(buf[2]) << 16) | (uint64(buf[3]) << 24) | (uint64(buf[4]) << 32) | (uint64(buf[5]) << 40) | (uint64(buf[6]) << 48) | (uint64(buf[7]) << 56)
//	        return nil
//	}
//
// Note that the serialization of Header objects during the serialization of a Request
// is inlined. For that reason, recursive definitions are not yet supported.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"go/types"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/cezarguimaraes/go-binenc-gen/encoder"
	"golang.org/x/tools/go/packages"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("binenc: ")

	flag.Parse()
	tags := []string{}
	args := flag.Args()
	if len(args) == 0 {
		args = []string{"."}
	}

	var dir string
	if len(args) == 1 && isDirectory(args[0]) {
		dir = args[0]
	} else {
		if len(tags) != 0 {
			log.Fatal("-tags option applies only to directories, not when files are specified")
		}
		dir = filepath.Dir(args[0])
	}

	g := &Generator{}
	g.parsePackage(args, tags)

	g.generate()

	fmt.Fprintf(&g.hdr, "// Code generated by \"gobinenc %s\"; DO NOT EDIT.\n", strings.Join(os.Args[1:], " "))
	fmt.Fprintf(&g.hdr, "\n")
	fmt.Fprintf(&g.hdr, "package %s", g.pkg.name)
	fmt.Fprintf(&g.hdr, "\n")

	fmt.Fprintf(&g.hdr, "import (\n")
	fmt.Fprintf(&g.hdr, "\t\"io\"\n")
	if g.needUnsafe {
		fmt.Fprintf(&g.hdr, "\t\"unsafe\"\n")
	}
	fmt.Fprintf(&g.hdr, ")")
	fmt.Fprintf(&g.hdr, "\n")

	src := g.format()
	baseName := fmt.Sprintf("%s_encoding.go", g.pkg.name)
	outputName := filepath.Join(dir, strings.ToLower(baseName))
	err := os.WriteFile(outputName, src, 0644)
	if err != nil {
		log.Fatalf("writing output: %s", err)
	}
}

// isDirectory reports whether the named file is a directory.
func isDirectory(name string) bool {
	info, err := os.Stat(name)
	if err != nil {
		log.Fatal(err)
	}
	return info.IsDir()
}

type Struct struct {
	Name string
	Type types.Type
}

type File struct {
	pkg  *Package
	file *ast.File
	// typeName string
	// values   []Value
	structs []*Struct
}

type Package struct {
	name     string
	typeInfo *types.Info
	files    []*File
}

type Generator struct {
	hdr   bytes.Buffer
	buf   bytes.Buffer
	pkg   *Package
	types *types.Package

	needUnsafe bool
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
	if len(pkgs) != 1 {
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
	g.types = pkg.Types

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
	src, err := format.Source(append(g.hdr.Bytes(), g.buf.Bytes()...))
	if err != nil {
		log.Printf("warning: internal error: invalid Go generated: %s", err)
		log.Printf("warning: compile the package to analyze the error")
		return g.buf.Bytes()
	}
	return src
}

func (g *Generator) generateWrite(s *Struct) {
	g.Printf("func (s *%s) Write(w io.Writer) (n int, err error) {\n", s.Name)
	e := encoder.NewWriter(g.types)
	e.Printf("\toffset := 0\n")
	e.WriteField("s", s.Type)
	e.Printf("\treturn w.Write(buf)\n")
	e.Printf("}\n\n")
	g.Printf(e.SizeExpr())
	g.Printf("\tbuf := make([]byte, size)\n")
	e.WriteTo(&g.buf)
}

func (g *Generator) generateRead(s *Struct) {
	g.Printf("func (s *%s) Read(r io.Reader) error {\n", s.Name)
	e := encoder.NewWriter(g.types)
	e.ReadField("s", s.Type)
	g.Printf(e.HeaderExpr())
	e.WriteTo(&g.buf)
	g.Printf("\treturn nil\n")
	g.Printf("}\n\n")
	if e.NeedUnsafe() {
		g.needUnsafe = true
	}
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
		t := f.pkg.typeInfo.TypeOf(st)
		f.structs = append(f.structs, &Struct{tspec.Name.Name, t})
	}
	return false
}

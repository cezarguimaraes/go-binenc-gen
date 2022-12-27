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
// Source adapted from https://cs.opensource.google/go/x/tools/+/refs/tags/v0.4.0:cmd/binenc/endtoend_test.go

package main

import (
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"testing"
)

// This file contains a test that compiles and runs each program in testdata
// after generating the encode and decode method for its type. The rule is that for testdata/x.go
// we run go-binenc-gen x.go and then compile and run the program. The resulting
// binary panics if encoding and then decoding a sample struct does not yield an identical one, including for error cases.

func TestEndToEnd(t *testing.T) {
	dir, binenc := buildBinenc(t)
	defer os.RemoveAll(dir)
	// Read the testdata directory.
	fd, err := os.Open("testdata")
	if err != nil {
		t.Fatal(err)
	}
	defer fd.Close()
	names, err := fd.Readdirnames(-1)
	if err != nil {
		t.Fatalf("Readdirnames: %s", err)
	}
	// Generate, compile, and run the test programs.
	for _, name := range names {
		if !strings.HasSuffix(name, ".go") {
			t.Errorf("%s is not a Go file", name)
			continue
		}
		if strings.HasSuffix(name, "_encoding.go") {
			t.Logf("ignoring generated %s file", name)
			continue
		}
		binencCompileAndRun(t, dir, binenc, name)
	}
}

// buildBinenc creates a temporary directory and installs binenc there.
func buildBinenc(t *testing.T) (dir string, binenc string) {
	t.Helper()
	// testenv.NeedsTool(t, "go")

	dir, err := os.MkdirTemp("", "binenc")
	if err != nil {
		t.Fatal(err)
	}
	binenc = filepath.Join(dir, "binenc.exe")
	err = run("go", "build", "-o", binenc)
	if err != nil {
		t.Fatalf("building binenc: %s", err)
	}
	return dir, binenc
}

// binencCompileAndRun runs binenc for the named file and compiles and
// runs the target binary in directory dir. That binary will panic if the String method is incorrect.
func binencCompileAndRun(t *testing.T, dir, binenc, fileName string) {
	t.Helper()
	t.Logf("run: %s\n", fileName)
	source := filepath.Join(dir, path.Base(fileName))
	err := copy(source, filepath.Join("testdata", fileName))
	if err != nil {
		t.Fatalf("copying file to temporary directory: %s", err)
	}
	encodeSource := filepath.Join(dir, "main_encoding.go")
	// Run binenc in temporary directory.
	err = run(binenc, source)
	if err != nil {
		t.Fatal(err)
	}
	// Run the binary in the temporary directory.
	err = run("go", "run", encodeSource, source)
	if err != nil {
		t.Fatal(err)
	}
}

// copy copies the from file to the to file.
func copy(to, from string) error {
	toFd, err := os.Create(to)
	if err != nil {
		return err
	}
	defer toFd.Close()
	fromFd, err := os.Open(from)
	if err != nil {
		return err
	}
	defer fromFd.Close()
	_, err = io.Copy(toFd, fromFd)
	return err
}

// run runs a single command and returns an error if it does not succeed.
// os/exec should have this function, to be honest.
func run(name string, arg ...string) error {
	return runInDir(".", name, arg...)
}

// runInDir runs a single command in directory dir and returns an error if
// it does not succeed.
func runInDir(dir, name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "GO111MODULE=auto")
	return cmd.Run()
}

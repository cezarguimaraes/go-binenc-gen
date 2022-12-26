package encoder_test

import (
	"go/types"
	"strings"
	"testing"

	"github.com/cezarguimaraes/go-binenc-gen/encoder"
	"github.com/google/go-cmp/cmp"
)

func parseOutput(t *testing.T, e *encoder.Writer) []string {
	t.Helper()
	raw := string(e.Bytes())
	lines := strings.Split(raw, "\n")
	for i := 0; i < len(lines); i++ {
		lines[i] = strings.TrimSpace(lines[i])
	}
	return lines
}

func TestWriteField_Pointers(t *testing.T) {
	cases := []struct {
		name string
		want []string
		t    types.Type
	}{
		{
			name: "**byte",
			want: []string{
				"buf[offset] = byte(**test)",
				"offset += 1",
				"",
			},
			t: types.NewPointer(types.NewPointer(types.Typ[types.Byte])),
		},
		{
			name: "*byte",
			want: []string{
				"buf[offset] = byte(*test)",
				"offset += 1",
				"",
			},
			t: types.NewPointer(types.Typ[types.Byte]),
		},
		{
			name: "*int8",
			want: []string{
				"buf[offset] = byte(uint8(*test))",
				"offset += 1",
				"",
			},
			t: types.NewPointer(types.Typ[types.Int8]),
		},
		{
			name: "*uint8",
			want: []string{
				"buf[offset] = byte(*test)",
				"offset += 1",
				"",
			},
			t: types.NewPointer(types.Typ[types.Uint8]),
		},
		{
			name: "*int16",
			want: []string{
				"buf[offset] = byte(uint16(*test))",
				"buf[offset + 1] = byte(uint16(*test) >> 8)",
				"offset += 2",
				"",
			},
			t: types.NewPointer(types.Typ[types.Int16]),
		},
		{
			name: "*uint16",
			want: []string{
				"buf[offset] = byte(*test)",
				"buf[offset + 1] = byte(*test >> 8)",
				"offset += 2",
				"",
			},
			t: types.NewPointer(types.Typ[types.Uint16]),
		},
		{
			name: "*uint32",
			want: []string{
				"buf[offset] = byte(*test)",
				"buf[offset + 1] = byte(*test >> 8)",
				"buf[offset + 2] = byte(*test >> 16)",
				"buf[offset + 3] = byte(*test >> 24)",
				"offset += 4",
				"",
			},
			t: types.NewPointer(types.Typ[types.Uint32]),
		},
		{
			name: "*uint64",
			want: []string{
				"buf[offset] = byte(*test)",
				"buf[offset + 1] = byte(*test >> 8)",
				"buf[offset + 2] = byte(*test >> 16)",
				"buf[offset + 3] = byte(*test >> 24)",
				"buf[offset + 4] = byte(*test >> 32)",
				"buf[offset + 5] = byte(*test >> 40)",
				"buf[offset + 6] = byte(*test >> 48)",
				"buf[offset + 7] = byte(*test >> 56)",
				"offset += 8",
				"",
			},
			t: types.NewPointer(types.Typ[types.Uint64]),
		},
		{
			name: "*Int32",
			want: []string{
				"buf[offset] = byte(uint32(*test))",
				"buf[offset + 1] = byte(uint32(*test) >> 8)",
				"buf[offset + 2] = byte(uint32(*test) >> 16)",
				"buf[offset + 3] = byte(uint32(*test) >> 24)",
				"offset += 4",
				"",
			},
			t: types.NewPointer(types.Typ[types.Int32]),
		},
		{
			name: "*Int64",
			want: []string{
				"buf[offset] = byte(uint64(*test))",
				"buf[offset + 1] = byte(uint64(*test) >> 8)",
				"buf[offset + 2] = byte(uint64(*test) >> 16)",
				"buf[offset + 3] = byte(uint64(*test) >> 24)",
				"buf[offset + 4] = byte(uint64(*test) >> 32)",
				"buf[offset + 5] = byte(uint64(*test) >> 40)",
				"buf[offset + 6] = byte(uint64(*test) >> 48)",
				"buf[offset + 7] = byte(uint64(*test) >> 56)",
				"offset += 8",
				"",
			},
			t: types.NewPointer(types.Typ[types.Int64]),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			e := encoder.NewWriter()
			e.WriteField("test", c.t)
			lines := parseOutput(t, e)
			if diff := cmp.Diff(c.want, lines); diff != "" {
				t.Errorf("e.WriteField(%q, %q): (-want, +got):\n%s", "test", c.t.String(), diff)
			}
		})
	}
}

func TestWriteField_IntegerTypes(t *testing.T) {
	cases := []struct {
		name string
		want []string
		t    types.Type
	}{
		{
			name: "byte",
			want: []string{
				"buf[offset] = byte(test)",
				"offset += 1",
				"",
			},
			t: types.Typ[types.Byte],
		},
		{
			name: "int8",
			want: []string{
				"buf[offset] = byte(uint8(test))",
				"offset += 1",
				"",
			},
			t: types.Typ[types.Int8],
		},
		{
			name: "uint8",
			want: []string{
				"buf[offset] = byte(test)",
				"offset += 1",
				"",
			},
			t: types.Typ[types.Uint8],
		},
		{
			name: "int16",
			want: []string{
				"buf[offset] = byte(uint16(test))",
				"buf[offset + 1] = byte(uint16(test) >> 8)",
				"offset += 2",
				"",
			},
			t: types.Typ[types.Int16],
		},
		{
			name: "uint16",
			want: []string{
				"buf[offset] = byte(test)",
				"buf[offset + 1] = byte(test >> 8)",
				"offset += 2",
				"",
			},
			t: types.Typ[types.Uint16],
		},
		{
			name: "uint32",
			want: []string{
				"buf[offset] = byte(test)",
				"buf[offset + 1] = byte(test >> 8)",
				"buf[offset + 2] = byte(test >> 16)",
				"buf[offset + 3] = byte(test >> 24)",
				"offset += 4",
				"",
			},
			t: types.Typ[types.Uint32],
		},
		{
			name: "uint64",
			want: []string{
				"buf[offset] = byte(test)",
				"buf[offset + 1] = byte(test >> 8)",
				"buf[offset + 2] = byte(test >> 16)",
				"buf[offset + 3] = byte(test >> 24)",
				"buf[offset + 4] = byte(test >> 32)",
				"buf[offset + 5] = byte(test >> 40)",
				"buf[offset + 6] = byte(test >> 48)",
				"buf[offset + 7] = byte(test >> 56)",
				"offset += 8",
				"",
			},
			t: types.Typ[types.Uint64],
		},
		{
			name: "Int32",
			want: []string{
				"buf[offset] = byte(uint32(test))",
				"buf[offset + 1] = byte(uint32(test) >> 8)",
				"buf[offset + 2] = byte(uint32(test) >> 16)",
				"buf[offset + 3] = byte(uint32(test) >> 24)",
				"offset += 4",
				"",
			},
			t: types.Typ[types.Int32],
		},
		{
			name: "Int64",
			want: []string{
				"buf[offset] = byte(uint64(test))",
				"buf[offset + 1] = byte(uint64(test) >> 8)",
				"buf[offset + 2] = byte(uint64(test) >> 16)",
				"buf[offset + 3] = byte(uint64(test) >> 24)",
				"buf[offset + 4] = byte(uint64(test) >> 32)",
				"buf[offset + 5] = byte(uint64(test) >> 40)",
				"buf[offset + 6] = byte(uint64(test) >> 48)",
				"buf[offset + 7] = byte(uint64(test) >> 56)",
				"offset += 8",
				"",
			},
			t: types.Typ[types.Int64],
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			e := encoder.NewWriter()
			e.WriteField("test", c.t)
			lines := parseOutput(t, e)
			if diff := cmp.Diff(c.want, lines); diff != "" {
				t.Errorf("e.WriteField(%q, %q): (-want, +got):\n%s", "test", c.t.String(), diff)
			}
		})
	}
}

func TestWriteField_Array(t *testing.T) {
	cases := []struct {
		name string
		want []string
		t    types.Type
	}{
		{
			name: "[]int16",
			want: []string{
				"buf[offset] = byte(len(test))",
				"buf[offset + 1] = byte(len(test) >> 8)",
				"offset += 2",
				"for _, v := range test {",
				"buf[offset] = byte(uint16(v))",
				"buf[offset + 1] = byte(uint16(v) >> 8)",
				"offset += 2",
				"}",
				"",
			},
			t: types.NewSlice(types.Typ[types.Int16]),
		},
		{
			name: "[][]int16",
			want: []string{
				"buf[offset] = byte(len(test))",
				"buf[offset + 1] = byte(len(test) >> 8)",
				"offset += 2",
				"for _, v := range test {",
				"buf[offset] = byte(len(v))",
				"buf[offset + 1] = byte(len(v) >> 8)",
				"offset += 2",
				"for _, v1 := range v {",
				"buf[offset] = byte(uint16(v1))",
				"buf[offset + 1] = byte(uint16(v1) >> 8)",
				"offset += 2",
				"}",
				"}",
				"",
			},
			t: types.NewSlice(types.NewSlice(types.Typ[types.Int16])),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			e := encoder.NewWriter()
			e.WriteField("test", c.t)
			lines := parseOutput(t, e)
			if diff := cmp.Diff(c.want, lines); diff != "" {
				t.Errorf("e.WriteField(%q, %q): (-want, +got):\n%s", "test", c.t.String(), diff)
			}
		})
	}
}

func TestWriteField_Boolean(t *testing.T) {
	e := encoder.NewWriter()
	e.WriteField("test", types.Typ[types.Bool])
	got := parseOutput(t, e)
	want := []string{
		"if test {",
		"buf[offset] = byte(0x01)",
		"} else {",
		"buf[offset] = byte(0x00)",
		"}",
		"",
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("e.WriteField(%q, %q): (-want, +got):\n%s", "test", types.Typ[types.Bool].String(), diff)
	}
}

func TestWriteField_String(t *testing.T) {
	e := encoder.NewWriter()
	e.WriteField("test", types.Typ[types.String])
	got := parseOutput(t, e)
	want := []string{
		"buf[offset] = byte(len(test))",
		"buf[offset + 1] = byte(len(test) >> 8)",
		"offset += 2",
		"copy(buf[offset:], test)",
		"offset += len(test)",
		"",
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("e.WriteField(%q, %q): (-want, +got):\n%s", "test", types.Typ[types.String].String(), diff)
	}
}

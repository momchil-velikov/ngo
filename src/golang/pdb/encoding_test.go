package pdb

import (
	"bytes"
	"io"
	"testing"
)

func eq(a []byte, b []byte) bool {
	if n := len(a); n == len(b) {
		for i := 0; i < n; i++ {
			if a[i] != b[i] {
				return false
			}
		}
		return true
	}
	return false
}

func TestULEB128(t *testing.T) {
	n := []uint64{
		0, 1, 5, 0x7e, 0x7f, 0x80, 0x81, 0xef, 0xfe, 0xff, 0x7ffe, 0x7fff,
		0x8000, 0x10000, 0x98765,
	}
	b := [][]byte{
		{0},
		{1},
		{5},
		{0x7e},
		{0x7f},
		{0x80, 0x01},
		{0x81, 0x01},
		{0xef, 0x01},
		{0xfe, 0x01},
		{0xff, 0x01},
		{0xfe, 0xff, 0x01},
		{0xff, 0xff, 0x01},
		{0x80, 0x80, 0x02},
		{0x80, 0x80, 0x04},
		{0xe5, 0x8e, 0x26},
	}

	for i := range n {
		buf := bytes.Buffer{}
		enc := new(Encoder).Init(&buf)
		if err := enc.WriteNum(n[i]); err != nil {
			t.Fatal(err)
		}
		if !eq(buf.Bytes(), b[i]) {
			t.Error("encoding", n[i], "expected", b[i], "actual", buf.Bytes())
		}
	}

	for i := range n {
		buf := bytes.NewBuffer(b[i])
		dec := new(Decoder).Init(buf)
		v, err := dec.ReadNum()
		if err != nil {
			t.Fatal(err)
		}
		if v != n[i] {
			t.Error("decoding", b[i], "expected", n[i], "actual", v)
		}
	}
}

func TestString(t *testing.T) {
	ss := []string{"", "a", "а", "αβγ"}
	es := [][]byte{
		{0},
		{1, 'a'},
		{2, 208, 176},
		{6, 206, 177, 206, 178, 206, 179},
	}

	for i, s := range ss {
		buf := bytes.Buffer{}
		enc := new(Encoder).Init(&buf)
		if err := enc.WriteString(s); err != nil {
			t.Fatal(err)
		}
		if !eq(buf.Bytes(), es[i]) {
			t.Error("encoding", s, "expected", es[i], "actual", buf.Bytes())
		}
	}

	for i, e := range es {
		buf := bytes.NewBuffer(e)
		dec := new(Decoder).Init(buf)
		s, err := dec.ReadString()
		if err != nil {
			t.Fatal(err)
		}
		if s != ss[i] {
			t.Error("decoding", e, "expected", s[i], "actual", s)
		}
	}
}

func TestBytes(t *testing.T) {
	buf := bytes.Buffer{}
	enc := new(Encoder).Init(&buf)
	if err := enc.WriteBytes(nil); err != nil {
		t.Fatal(err)
	}
	if err := enc.WriteBytes([]byte{}); err != nil {
		t.Fatal(err)
	}
	bs := [][]byte{
		{1},
		{1, 2, 3},
		{1, 2, 3, 4, 5},
	}
	for _, b := range bs {
		if err := enc.WriteByte(b[0]); err != nil {
			t.Fatal(err)
		}
		if err := enc.WriteBytes(b[1:]); err != nil {
			t.Fatal(err)
		}
	}
	b := make([]byte, 9)
	dec := new(Decoder).Init(&buf)
	if err := dec.ReadBytes(b[0:2]); err != nil {
		t.Fatal(err)
	}
	if err := dec.ReadBytes(b[2:5]); err != nil {
		t.Fatal(err)
	}
	if !eq(b[:5], []byte{1, 1, 2, 3, 1}) {
		t.Error("invalid read of ", b[:5])
	}
	if b, err := dec.ReadByte(); err != nil {
		t.Fatal(err)
	} else if b != 2 {
		t.Error("invalid byte read", b, "expected", 2)
	}
	if err := dec.ReadBytes(b); err == nil {
		t.Error("expected EOF error")
	} else if err != io.EOF {
		t.Fatal(err)
	}
}

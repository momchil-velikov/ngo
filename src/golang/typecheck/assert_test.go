package typecheck

import (
	"fmt"
	"golang/ast"
	"testing"
)

func TestAssert(t *testing.T) {
	p, err := compilePackage("_test/src/assert", []string{"assert.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	A := p.Find("A").(*ast.TypeDecl)
	B := p.Find("B").(*ast.TypeDecl)
	S := p.Find("S").(*ast.TypeDecl)
	T := p.Find("T").(*ast.TypeDecl)

	for _, cs := range []struct {
		name string
		typ  ast.Type
	}{
		{"u", B},
		{"v", A},
		{"w", S},
		{"x", S},
		{"q", ast.BuiltinInt},
	} {
		v := p.Find(cs.name).(*ast.Var)
		if v.Type != cs.typ {
			t.Errorf("`%s` must have type `%s`\n", cs.name, cs.typ)
		}
	}

	v := p.Find("y").(*ast.Var)
	if ptr, ok := v.Type.(*ast.PtrType); !ok || ptr.Base != T {
		t.Error("`y` must have type `*T`\n")
	}

	for i := 0; i <= 4; i++ {
		name := fmt.Sprintf("o%d", i)
		v := p.Find(name).(*ast.Var)
		if v.Type != ast.BuiltinBool {
			t.Errorf("`%s` must have type `untyped bool`\n", name)
		}
	}
}

func TestAssertErr(t *testing.T) {
	expectError(t, "_test/src/assert", []string{"assert-err-01.go"},
		"invalid type asserion: `[]int` is not an interface type")
	expectError(t, "_test/src/assert", []string{"assert-err-02.go"},
		"type `[]int` does not implement interface `I`")
	expectError(t, "_test/src/assert", []string{"assert-err-03.go"},
		"type `S` does not implement interface `I`")
	expectError(t, "_test/src/assert", []string{"assert-err-04.go"},
		"invalid use of `.(type)` outside type switch")
}

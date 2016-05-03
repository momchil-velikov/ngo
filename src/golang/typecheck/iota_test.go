package typecheck

import (
	"golang/ast"
	"testing"
)

func TestIota(t *testing.T) {
	p, err := compilePackage("_test/src/iota", []string{"iota.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	for _, cs := range []struct {
		name  string
		typ   ast.Type
		value string
	}{
		{"A", ast.BuiltinUntypedInt, "0"},
		{"B", ast.BuiltinUntypedInt, "1"},
		{"C", ast.BuiltinInt, "2"},
		{"D", ast.BuiltinInt, "0"},
		{"E", ast.BuiltinInt, "1"},
		{"F", ast.BuiltinInt, "2"},
		{"G", ast.BuiltinUint16, "65519"},
		{"U", ast.BuiltinUint16, "65527"},
		{"V", ast.BuiltinUint16, "65519"},
		{"N", ast.BuiltinUntypedInt, "0"},
		{"X", ast.BuiltinUntypedInt, "-1"},
		{"Y", ast.BuiltinUntypedInt, "1"},
		{"Z", ast.BuiltinUntypedInt, "2"},
		{"AA", ast.BuiltinUntypedInt, "0"},
		{"BB", ast.BuiltinUntypedInt, "2"},
		{"CC", ast.BuiltinUntypedInt, "3"},
		{"DD", ast.BuiltinUntypedInt, "1"},
	} {
		C := p.Find(cs.name).(*ast.Const)
		if C.Type != cs.typ {
			t.Errorf("`%s` must have type `%s`\n", cs.name, cs.typ)
		}
		c := C.Init.(*ast.ConstValue)
		if s := c.String(); s != cs.value {
			t.Errorf("`%s` must have value `%s`\n", cs.name, s)
		}
	}
}

func TestIotaErr(t *testing.T) {
	expectError(t, "_test/src/iota", []string{"iota-err-01.go"},
		"`iota` used outside const declaration")
}

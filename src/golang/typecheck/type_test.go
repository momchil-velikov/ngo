package typecheck

import (
	"golang/ast"
	"golang/parser"
	"strings"
	"testing"
)

func compilePackage(
	dir string, srcs []string, loc ast.PackageLocator) (*ast.Package, error) {

	p, err := parser.ParsePackage(dir, srcs)
	if err != nil {
		return nil, err
	}
	err = parser.ResolvePackage(p, loc)
	if err != nil {
		return nil, err
	}
	return p, CheckPackage(p)
}

func expectError(t *testing.T, pkg string, srcs []string, msg string) {
	_, err := compilePackage(pkg, srcs, nil)
	if err == nil || !strings.Contains(err.Error(), msg) {
		t.Errorf("%s:%v: expected `%s` error", pkg, srcs, msg)
		if err == nil {
			t.Log("actual: no error")
		} else {
			t.Logf("actual: %s", err.Error())
		}
	}
}

func TestBasicType(t *testing.T) {
	_, err := compilePackage("_test/src/typ", []string{"basic.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestBasicLoopErr(t *testing.T) {
	expectError(t, "_test/src/typ", []string{"basic-loop-1.go"}, "invalid recursive")
	expectError(t, "_test/src/typ", []string{"basic-loop-2.go"}, "invalid recursive")
	expectError(t, "_test/src/typ", []string{"basic-loop-3.go"}, "invalid recursive")
}

func TestArrayType(t *testing.T) {
	_, err := compilePackage("_test/src/typ", []string{"array.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestArrayErr(t *testing.T) {
	expectError(t, "_test/src/typ", []string{"array-loop-1.go"}, "invalid recursive")
	expectError(t, "_test/src/typ", []string{"array-loop-2.go"}, "invalid recursive")
	expectError(t, "_test/src/typ", []string{"array-loop-3.go"}, "invalid recursive")
	expectError(t, "_test/src/typ", []string{"array-loop-4.go"}, "invalid recursive")
}

func TestSliceType(t *testing.T) {
	_, err := compilePackage("_test/src/typ", []string{"slice.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPtrType(t *testing.T) {
	_, err := compilePackage("_test/src/typ", []string{"ptr.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestChanType(t *testing.T) {
	_, err := compilePackage("_test/src/typ", []string{"chan.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestStructType(t *testing.T) {
	_, err := compilePackage("_test/src/typ", []string{"struct.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestStructErr(t *testing.T) {
	expectError(t, "_test/src/typ", []string{"struct-loop-1.go"}, "invalid recursive")
	expectError(t, "_test/src/typ", []string{"struct-loop-2.go"}, "invalid recursive")
	expectError(t, "_test/src/typ", []string{"struct-loop-3.go"}, "invalid recursive")
	expectError(t, "_test/src/typ", []string{"struct-dup-1.go"},
		"non-unique field name `X`")
	expectError(t, "_test/src/typ", []string{"struct-dup-2.go"},
		"non-unique field name `X`")
	expectError(t, "_test/src/typ", []string{"struct-dup-2.go"},
		"non-unique field name `X`")
	expectError(t, "_test/src/typ", []string{"struct-anon-1.go"},
		"embedded type cannot be a pointer type")
	expectError(t, "_test/src/typ", []string{"struct-anon-2.go"},
		"embedded type cannot be a pointer type")
	expectError(t, "_test/src/typ", []string{"struct-anon-3.go"},
		"embedded type cannot be a pointer to interface type")
}

func TestMapType(t *testing.T) {
	_, err := compilePackage("_test/src/typ", []string{"map.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestMapErr(t *testing.T) {
	expectError(t, "_test/src/typ", []string{"map-loop-1.go"}, "invalid recursive")
	expectError(t, "_test/src/typ", []string{"map-loop-2.go"}, "invalid recursive")
	expectError(t, "_test/src/typ", []string{"map-loop-3.go"}, "invalid recursive")
	expectError(t, "_test/src/typ", []string{"map-eq-1.go"}, "invalid type for map key")
	expectError(t, "_test/src/typ", []string{"map-eq-2.go"}, "invalid type for map key")
	expectError(t, "_test/src/typ", []string{"map-eq-3.go"}, "invalid type for map key")
	expectError(t, "_test/src/typ", []string{"map-eq-4.go"}, "invalid type for map key")
	expectError(t, "_test/src/typ", []string{"map-eq-5.go"}, "invalid type for map key")
	expectError(t, "_test/src/typ", []string{"map-eq-6.go"}, "invalid type for map key")
}

func TestFuncType(t *testing.T) {
	_, err := compilePackage("_test/src/typ", []string{"func.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestIfaceType(t *testing.T) {
	_, err := compilePackage("_test/src/typ", []string{"iface.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestIfaceErr(t *testing.T) {
	expectError(t, "_test/src/typ", []string{"iface-embad-1.go"},
		"embeds non-interface type")
	expectError(t, "_test/src/typ", []string{"iface-embad-2.go"},
		"expected <id>")
	expectError(t, "_test/src/typ", []string{"iface-embad-3.go"},
		"invalid recursive type")
	expectError(t, "_test/src/typ", []string{"iface-embad-4.go"},
		"invalid recursive type")
	expectError(t, "_test/src/typ", []string{"iface-err-1.go"},
		"invalid type for map key")
	expectError(t, "_test/src/typ", []string{"method-dup-1.go"},
		"in declaration of J: duplicate method name F")
	expectError(t, "_test/src/typ", []string{"method-dup-2.go"},
		"in declaration of K: duplicate method name F")
	expectError(t, "_test/src/typ", []string{"method-dup-3.go"},
		"in declaration of K: duplicate method name F")
	expectError(t, "_test/src/typ", []string{"method-dup-4.go"},
		"in declaration of K: duplicate method name F")
	expectError(t, "_test/src/typ", []string{"method-dup-5.go"},
		": duplicate method name F")
	expectError(t, "_test/src/typ", []string{"method-dup-6.go"},
		": duplicate method name F")
	expectError(t, "_test/src/typ", []string{"method-dup-7.go"},
		"F conflicts with field name")
	expectError(t, "_test/src/typ", []string{"method-dup-8.go"},
		"F conflicts with field name")
	expectError(t, "_test/src/typ", []string{"method-dup-9.go"},
		"F conflicts with field name")
	expectError(t, "_test/src/typ", []string{"method-dup-10.go"},
		"in declaration of J: duplicate method name F")
}

func TestConstType(t *testing.T) {
	_, err := compilePackage("_test/src/typ", []string{"const.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestConstTypeErr(t *testing.T) {
	expectError(t, "_test/src/typ", []string{"const-err.go"}, "invalid constant type")
}

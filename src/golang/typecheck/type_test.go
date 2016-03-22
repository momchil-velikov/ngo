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

func expectErrorWithLoc(
	t *testing.T, pkg string, srcs []string, loc ast.PackageLocator, msg string) {
	_, err := compilePackage(pkg, srcs, loc)
	if err == nil || !strings.Contains(err.Error(), msg) {
		t.Errorf("%s:%v: expected `%s` error", pkg, srcs, msg)
		if err == nil {
			t.Log("actual: no error")
		} else {
			t.Logf("actual: %s", err.Error())
		}
	}
}

func expectError(t *testing.T, pkg string, srcs []string, msg string) {
	expectErrorWithLoc(t, pkg, srcs, nil, msg)
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
	p, err := compilePackage("_test/src/typ", []string{"array.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Check the array length is a constant, converted to `int`.
	B := p.Find("B").(*ast.TypeDecl)
	typ := B.Type.(*ast.ArrayType)
	c, ok := typ.Dim.(*ast.ConstValue)
	if !ok || c.Typ != ast.BuiltinInt {
		t.Error("array length must be `int` constant")
	}
}

func TestArrayErr(t *testing.T) {
	expectError(t, "_test/src/typ", []string{"array-loop-1.go"}, "invalid recursive")
	expectError(t, "_test/src/typ", []string{"array-loop-2.go"}, "invalid recursive")
	expectError(t, "_test/src/typ", []string{"array-loop-3.go"}, "invalid recursive")
	expectError(t, "_test/src/typ", []string{"array-loop-4.go"}, "invalid recursive")
	expectError(t, "_test/src/typ", []string{"array-err-1.go"},
		"array length is not a constant")
	expectError(t, "_test/src/typ", []string{"array-err-2.go"},
		"cannot be converted to int")
	expectError(t, "_test/src/typ", []string{"array-err-3.go"},
		"array length must be non-negative")
	expectError(t, "_test/src/typ", []string{"array-err-4.go"},
		"unspecified array length not allowed")
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

func TestTypeIdentity(t *testing.T) {
	pa, err := compilePackage("_test/src/typ/identity/a", []string{"a.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Builtin types
	A0 := pa.Find("A0").(*ast.TypeDecl)
	if !identicalTypes(A0, A0) {
		t.Error("`A0` must be identical to itself")
	}
	if identicalTypes(A0.Type, A0) {
		t.Error("`A0` and `int32` must be distinct types")
	}
	if identicalTypes(A0, A0.Type) {
		t.Error("`A0` and `int32` must be distinct types")
	}
	A1 := pa.Find("A1").(*ast.TypeDecl)
	if identicalTypes(A0, A1) {
		t.Error("`A0` and `A1` must be distinct types")
	}
	if !identicalTypes(A0.Type, A1.Type) {
		t.Error("`A0` and `A1` must have the same underlying type")
	}
	A2 := pa.Find("A2").(*ast.TypeDecl)
	if identicalTypes(A0.Type, A2.Type) {
		t.Error("`A0` and `A1` must have distinct underlying types")
	}
	if !identicalTypes(A0.Type, A1.Type) {
		t.Error("`A0` and `A1` must have the same underlying type")
	}

	// Array types
	B0 := pa.Find("B0").(*ast.TypeDecl)
	B1 := pa.Find("B1").(*ast.TypeDecl)
	if identicalTypes(B0, B1) {
		t.Error("`B0` and `B1` must be distinct types")
	}
	if identicalTypes(A0.Type, B0.Type) {
		t.Error("`A0` and `B0` must have distinct underlying types")
	}
	if identicalTypes(B1.Type, A0.Type) {
		t.Error("`B1` and `A0` must have distinct underlying types")
	}
	if !identicalTypes(B0.Type, B1.Type) {
		t.Error("`B0` and `B1` must have the same underlying type")
	}
	B2 := pa.Find("B2").(*ast.TypeDecl)
	if identicalTypes(B2.Type, B0.Type) {
		t.Error("`B2` and `B0` must have distinct underlying types")
	}
	B3 := pa.Find("B3").(*ast.TypeDecl)
	if identicalTypes(B3.Type, B0.Type) {
		t.Error("`B3` and `B0` must have distinct underlying types")
	}

	// FIXME: reaload package after checking array types, until we handle
	// constants in the PDB.
	loc := &MockPackageLocator{pkgs: make(map[string]*ast.Package)}
	pa, err = reloadPackage(pa, loc)
	if err != nil {
		t.Fatal(err)
	}
	loc.pkgs["identity/a"] = pa
	B0 = pa.Find("B0").(*ast.TypeDecl)

	// Slice types
	C0 := pa.Find("C0").(*ast.TypeDecl)
	if identicalTypes(C0.Type, B0.Type) {
		t.Error("`C0` and `B0` must have distinct underlying types")
	}
	C1 := pa.Find("C1").(*ast.TypeDecl)
	if identicalTypes(C0, C1) {
		t.Error("`C0` and `C1` must be distinct types")
	}
	if !identicalTypes(C0.Type, C1.Type) {
		t.Error("`C0` and `C1` must have the same underlying type")
	}
	C2 := pa.Find("C2").(*ast.TypeDecl)
	if identicalTypes(C0.Type, C2.Type) {
		t.Error("`C0` and `C2` must have distinct underlying types")
	}

	// Pointer types
	D0 := pa.Find("D0").(*ast.TypeDecl)
	if identicalTypes(D0.Type, B0.Type) {
		t.Error("`D0` and `B0` must have distinct underlying types")
	}
	D1 := pa.Find("D1").(*ast.TypeDecl)
	if identicalTypes(D0, D1) {
		t.Error("`D0` and `D1` must be distinct types")
	}
	if !identicalTypes(D0.Type, D1.Type) {
		t.Error("`D0` and `D1` must have the same underlying type")
	}
	D2 := pa.Find("D2").(*ast.TypeDecl)
	if identicalTypes(D0.Type, D2.Type) {
		t.Error("`D0` and `D2` must have distinct underlying types")
	}

	// Map types
	E0 := pa.Find("E0").(*ast.TypeDecl)
	if identicalTypes(E0.Type, B0.Type) {
		t.Error("`E0` and `B0` must have distinct underlying types")
	}
	E1 := pa.Find("E1").(*ast.TypeDecl)
	if identicalTypes(E0, E1) {
		t.Error("`E0` and `E1` must be distinct types")
	}
	if !identicalTypes(E0.Type, E1.Type) {
		t.Error("`E0` and `E1` must have the same underlying type")
	}
	E2 := pa.Find("E2").(*ast.TypeDecl)
	if identicalTypes(E0.Type, E2.Type) {
		t.Error("`E0` and `E2` must have distinct underlying types")
	}
	E3 := pa.Find("E3").(*ast.TypeDecl)
	if identicalTypes(E0.Type, E3.Type) {
		t.Error("`E0` and `E3` must have distinct underlying types")
	}

	// Channel types
	F0 := pa.Find("F0").(*ast.TypeDecl)
	if identicalTypes(F0.Type, B0.Type) {
		t.Error("`F0` and `B0` must have distinct underlying types")
	}
	F1 := pa.Find("F1").(*ast.TypeDecl)
	if identicalTypes(F0, F1) {
		t.Error("`F0` and `F1` must be distinct types")
	}
	if !identicalTypes(F0.Type, F1.Type) {
		t.Error("`F0` and `F1` must have the same underlying type")
	}
	F2 := pa.Find("F2").(*ast.TypeDecl)
	if identicalTypes(F0.Type, F2.Type) {
		t.Error("`F0` and `F2` must have distinct underlying types")
	}
	F3 := pa.Find("F3").(*ast.TypeDecl)
	if identicalTypes(F0.Type, F3.Type) {
		t.Error("`F0` and `F3` must have distinct underlying types")
	}
	F4 := pa.Find("F4").(*ast.TypeDecl)
	if identicalTypes(F0.Type, F4.Type) {
		t.Error("`F0` and `F4` must have distinct underlying types")
	}
	if identicalTypes(F3.Type, F4.Type) {
		t.Error("`F3` and `F4` must have distinct underlying types")
	}

	// Struct  types
	G0 := pa.Find("G0").(*ast.TypeDecl)
	//	t.Logf("D0: %#v\n", G0.Type)
	//	t.Logf("B0: %#v\n", B0.Type)
	if identicalTypes(G0.Type, B0.Type) {
		t.Error("`G0` and `B0` must have distinct underlying types")
	}
	G1 := pa.Find("G1").(*ast.TypeDecl)
	if identicalTypes(G0, G1) {
		t.Error("`G0` and `G1` must be distinct types")
	}
	if !identicalTypes(G0.Type, G1.Type) {
		t.Error("`G0` and `G1` must have the same underlying type")
	}
	G2 := pa.Find("G2").(*ast.TypeDecl)
	if identicalTypes(G0.Type, G2.Type) {
		t.Error("`G0` and `G2` must have distinct underlying types")
	}
	G3 := pa.Find("G3").(*ast.TypeDecl)
	if identicalTypes(G0.Type, G3.Type) {
		t.Error("`G0` and `G3` must have have distinct underlying types")
	}
	G4 := pa.Find("G4").(*ast.TypeDecl)
	if identicalTypes(G3.Type, G4.Type) {
		t.Error("`G3` and `G4` must have have distinct underlying types")
	}
	G5 := pa.Find("G5").(*ast.TypeDecl)
	G6 := pa.Find("G6").(*ast.TypeDecl)
	if identicalTypes(G5, G6) {
		t.Error("`G5` and `G6` must be distinct types")
	}
	if !identicalTypes(G5.Type, G6.Type) {
		t.Error("`G5` and `G6` must have the same underlying type")
	}
	G7 := pa.Find("G7").(*ast.TypeDecl)
	if identicalTypes(G7.Type, G6.Type) {
		t.Error("`G7` and `G6` must have distinct underlying types")
	}

	pb, err := compilePackage("_test/src/typ/identity/b", []string{"b.go"}, loc)
	if err != nil {
		t.Fatal(err)
	}

	bG3 := pb.Find("G3").(*ast.TypeDecl)
	if identicalTypes(G3, bG3) {
		t.Error("`a.G3` and `b.G3` must be distinct types")
	}
	if !identicalTypes(G3.Type, bG3.Type) {
		t.Error("`a.G3` and `b.G3` must have the same underlying type")
	}

	bG5 := pb.Find("G5").(*ast.TypeDecl)
	if identicalTypes(G5, bG5) {
		t.Error("`a.G5` and `b.G5` must be distinct types")
	}
	if identicalTypes(G5.Type, bG5.Type) {
		t.Error("`a.G5` and `b.G5` must have distinct underlying types")
	}

	// Function types
	H0 := pa.Find("H0").(*ast.TypeDecl)
	//	t.Logf("D0: %#v\n", G0.Type)
	//	t.Logf("B0: %#v\n", B0.Type)
	if identicalTypes(H0.Type, B0.Type) {
		t.Error("`H0` and `B0` must have distinct underlying types")
	}
	H1 := pa.Find("H1").(*ast.TypeDecl)
	if identicalTypes(H0, H1) {
		t.Error("`H0` and `H1` must be distinct types")
	}
	if !identicalTypes(H0.Type, H1.Type) {
		t.Error("`H0` and `H1` must have the same underlying type")
	}
	H2 := pa.Find("H2").(*ast.TypeDecl)
	H3 := pa.Find("H3").(*ast.TypeDecl)
	if identicalTypes(H2, H3) {
		t.Error("`H2` and `H3` must be distinct types")
	}
	if !identicalTypes(H2.Type, H3.Type) {
		t.Error("`H2` and `H3` must have the same underlying type")
	}
	H4 := pa.Find("H4").(*ast.TypeDecl)
	if identicalTypes(H2.Type, H4.Type) {
		t.Error("`H2` and `H4` must have distinct underlying types")
	}
	H5 := pa.Find("H5").(*ast.TypeDecl)
	if identicalTypes(H2.Type, H5.Type) {
		t.Error("`H2` and `H4` must have distinct underlying types")
	}
	H6 := pa.Find("H6").(*ast.TypeDecl)
	if identicalTypes(H2.Type, H6.Type) {
		t.Error("`H2` and `H6` must have distinct underlying types")
	}
	H7 := pa.Find("H7").(*ast.TypeDecl)
	if identicalTypes(H2.Type, H7.Type) {
		t.Error("`H2` and `H7` must have distinct underlying types")
	}
	H8 := pa.Find("H8").(*ast.TypeDecl)
	if identicalTypes(H2.Type, H8.Type) {
		t.Error("`H2` and `H8` must have distinct underlying types")
	}
	H9 := pa.Find("H9").(*ast.TypeDecl)
	if !identicalTypes(H8.Type, H9.Type) {
		t.Error("`H8` and `H9` must have the same underlying type")
	}

	// Interface types.
	I0 := pa.Find("I0").(*ast.TypeDecl)
	if identicalTypes(I0.Type, B0.Type) {
		t.Error("`I0` and `B0` must have distinct underlying types")
	}
	I1 := pa.Find("I1").(*ast.TypeDecl)
	if identicalTypes(I0, I1) {
		t.Error("`I0` and `I1` must be distinct types")
	}
	if !identicalTypes(I0.Type, I1.Type) {
		t.Error("`I0` and `I1` must have the same underlying type")
	}
	I2 := pa.Find("I2").(*ast.TypeDecl)
	I3 := pa.Find("I3").(*ast.TypeDecl)
	if identicalTypes(I2, I3) {
		t.Error("`I2` and `I3` must be distinct types")
	}
	if !identicalTypes(I2.Type, I3.Type) {
		t.Error("`I2` and `I3` must have the same underlying type")
	}
	I4 := pa.Find("I4").(*ast.TypeDecl)
	I5 := pa.Find("I5").(*ast.TypeDecl)
	if identicalTypes(I4, I5) {
		t.Error("`I4` and `I5` must be distinct types")
	}
	if !identicalTypes(I4.Type, I5.Type) {
		t.Error("`I4` and `I5` must have the same underlying type")
	}
	I6 := pa.Find("I6").(*ast.TypeDecl)
	I7 := pa.Find("I7").(*ast.TypeDecl)
	if identicalTypes(I6, I7) {
		t.Error("`I6` and `I7` must be distinct types")
	}
	if !identicalTypes(I6.Type, I7.Type) {
		t.Error("`I6` and `I7` must have the same underlying type")
	}
	I8 := pa.Find("I8").(*ast.TypeDecl)
	if identicalTypes(I8.Type, I7.Type) {
		t.Error("`I8` and `I7` must have distinct underlying types")
	}
	I9 := pa.Find("I9").(*ast.TypeDecl)
	if identicalTypes(I8.Type, I9.Type) {
		t.Error("`I8` and `I9` must have distinct underlying types")
	}

	bI5 := pb.Find("I5").(*ast.TypeDecl)
	if identicalTypes(I5, bI5) {
		t.Error("`a.I5` and `b.I5` must be distinct types")
	}
	if !identicalTypes(I5.Type, bI5.Type) {
		t.Error("`a.I5` and `b.I5` must have the same underlying type")
	}
	bI7 := pb.Find("I7").(*ast.TypeDecl)
	if identicalTypes(I7, bI7) {
		t.Error("`a.I7` and `b.I7` must be distinct types")
	}
	if identicalTypes(I7.Type, bI7.Type) {
		t.Error("`a.I7` and `b.I7` must have distinct underlying types")
	}
}

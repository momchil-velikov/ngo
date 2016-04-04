package typecheck

import (
	"golang/ast"
	"testing"
)

func TestUnaryPlus(t *testing.T) {
	p, err := compilePackage("_test/src/unary", []string{"unary-plus.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	B := p.Find("B").(*ast.Const)
	c := B.Init.(*ast.ConstValue)
	if !isUntyped(c.Typ) {
		t.Error("`B` should be untyped")
	}
	if _, ok := c.Value.(ast.UntypedInt); !ok {
		t.Error("`B` should be untyped integer")
	}
	if v, ok := toInt(c); !ok || v != 1 {
		t.Error("`B` should have value 1")
	}

	D := p.Find("D").(*ast.Const)
	c = D.Init.(*ast.ConstValue)
	if !isUntyped(c.Typ) {
		t.Error("`D` should be untyped")
	}
	if _, ok := c.Value.(ast.Rune); !ok {
		t.Error("`D` should be untyped rune")
	}
	if v, ok := toInt(c); !ok || v != 'a' {
		t.Error("`D` should have value 'a'")
	}

	F := p.Find("F").(*ast.Const)
	c = F.Init.(*ast.ConstValue)
	if !isUntyped(c.Typ) {
		t.Error("`F` should be untyped")
	}
	if _, ok := c.Value.(ast.UntypedFloat); !ok {
		t.Error("`F` should be untyped float")
	}
	if v, ok := toFloat(c); !ok || v != 1.1 {
		t.Error("`F` should have value 1.1")
	}

	cB := p.Find("cB").(*ast.Const)
	c = cB.Init.(*ast.ConstValue)
	if c.Typ != ast.BuiltinInt {
		t.Error("`cB` should have type `int`")
	}
	if v, ok := toInt(c); !ok || v != 1 {
		t.Error("`B` should have value 1")
	}

	cD := p.Find("cD").(*ast.Const)
	c = cD.Init.(*ast.ConstValue)
	if c.Typ != ast.BuiltinInt32 {
		t.Error("`cD` should have type `rune`")
	}
	if v, ok := toInt(c); !ok || v != 'a' {
		t.Error("`cD` should have value 'a'")
	}

	cF := p.Find("cF").(*ast.Const)
	c = cF.Init.(*ast.ConstValue)
	if c.Typ != ast.BuiltinFloat32 {
		t.Error("`cF` should have type `float32`")
	}
	if v, ok := toFloat(c); !ok || v != 1.125 {
		t.Error("`cF` should have value 1.125")
	}

	// FIXME: check complex constants

	vB := p.Find("vB").(*ast.Var)
	if vB.Type != ast.BuiltinInt {
		t.Error("`vB` should have type `int`")
	}

	vD := p.Find("vD").(*ast.Var)
	if vD.Type != ast.BuiltinInt32 {
		t.Error("`vD` should have type `rune`")
	}

	vF := p.Find("vF").(*ast.Var)
	if vF.Type != ast.BuiltinFloat64 {
		t.Error("`vF` should have type `float64`")
	}

	z0 := p.Find("z0").(*ast.Const)
	if z0.Type != ast.BuiltinInt {
		t.Error("`z0` should have type `int`")
	}
	c = z0.Init.(*ast.ConstValue)
	if v, ok := c.Value.(ast.Int); !ok || v != 1 {
		t.Error("`z0` should have value 1")
	}
	z1 := p.Find("z1").(*ast.Const)
	if z1.Type != ast.BuiltinInt {
		t.Error("`z1` should have type `int`")
	}
	c = z1.Init.(*ast.ConstValue)
	if v, ok := c.Value.(ast.Int); !ok || v != 1 {
		t.Error("`z1` should have value 1")
	}
}

func TestUnaryPlusErr(t *testing.T) {
	expectError(t, "_test/src/unary", []string{"unary-plus-err-1.go"}, "invalid operand")
	expectError(t, "_test/src/unary", []string{"unary-plus-err-2.go"}, "invalid operand")
}

func TestUnaryMinus(t *testing.T) {
	p, err := compilePackage("_test/src/unary", []string{"unary-minus.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	B := p.Find("B").(*ast.Const)
	c := B.Init.(*ast.ConstValue)
	if !isUntyped(c.Typ) {
		t.Error("`B` should be untyped")
	}
	if _, ok := c.Value.(ast.UntypedInt); !ok {
		t.Error("`B` should be untyped integer")
	}
	if v, ok := toInt(c); !ok || v != -1 {
		t.Error("`B` should have value -1")
	}

	D := p.Find("D").(*ast.Const)
	c = D.Init.(*ast.ConstValue)
	if !isUntyped(c.Typ) {
		t.Error("`D` should be untyped")
	}
	if _, ok := c.Value.(ast.Rune); !ok {
		t.Error("`D` should be untyped rune")
	}
	if v, ok := toInt(c); !ok || v != -97 {
		t.Error("`D` should have value -97")
	}

	F := p.Find("F").(*ast.Const)
	c = F.Init.(*ast.ConstValue)
	if !isUntyped(c.Typ) {
		t.Error("`F` should be untyped")
	}
	if _, ok := c.Value.(ast.UntypedFloat); !ok {
		t.Error("`F` should be untyped float")
	}
	if v, ok := toFloat(c); !ok || v != -1.1 {
		t.Error("`F` should have value -1.1")
	}

	cB := p.Find("cB").(*ast.Const)
	c = cB.Init.(*ast.ConstValue)
	if c.Typ != ast.BuiltinInt {
		t.Error("`cB` should have type `int`")
	}
	if v, ok := toInt(c); !ok || v != -1 {
		t.Error("`B` should have value -1")
	}

	cD := p.Find("cD").(*ast.Const)
	c = cD.Init.(*ast.ConstValue)
	if c.Typ != ast.BuiltinInt32 {
		t.Error("`cD` should have type `rune`")
	}
	if v, ok := toInt(c); !ok || v != -97 {
		t.Error("`cD` should have value -97")
	}

	cF := p.Find("cF").(*ast.Const)
	c = cF.Init.(*ast.ConstValue)
	if c.Typ != ast.BuiltinFloat32 {
		t.Error("`cF` should have type `float32`")
	}
	if v, ok := toFloat(c); !ok || v != -1.125 {
		t.Error("`cF` should have value -1.125")
	}

	// FIXME: check complex constants

	vB := p.Find("vB").(*ast.Var)
	if vB.Type != ast.BuiltinInt {
		t.Error("`vB` should have type `int`")
	}

	vD := p.Find("vD").(*ast.Var)
	if vD.Type != ast.BuiltinInt32 {
		t.Error("`vD` should have type `rune`")
	}

	vF := p.Find("vF").(*ast.Var)
	if vF.Type != ast.BuiltinFloat64 {
		t.Error("`vF` should have type `float64`")
	}

	v := p.Find("v").(*ast.Var)
	if v.Type != ast.BuiltinUint {
		t.Error("`v` should have type `uint`")
	}

	z0 := p.Find("z0").(*ast.Const)
	if z0.Type != ast.BuiltinInt {
		t.Error("`z0` should have type `int`")
	}
	c = z0.Init.(*ast.ConstValue)
	if v, ok := c.Value.(ast.Int); !ok || int64(v) != -1 {
		t.Error("`z0` should have value -1")
	}
	z1 := p.Find("z1").(*ast.Const)
	if z1.Type != ast.BuiltinInt {
		t.Error("`z1` should have type `int`")
	}
	c = z1.Init.(*ast.ConstValue)
	if v, ok := c.Value.(ast.Int); !ok || int64(v) != -1 {
		t.Error("`z1` should have value -1")
	}
}

func TestUnaryMinusErr(t *testing.T) {
	expectError(t, "_test/src/unary", []string{"unary-minus-err-1.go"}, "invalid operand")
	expectError(t, "_test/src/unary", []string{"unary-minus-err-2.go"}, "invalid operand")
	expectError(t, "_test/src/unary", []string{"unary-minus-err-3.go"}, "invalid operand")
}

func TestNot(t *testing.T) {
	p, err := compilePackage("_test/src/unary", []string{"not.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	B := p.Find("B").(*ast.Const)
	c := B.Init.(*ast.ConstValue)
	if !isUntyped(c.Typ) {
		t.Error("`B` should be untyped")
	}
	v, ok := c.Value.(ast.Bool)
	if !ok {
		t.Fatal("`B` should be untyped boolean")
	}
	if !v {
		t.Error("`B` should have value false")
	}

	cB := p.Find("cB").(*ast.Const)
	c = cB.Init.(*ast.ConstValue)
	if c.Typ != ast.BuiltinBool {
		t.Error("`cB` shold have type `bool`")
	}
	if v = c.Value.(ast.Bool); v {
		t.Error("`cB` should have value false")
	}

	cD := p.Find("cD").(*ast.Const)
	c = cD.Init.(*ast.ConstValue)
	if c.Typ != ast.BuiltinBool {
		t.Error("`cD` shold have type `bool`")
	}
	if v = c.Value.(ast.Bool); !v {
		t.Error("`cD` should have value true")
	}

	D := p.Find("D").(*ast.Const)
	c = D.Init.(*ast.ConstValue)
	if !isUntyped(c.Typ) {
		t.Error("`D` should be untyped")
	}
	v, ok = c.Value.(ast.Bool)
	if !ok {
		t.Fatal("`D` should be untyped boolean")
	}
	if v {
		t.Error("`D` should have value false")
	}

	vB := p.Find("vB").(*ast.Var)
	if vB.Type != ast.BuiltinBool {
		t.Error("`B` should have type `bool`")
	}

	z0 := p.Find("z0").(*ast.Const)
	if z0.Type != ast.BuiltinBool {
		t.Error("`z0` should have type `bool`")
	}
	c = z0.Init.(*ast.ConstValue)
	if v, ok := c.Value.(ast.Bool); !ok || v == true {
		t.Error("`z0` should have value false")
	}
	z1 := p.Find("z1").(*ast.Const)
	if z1.Type != ast.BuiltinBool {
		t.Error("`z1` should have type `bool`")
	}
	c = z1.Init.(*ast.ConstValue)
	if v, ok := c.Value.(ast.Bool); !ok || v == true {
		t.Error("`z1` should have value false")
	}

}

func TestNotErr(t *testing.T) {
	expectError(t, "_test/src/unary", []string{"not-err-1.go"}, "invalid operand")
	expectError(t, "_test/src/unary", []string{"not-err-2.go"}, "invalid operand")
	expectError(t, "_test/src/unary", []string{"not-err-3.go"}, "invalid operand")
}

func TestCompl(t *testing.T) {
	p, err := compilePackage("_test/src/unary", []string{"compl.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	B := p.Find("B").(*ast.Const)
	c := B.Init.(*ast.ConstValue)
	if !isUntyped(c.Typ) {
		t.Error("`B` should be untyped")
	}
	if _, ok := c.Value.(ast.UntypedInt); !ok {
		t.Fatal("`B` should be untyped integer")
	}
	if v, ok := toInt(c); !ok || v != -2 {
		t.Error("`B` should have value -2")
	}

	cB := p.Find("cB").(*ast.Const)
	c = cB.Init.(*ast.ConstValue)
	if typ, ok := c.Typ.(*ast.TypeDecl); !ok || typ.Name != "T" {
		t.Error("`cB` should have type `T`")
	}
	if v, ok := toInt(c); !ok || v != 0xfffe {
		t.Error("`cB` should have value 0xfffe")
	}

	cu8 := p.Find("cu8").(*ast.Const)
	c = cu8.Init.(*ast.ConstValue)
	if c.Typ != ast.BuiltinUint8 {
		t.Error("`cu8` should have type `uint8`")
	}
	if v, ok := toInt(c); !ok || v != 0xfe {
		t.Error("`cu8` should have value 0xfe")
	}

	cu16 := p.Find("cu16").(*ast.Const)
	c = cu16.Init.(*ast.ConstValue)
	if c.Typ != ast.BuiltinUint16 {
		t.Error("`cu16` should have type `uint16`")
	}
	if v, ok := toInt(c); !ok || v != 0xfffe {
		t.Error("`cu16` should have value 0xfffe")
	}

	cu32 := p.Find("cu32").(*ast.Const)
	c = cu32.Init.(*ast.ConstValue)
	if c.Typ != ast.BuiltinUint32 {
		t.Error("`cu32` should have type `uint32`")
	}
	if v, ok := toInt(c); !ok || v != 0xfffffffe {
		t.Error("`cu32` should have value 0xfffffffe")
	}

	cu64 := p.Find("cu64").(*ast.Const)
	c = cu64.Init.(*ast.ConstValue)
	if c.Typ != ast.BuiltinUint64 {
		t.Error("`cu64` should have type `uint64`")
	}
	if v := uint64(c.Value.(ast.Int)); v != 0xfffffffffffffffe {
		t.Error("`cu64` should have value 0xfffffffffffffffe")
	}

	cu := p.Find("cu").(*ast.Const)
	c = cu.Init.(*ast.ConstValue)
	if c.Typ != ast.BuiltinUint {
		t.Error("`cu` should have type `uint`")
	}
	if v := uint64(c.Value.(ast.Int)); v != 0xfffffffffffffffe {
		t.Error("`cu` should have value 0xfffffffffffffffe")
	}

	cp := p.Find("cp").(*ast.Const)
	c = cp.Init.(*ast.ConstValue)
	if c.Typ != ast.BuiltinUintptr {
		t.Error("`cp` should have type `uintptr`")
	}
	if v := uint64(c.Value.(ast.Int)); v != 0xfffffffffffffffe {
		t.Error("`cp` should have value 0xfffffffffffffffe")
	}

	ci8 := p.Find("ci8").(*ast.Const)
	c = ci8.Init.(*ast.ConstValue)
	if c.Typ != ast.BuiltinInt8 {
		t.Error("`ci8` should have type `int8`")
	}
	if v, ok := toInt(c); !ok || v != -2 {
		t.Error("`ci8` should have value -2")
	}

	ci16 := p.Find("ci16").(*ast.Const)
	c = ci16.Init.(*ast.ConstValue)
	if c.Typ != ast.BuiltinInt16 {
		t.Error("`ci16` should have type `int16`")
	}
	if v, ok := toInt(c); !ok || v != -2 {
		t.Error("`ci16` should have value -2")
	}

	ci32 := p.Find("ci32").(*ast.Const)
	c = ci32.Init.(*ast.ConstValue)
	if c.Typ != ast.BuiltinInt32 {
		t.Error("`ci32` should have type `int32`")
	}
	if v, ok := toInt(c); !ok || v != -2 {
		t.Error("`ci32` should have value -2")
	}

	ci64 := p.Find("ci64").(*ast.Const)
	c = ci64.Init.(*ast.ConstValue)
	if c.Typ != ast.BuiltinInt64 {
		t.Error("`ci64` should have type `int64`")
	}
	if v, ok := toInt(c); !ok || v != -2 {
		t.Error("`ci64` should have value -2")
	}

	ci := p.Find("ci").(*ast.Const)
	c = ci.Init.(*ast.ConstValue)
	if c.Typ != ast.BuiltinInt {
		t.Error("`ci` should have type `int`")
	}
	if v, ok := toInt(c); !ok || v != -2 {
		t.Error("`ci` should have value -2")
	}

	cr0 := p.Find("cr0").(*ast.Const)
	c = cr0.Init.(*ast.ConstValue)
	if !isUntyped(c.Typ) {
		t.Error("`cr0` should be untyped")
	}
	if _, ok := c.Value.(ast.Rune); !ok {
		t.Error("`cr0` should be untyped rune")
	}
	if v, ok := toInt(c); !ok || v != -98 {
		t.Error("`cr0` should have value -98")
	}

	cr1 := p.Find("cr1").(*ast.Const)
	c = cr1.Init.(*ast.ConstValue)
	if c.Typ != ast.BuiltinInt32 {
		t.Error("`cr1` should have type `rune`")
	}
	if v, ok := toInt(c); !ok || v != -98 {
		t.Error("`cr1` should have value -98")
	}

	vB := p.Find("vB").(*ast.Var)
	if typ, ok := vB.Type.(*ast.TypeDecl); !ok || typ.Name != "T" {
		t.Error("`vB` should have type `T`")
	}

	z0 := p.Find("z0").(*ast.Const)
	c = z0.Init.(*ast.ConstValue)
	if c.Typ != ast.BuiltinInt {
		t.Error("`z0` should have type `int`")
	}
	if v, ok := toInt(c); !ok || v != -2 {
		t.Error("`z0` should have value -2")
	}
	z1 := p.Find("z1").(*ast.Const)
	c = z1.Init.(*ast.ConstValue)
	if c.Typ != ast.BuiltinInt {
		t.Error("`z1` should have type `int`")
	}
	if v, ok := toInt(c); !ok || v != -2 {
		t.Error("`z1` should have value -2")
	}
}

func TestComplErr(t *testing.T) {
	expectError(t, "_test/src/unary", []string{"compl-err-1.go"}, "invalid operand")
	expectError(t, "_test/src/unary", []string{"compl-err-2.go"}, "invalid operand")
	expectError(t, "_test/src/unary", []string{"compl-err-3.go"}, "invalid operand")
}

func TestIndirection(t *testing.T) {
	p, err := compilePackage("_test/src/unary", []string{"deref.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	B := p.Find("B").(*ast.Var)
	if typ, ok := B.Type.(*ast.TypeDecl); !ok || typ.Name != "T" {
		t.Error("`B` should have type `T`")
	}
}

func TestIndirectionlErr(t *testing.T) {
	expectError(t, "_test/src/unary", []string{"deref-err-1.go"}, "invalid operand")
	expectError(t, "_test/src/unary", []string{"deref-err-2.go"}, "invalid operand")
}

func TestAddr(t *testing.T) {
	p, err := compilePackage("_test/src/unary", []string{"addr.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	T := p.Find("T").(*ast.TypeDecl)
	S := p.Find("S").(*ast.TypeDecl)

	B := p.Find("B").(*ast.Var)
	if ptr, ok := B.Type.(*ast.PtrType); !ok || ptr.Base != T {
		t.Error("`B` should have type `*T`")
	}

	C := p.Find("C").(*ast.Var)
	if ptr, ok := C.Type.(*ast.PtrType); !ok || ptr.Base != T {
		t.Error("`C` should have type `*T`")
	}

	E := p.Find("E").(*ast.Var)
	if ptr, ok := E.Type.(*ast.PtrType); !ok || ptr.Base != S {
		t.Error("`E` should have type `*S`")
	}

	F := p.Find("F").(*ast.Var)
	if ptr, ok := F.Type.(*ast.PtrType); !ok || ptr.Base != ast.BuiltinInt {
		t.Error("`F` should have type `*int`")
	}

	H := p.Find("H").(*ast.Var)
	if ptr, ok := H.Type.(*ast.PtrType); !ok || ptr.Base != T {
		t.Error("`H` should have type `*T`")
	}

	I := p.Find("I").(*ast.Var)
	if ptr, ok := I.Type.(*ast.PtrType); !ok || ptr.Base != ast.BuiltinInt {
		t.Error("`I` should have type `*int`")
	}

	X := p.Find("X").(*ast.Var)
	if ptr, ok := X.Type.(*ast.PtrType); !ok || ptr.Base != S {
		t.Error("`X` should have type `*S`")
	}

	Y := p.Find("Y").(*ast.Var)
	ptr, ok := Y.Type.(*ast.PtrType)
	if !ok {
		t.Error("`Y` should have type `*[...]T`")
	}
	if a, ok := ptr.Base.(*ast.ArrayType); !ok || a.Elt != T {
		t.Error("`Y` should have type `*[...]T`")
	}

	Z := p.Find("Z").(*ast.Var)
	ptr, ok = Z.Type.(*ast.PtrType)
	if !ok {
		t.Error("`Z` should have type `*[]T`")
	}
	if s, ok := ptr.Base.(*ast.SliceType); !ok || s.Elt != T {
		t.Error("`Y` should have type `*[]T`")
	}
}

func TestAddrErr(t *testing.T) {
	expectError(t, "_test/src/unary", []string{"addr-err-1.go"}, "invalid operand")
	expectError(t, "_test/src/unary", []string{"addr-err-2.go"}, "invalid operand")
	expectError(t, "_test/src/unary", []string{"addr-err-3.go"}, "invalid operand")
	expectError(t, "_test/src/unary", []string{"addr-err-4.go"}, "invalid operand")
}

func TestRecv(t *testing.T) {
	p, err := compilePackage("_test/src/unary", []string{"recv.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	T := p.Find("T").(*ast.TypeDecl)

	A := p.Find("A").(*ast.Var)
	if A.Type != T {
		t.Error("`A` should have type `T`")
	}

	B := p.Find("B").(*ast.Var)
	if B.Type != T {
		t.Error("`B` should have type `T`")
	}

	ok := p.Find("ok").(*ast.Var)
	if ok.Type != ast.BuiltinBool {
		t.Error("`ok` should have type `bool`")
	}
}

func TestRecvErr(t *testing.T) {
	expectError(t, "_test/src/unary", []string{"recv-err-1.go"}, "invalid operand")
	expectError(t, "_test/src/unary", []string{"recv-err-2.go"}, "invalid operand")
	expectError(t, "_test/src/unary", []string{"recv-err-3.go"}, "invalid operand")
}

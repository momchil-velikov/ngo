package ast

var UniverseScope *_UniverseScope

var (
	BuiltinVoidType   *BuiltinType
	BuiltinNilType    *BuiltinType
	BuiltinBool       *BuiltinType
	BuiltinUint8      *BuiltinType
	BuiltinUint16     *BuiltinType
	BuiltinUint32     *BuiltinType
	BuiltinUint64     *BuiltinType
	BuiltinInt8       *BuiltinType
	BuiltinInt16      *BuiltinType
	BuiltinInt32      *BuiltinType
	BuiltinInt64      *BuiltinType
	BuiltinFloat32    *BuiltinType
	BuiltinFloat64    *BuiltinType
	BuiltinComplex64  *BuiltinType
	BuiltinComplex128 *BuiltinType
	BuiltinUint       *BuiltinType
	BuiltinInt        *BuiltinType
	BuiltinUintptr    *BuiltinType
	BuiltinString     *BuiltinType

	BuiltinNil   Expr
	BuiltinTrue  Expr
	BuiltinFalse Expr
	BuiltinIota  Expr

	BuiltinAppend  *FuncDecl
	BuiltinCap     *FuncDecl
	BuiltinClose   *FuncDecl
	BuiltinComplex *FuncDecl
	BuiltinCopy    *FuncDecl
	BuiltinDelete  *FuncDecl
	BuiltinImag    *FuncDecl
	BuiltinLen     *FuncDecl
	BuiltinMake    *FuncDecl
	BuiltinNew     *FuncDecl
	BuiltinPanic   *FuncDecl
	BuiltinPrint   *FuncDecl
	BuiltinPrintln *FuncDecl
	BuiltinReal    *FuncDecl
	BuiltinRecover *FuncDecl
)

func init() {
	BuiltinVoidType = &BuiltinType{BUILTIN_VOID_TYPE}
	BuiltinNilType = &BuiltinType{BUILTIN_NIL_TYPE}
	BuiltinBool = &BuiltinType{BUILTIN_BOOL}
	BuiltinUint8 = &BuiltinType{BUILTIN_UINT8}
	BuiltinUint16 = &BuiltinType{BUILTIN_UINT16}
	BuiltinUint32 = &BuiltinType{BUILTIN_UINT32}
	BuiltinUint64 = &BuiltinType{BUILTIN_UINT64}
	BuiltinInt8 = &BuiltinType{BUILTIN_INT8}
	BuiltinInt16 = &BuiltinType{BUILTIN_INT16}
	BuiltinInt32 = &BuiltinType{BUILTIN_INT32}
	BuiltinInt64 = &BuiltinType{BUILTIN_INT64}
	BuiltinFloat32 = &BuiltinType{BUILTIN_FLOAT32}
	BuiltinFloat64 = &BuiltinType{BUILTIN_FLOAT64}
	BuiltinComplex64 = &BuiltinType{BUILTIN_COMPLEX64}
	BuiltinComplex128 = &BuiltinType{BUILTIN_COMPLEX128}
	BuiltinUint = &BuiltinType{BUILTIN_UINT}
	BuiltinInt = &BuiltinType{BUILTIN_INT}
	BuiltinUintptr = &BuiltinType{BUILTIN_UINTPTR}
	BuiltinString = &BuiltinType{BUILTIN_STRING}

	UniverseScope = &_UniverseScope{dcl: make(map[string]Symbol)}
	for _, c := range []struct {
		name string
		typ  Type
	}{
		{"bool", BuiltinBool},
		{"byte", BuiltinUint8},
		{"uint8", BuiltinUint8},
		{"uint16", BuiltinUint16},
		{"uint32", BuiltinUint32},
		{"uint64", BuiltinUint64},
		{"int8", BuiltinInt8},
		{"int16", BuiltinInt16},
		{"rune", BuiltinInt32},
		{"int32", BuiltinInt32},
		{"int64", BuiltinInt64},
		{"float32", BuiltinFloat32},
		{"float64", BuiltinFloat64},
		{"complex64", BuiltinComplex64},
		{"complex128", BuiltinComplex128},
		{"uint", BuiltinUint},
		{"int", BuiltinInt},
		{"uintptr", BuiltinUintptr},
		{"string", BuiltinString},
	} {
		UniverseScope.dcl[c.name] = &TypeDecl{Name: c.name, Type: c.typ}
	}

	BuiltinNil = &ConstValue{Value: BuiltinValue(BUILTIN_NIL)}
	BuiltinTrue = &ConstValue{Value: Bool(true)}
	BuiltinFalse = &ConstValue{Value: Bool(false)}
	BuiltinIota = &ConstValue{Value: BuiltinValue(BUILTIN_IOTA)}

	for _, c := range []struct {
		name string
		x    Expr
	}{
		{"nil", BuiltinNil},
		{"true", BuiltinTrue},
		{"false", BuiltinFalse},
		{"iota", BuiltinIota},
	} {
		UniverseScope.dcl[c.name] = &Const{Name: c.name, Init: c.x}
	}

	BuiltinAppend = &FuncDecl{Name: "append"}
	BuiltinCap = &FuncDecl{Name: "cap"}
	BuiltinClose = &FuncDecl{Name: "close"}
	BuiltinComplex = &FuncDecl{Name: "complex"}
	BuiltinCopy = &FuncDecl{Name: "copy"}
	BuiltinDelete = &FuncDecl{Name: "delete"}
	BuiltinImag = &FuncDecl{Name: "imag"}
	BuiltinLen = &FuncDecl{Name: "len"}
	BuiltinMake = &FuncDecl{Name: "make"}
	BuiltinNew = &FuncDecl{Name: "new"}
	BuiltinPanic = &FuncDecl{Name: "panic"}
	BuiltinPrint = &FuncDecl{Name: "print"}
	BuiltinPrintln = &FuncDecl{Name: "println"}
	BuiltinReal = &FuncDecl{Name: "real"}
	BuiltinRecover = &FuncDecl{Name: "recover"}
	for _, c := range []struct {
		name string
		fn   *FuncDecl
	}{

		{"append", BuiltinAppend},
		{"cap", BuiltinCap},
		{"close", BuiltinClose},
		{"complex", BuiltinComplex},
		{"copy", BuiltinCopy},
		{"delete", BuiltinDelete},
		{"imag", BuiltinImag},
		{"len", BuiltinLen},
		{"make", BuiltinMake},
		{"new", BuiltinNew},
		{"panic", BuiltinPanic},
		{"print", BuiltinPrint},
		{"println", BuiltinPrintln},
		{"real", BuiltinReal},
		{"recover", BuiltinRecover},
	} {
		UniverseScope.dcl[c.name] = c.fn
	}
}

var blank = &Var{Name: "_"}
var Blank = &OperandName{Decl: blank}

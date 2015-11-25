package ast

var UniverseScope *_UniverseScope

var (
	BuiltinNil        Type
	BuiltinBool       Type
	BuiltinUint8      Type
	BuiltinUint16     Type
	BuiltinUint32     Type
	BuiltinUint64     Type
	BuiltinInt8       Type
	BuiltinInt16      Type
	BuiltinInt32      Type
	BuiltinInt64      Type
	BuiltinFloat32    Type
	BuiltinFloat64    Type
	BuiltinComplex64  Type
	BuiltinComplex128 Type
	BuiltinUint       Type
	BuiltinInt        Type
	BuiltinUintptr    Type
	BuiltinString     Type
)

func init() {
	BuiltinNil = &BuiltinType{BUILTIN_NIL}
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
		{"#nil", BuiltinNil},
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
}

var Blank = &Var{Name: "_"}

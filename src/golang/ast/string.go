package ast

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

const floatPrecision = 6

func intToString(typ *BuiltinType, v uint64) string {
	if typ.IsSigned() {
		return strconv.FormatInt(int64(v), 10)
	} else {
		return strconv.FormatUint(v, 10)
	}
}

func removeTrailingZeroes(s string) string {
	i := strings.IndexRune(s, '.')
	if i == -1 {
		return s
	}

	j := len(s)
	for j-1 > i {
		j--
		if s[j] != '0' {
			break
		}
	}
	return s[0 : j+1]
}

func floatToString(typ *BuiltinType, v float64) string {
	return removeTrailingZeroes(strconv.FormatFloat(v, 'f', floatPrecision, 64))
}

func complexToString(typ *BuiltinType, v complex128) string {
	if typ.Kind == BUILTIN_COMPLEX128 {
		typ = BuiltinFloat64
	} else {
		typ = BuiltinFloat32
	}
	if real(v) == 0.0 {
		return floatToString(typ, imag(v)) + "i"
	} else {
		re, im := floatToString(typ, real(v)), floatToString(typ, imag(v))
		return strings.Join([]string{"(", re, " + ", im, "i)"}, "")
	}
}

func builtinType(typ Type) *BuiltinType {
	for {
		switch t := typ.(type) {
		case *BuiltinType:
			return t
		case *TypeDecl:
			typ = t.Type
		default:
			return nil
		}
	}
}

func (c *ConstValue) String() string {
	switch v := c.Value.(type) {
	case Bool:
		if bool(v) {
			return "true"
		} else {
			return "false"
		}
	case Rune:
		return strconv.QuoteRune(rune(v.Int64()))
	case UntypedInt:
		return v.String()
	case Int:
		return intToString(builtinType(c.Typ), uint64(v))
	case UntypedFloat:
		return removeTrailingZeroes(v.Text('f', floatPrecision))
	case Float:
		return floatToString(builtinType(c.Typ), float64(v))
	case UntypedComplex:
		re := removeTrailingZeroes(v.Re.Text('f', floatPrecision))
		im := removeTrailingZeroes(v.Im.Text('f', floatPrecision))
		if re == "0.0" {
			return im + "i"
		} else {
			return strings.Join([]string{"(", re, " + ", im, "i)"}, "")
		}
	case Complex:
		return complexToString(builtinType(c.Typ), complex128(v))
	case String:
		return "\"" + string(v) + "\""
	default:
		panic("not reached")
	}
}

func (*Error) String() string {
	return "<error>"
}

func (id *QualifiedId) String() string {
	if len(id.Pkg) > 0 {
		return id.Pkg + "." + id.Id
	} else {
		return id.Id
	}
}

func (dcl *TypeDecl) String() string { return dcl.Name }

func (typ *BuiltinType) String() string {
	switch typ.Kind {
	case BUILTIN_VOID_TYPE:
		return "void"
	case BUILTIN_NIL_TYPE:
		return "nil"
	case BUILTIN_BOOL:
		return "bool"
	case BUILTIN_UINT8:
		return "uint8"
	case BUILTIN_UINT16:
		return "uint16"
	case BUILTIN_UINT32:
		return "uint32"
	case BUILTIN_UINT64:
		return "uint64"
	case BUILTIN_INT8:
		return "int8"
	case BUILTIN_INT16:
		return "int16"
	case BUILTIN_INT32:
		return "int32"
	case BUILTIN_INT64:
		return "int64"
	case BUILTIN_FLOAT32:
		return "float32"
	case BUILTIN_FLOAT64:
		return "float64"
	case BUILTIN_COMPLEX64:
		return "complex64"
	case BUILTIN_COMPLEX128:
		return "complex128"
	case BUILTIN_UINT:
		return "uint"
	case BUILTIN_INT:
		return "int"
	case BUILTIN_UINTPTR:
		return "uintptr"
	case BUILTIN_STRING:
		return "string"
	case BUILTIN_UNTYPED_BOOL:
		return "untyped bool"
	case BUILTIN_UNTYPED_RUNE:
		return "untyped rune"
	case BUILTIN_UNTYPED_INT:
		return "untyped int"
	case BUILTIN_UNTYPED_FLOAT:
		return "untyped float"
	case BUILTIN_UNTYPED_COMPLEX:
		return "untyped complex"
	case BUILTIN_UNTYPED_STRING:
		return "untyped string"
	case BUILTIN_DEFAULT:
		return "untyped default"
	default:
		panic("not reached")
	}
}

func (typ *ArrayType) String() string {
	if n, ok := typ.Dim.(*ConstValue); ok {
		return fmt.Sprintf("[%s]%s", n, typ.Elt)
	} else {
		return "[...]" + typ.Elt.String()
	}
}

func (typ *SliceType) String() string {
	return "[]" + typ.Elt.String()
}

func (typ *PtrType) String() string {
	return "*" + typ.Base.String()
}

func (typ *MapType) String() string {
	return fmt.Sprintf("map[%s]%s", typ.Key, typ.Elt)
}

func (typ *ChanType) String() string {
	if typ.Send && typ.Recv {
		return "chan " + typ.Elt.String()
	} else if typ.Send {
		return "chan<- " + typ.Elt.String()
	} else {
		return "<-chan " + typ.Elt.String()
	}
}

func (typ *StructType) String() string {
	if len(typ.Fields) == 0 {
		return "struct{}"
	} else {
		return "struct{...}"
	}
}

func (typ *TupleType) String() string {
	if n := len(typ.Types); n > 0 {
		b := bytes.Buffer{}
		b.WriteRune('<')
		b.WriteString(typ.Types[0].String())
		for i := 1; i < n; i++ {
			b.WriteString(", ")
			b.WriteString(typ.Types[i].String())
		}
		b.WriteRune('>')
		return b.String()
	} else {
		return "<>"
	}
}

func (typ *FuncType) String() string {
	b := bytes.Buffer{}
	if n := len(typ.Params); n > 0 {
		b.WriteString("func(")
		b.WriteString(typ.Params[0].Type.String())
		for i := 1; i < n; i++ {
			b.WriteString(", ")
			b.WriteString(typ.Params[i].Type.String())
		}
		b.WriteRune(')')
	} else {
		b.WriteString("func()")
	}
	if n := len(typ.Returns); n > 1 {
		b.WriteString(" (")
		b.WriteString(typ.Returns[0].Type.String())
		for i := 1; i < n; i++ {
			b.WriteString(", ")
			b.WriteString(typ.Returns[i].Type.String())
		}
		b.WriteRune(')')
	} else if n == 1 {
		b.WriteRune(' ')
		b.WriteString(typ.Returns[0].Type.String())
	}
	return b.String()
}

func (typ *InterfaceType) String() string {
	if len(typ.Embedded) == 0 && len(typ.Methods) == 0 {
		return "interface{}"
	} else {
		return "interface{...}"
	}
}

func (c *ConstValue) TypeString() string {
	return c.Typ.String()
}

func (op Operation) String() string {
	switch uint(op) {
	case PLUS:
		return "+"
	case MINUS:
		return "-"
	case MUL:
		return "*"
	case DIV:
		return "/"
	case REM:
		return "%"
	case BITAND:
		return "&"
	case BITOR:
		return "|"
	case BITXOR:
		return "^"
	case LT:
		return "<"
	case GT:
		return ">"
	case NOT:
		return "!"
	case SHL:
		return "<<"
	case SHR:
		return ">>"
	case ANDN:
		return "&^"
	case AND:
		return "&&"
	case OR:
		return "||"
	case RECV:
		return "<-"
	case EQ:
		return "=="
	case NE:
		return "!="
	case LE:
		return "<="
	case GE:
		return "=>"
	default:
		panic("not reached")
	}
}

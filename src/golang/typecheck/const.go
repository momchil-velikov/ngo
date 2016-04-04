package typecheck

import (
	"fmt"
	"golang/ast"
	"math"
	"math/big"
	"unicode/utf8"
)

func convertBool(dst *ast.BuiltinType, val ast.Bool) ast.Value {
	if dst.Kind != ast.BUILTIN_BOOL {
		return nil
	}
	return val
}

func convertInt(dst *ast.BuiltinType, src *ast.BuiltinType, val uint64) ast.Value {
	sval := int64(val)
	switch dst.Kind {
	case ast.BUILTIN_BOOL:
		return nil
	case ast.BUILTIN_INT8:
		if src.IsSigned() {
			if sval < math.MinInt8 || sval > math.MaxInt8 {
				return nil
			}
		} else if val > math.MaxInt8 {
			return nil
		}
		return ast.Int(val)
	case ast.BUILTIN_INT16:
		if src.IsSigned() {
			if sval < math.MinInt16 || sval > math.MaxInt16 {
				return nil
			}
		} else if val > math.MaxInt16 {
			return nil
		}
		return ast.Int(val)
	case ast.BUILTIN_INT32:
		if src.IsSigned() {
			if sval < math.MinInt32 || sval > math.MaxInt32 {
				return nil
			}
		} else if val > math.MaxInt32 {
			return nil
		}
		return ast.Int(val)
	case ast.BUILTIN_INT:
		fallthrough // XXX: 64-bit assumption
	case ast.BUILTIN_INT64:
		if !src.IsSigned() && val > math.MaxInt64 {
			return nil
		}
		return ast.Int(val)
	case ast.BUILTIN_UINT8:
		if src.IsSigned() {
			if sval < 0 || sval > math.MaxUint8 {
				return nil
			}
		} else if val > math.MaxUint8 {
			return nil
		}
		return ast.Int(val)
	case ast.BUILTIN_UINT16:
		if src.IsSigned() {
			if sval < 0 || sval > math.MaxUint16 {
				return nil
			}
		} else if val > math.MaxUint16 {
			return nil
		}
		return ast.Int(val)
	case ast.BUILTIN_UINT32:
		if src.IsSigned() {
			if sval < 0 || sval > math.MaxUint32 {
				return nil
			}
		} else if val > math.MaxUint32 {
			return nil
		}
		return ast.Int(val)
	case ast.BUILTIN_UINT:
		fallthrough // XXX 64-bit assumption
	case ast.BUILTIN_UINTPTR:
		fallthrough // XXX 64-bit assumption
	case ast.BUILTIN_UINT64:
		if src.IsSigned() && sval < 0 {
			return nil
		}
		return ast.Int(val)
	case ast.BUILTIN_COMPLEX64:
		fallthrough
	case ast.BUILTIN_FLOAT32:
		var f float32
		if src.IsSigned() && sval < 0 {
			f = float32(sval)
		} else {
			f = float32(val)
		}
		if dst.Kind == ast.BUILTIN_FLOAT32 {
			return ast.Float(f)
		} else {
			return ast.Complex(complex(f, 0))
		}
	case ast.BUILTIN_COMPLEX128:
		fallthrough
	case ast.BUILTIN_FLOAT64:
		var f float64
		if src.IsSigned() && sval < 0 {
			f = float64(sval)
		} else {
			f = float64(val)
		}
		if dst.Kind == ast.BUILTIN_FLOAT64 {
			return ast.Float(f)
		} else {
			return ast.Complex(complex(f, 0))
		}
	case ast.BUILTIN_STRING:
		r := rune(val)
		if src.IsSigned() {
			if sval <= 0 || sval > math.MaxInt32 || !utf8.ValidRune(r) {
				r = 0xfffd
			}
		} else if val == 0 || val > math.MaxInt32 || !utf8.ValidRune(r) {
			r = 0xfffd
		}
		return ast.String(r)
	default:
		panic("not reached")
	}
}

var (
	MinInt64  = big.NewInt(math.MinInt64)
	MaxInt64  = big.NewInt(math.MaxInt64)
	MaxUint64 = new(big.Int).SetUint64(math.MaxUint64)
)

const MaxShift = 511

func convertUntypedInt(dst *ast.BuiltinType, val *big.Int) ast.Value {
	if dst.IsInteger() || dst.Kind == ast.BUILTIN_STRING {
		// If the destination type is integral, a successful conversion it not
		// possible if the untyped constant is outside the range of (u)int64.
		if val.Cmp(MinInt64) < 0 || val.Cmp(MaxUint64) > 0 {
			if dst.Kind == ast.BUILTIN_STRING {
				return ast.String(0xfffd)
			} else {
				return nil
			}
		}
		if val.Sign() < 0 {
			return convertInt(dst, ast.BuiltinInt64, uint64(val.Int64()))
		} else {
			return convertInt(dst, ast.BuiltinUint64, val.Uint64())
		}
	}
	switch dst.Kind {
	case ast.BUILTIN_BOOL:
		return nil
	case ast.BUILTIN_COMPLEX64:
		fallthrough
	case ast.BUILTIN_FLOAT32:
		f := new(big.Float).SetInt(val)
		f32, _ := f.Float32()
		if math.IsInf(float64(f32), 0) {
			return nil
		}
		if dst.Kind == ast.BUILTIN_FLOAT32 {
			return ast.Float(f32)
		} else {
			return ast.Complex(complex(float64(f32), 0))
		}
	case ast.BUILTIN_COMPLEX128:
		fallthrough
	case ast.BUILTIN_FLOAT64:
		f := new(big.Float).SetInt(val)
		f64, _ := f.Float64()
		if math.IsInf(f64, 0) {
			return nil
		}
		if dst.Kind == ast.BUILTIN_FLOAT64 {
			return ast.Float(f64)
		} else {
			return ast.Complex(complex(f64, 0))
		}
	default:
		panic("not reached")
	}
}

func floatToUint64(x float64) (uint64, bool) {
	if x < 0 {
		return 0, false
	}
	b := math.Float64bits(x)
	e := int((b >> 52) & 0x7ff)
	if e == 0 {
		return 0, true // zero
	}
	f := (b & 0x000fffffffffffff) | 0x0010000000000000
	e = e - 1023 - 52
	if e < -52 {
		return 0, true // too small
	}
	if e > 11 {
		return 0, false
	}
	if e < 0 {
		return f >> uint(-e), true
	} else {
		return f << uint(e), true
	}
}

func floatToInt64(x float64) (int64, bool) {
	neg := x < 0
	u, ok := uint64(0), false
	if neg {
		u, ok = floatToUint64(-x)
	} else {
		u, ok = floatToUint64(x)
	}
	if !ok {
		return 0, false
	}
	if neg && u > math.MaxInt64+1 {
		return 0, false
	}
	if !neg && u > math.MaxInt64 {
		return 0, false
	}
	if neg {
		return -int64(u), true
	} else {
		return int64(u), true
	}
}

func convertFloat(dst *ast.BuiltinType, val float64) ast.Value {
	if dst.IsInteger() {
		if val != math.Floor(val) {
			return nil
		}
		if val < 0 {
			v, ok := floatToInt64(val)
			if !ok {
				return nil
			}
			return convertInt(dst, ast.BuiltinInt64, uint64(v))
		} else {
			v, ok := floatToUint64(val)
			if !ok {
				return nil
			}
			return convertInt(dst, ast.BuiltinUint64, v)
		}
	}
	switch dst.Kind {
	case ast.BUILTIN_BOOL:
		return nil
	case ast.BUILTIN_FLOAT32:
		val = float64(float32(val))
		fallthrough
	case ast.BUILTIN_FLOAT64:
		return ast.Float(val)
	case ast.BUILTIN_COMPLEX64:
		val = float64(float32(val))
		fallthrough
	case ast.BUILTIN_COMPLEX128:
		return ast.Complex(complex(val, 0))
	case ast.BUILTIN_STRING:
		return nil
	default:
		panic("not reached")
	}
}

func convertComplex(dst *ast.BuiltinType, re float64, im float64) ast.Value {
	if dst.IsInteger() ||
		dst.Kind == ast.BUILTIN_FLOAT32 || dst.Kind == ast.BUILTIN_FLOAT64 {
		if im != 0.0 {
			return nil
		}
		return convertFloat(dst, re)
	}
	switch dst.Kind {
	case ast.BUILTIN_BOOL:
		return nil
	case ast.BUILTIN_COMPLEX64:
		re = float64(float32(re))
		im = float64(float32(im))
		fallthrough
	case ast.BUILTIN_COMPLEX128:
		return ast.Complex(complex(re, im))
	case ast.BUILTIN_STRING:
		return nil
	default:
		panic("not reached")
	}
}

func convertConst(dst, src *ast.BuiltinType, val ast.Value) ast.Value {
	if dst.Kind == ast.BUILTIN_NIL_TYPE || dst.Kind == ast.BUILTIN_VOID_TYPE {
		panic("not reached")
	}

	var res ast.Value
	switch src.Kind {
	case ast.BUILTIN_BOOL, ast.BUILTIN_UNTYPED_BOOL:
		res = convertBool(dst, val.(ast.Bool))
	case ast.BUILTIN_UNTYPED_RUNE:
		res = convertUntypedInt(dst, val.(ast.Rune).Int)
	case ast.BUILTIN_UNTYPED_INT:
		res = convertUntypedInt(dst, val.(ast.UntypedInt).Int)
	case ast.BUILTIN_UNTYPED_FLOAT:
		v := val.(ast.UntypedFloat)
		if dst.IsInteger() {
			if v.IsInt() {
				i, _ := v.Int(nil)
				res = convertUntypedInt(dst, i)
			}
		} else {
			x, _ := v.Float64()
			if !math.IsInf(x, 0) {
				res = convertFloat(dst, x)
			}
		}
	case ast.BUILTIN_FLOAT32, ast.BUILTIN_FLOAT64:
		res = convertFloat(dst, float64(val.(ast.Float)))
	case ast.BUILTIN_UNTYPED_COMPLEX:
		v := val.(ast.UntypedComplex)
		if dst.IsInteger() {
			if v.Re.IsInt() && v.Im.Sign() == 0 {
				i, _ := v.Re.Int(nil)
				res = convertUntypedInt(dst, i)
			} else {
				res = nil
			}
		} else {
			re, _ := v.Re.Float64()
			im, _ := v.Im.Float64()
			if math.IsInf(re, 0) || math.IsInf(im, 0) {
				res = nil
			} else {
				res = convertComplex(dst, re, im)
			}
		}
	case ast.BUILTIN_COMPLEX64, ast.BUILTIN_COMPLEX128:
		v := val.(ast.Complex)
		re, im := real(v), imag(v)
		res = convertComplex(dst, re, im)
	case ast.BUILTIN_STRING, ast.BUILTIN_UNTYPED_STRING:
		if dst.Kind == ast.BUILTIN_STRING {
			res = val
		} else {
			res = nil
		}
	default:
		if !src.IsInteger() {
			panic("not reached")
		}
		res = convertInt(dst, src, uint64(val.(ast.Int)))
	}
	return res
}

func toInt(c *ast.ConstValue) (int64, bool) {
	src := builtinType(c.Typ)
	if src == nil {
		return 0, false
	}
	v := convertConst(ast.BuiltinInt, src, c.Value)
	if v == nil {
		return 0, false
	}
	return int64(v.(ast.Int)), true
}

func toFloat(c *ast.ConstValue) (float64, bool) {
	src := builtinType(c.Typ)
	if src == nil {
		return 0, false
	}
	v := convertConst(ast.BuiltinFloat64, src, c.Value)
	if v == nil {
		return 0, false
	}
	return float64(v.(ast.Float)), true
}

func minus(typ *ast.BuiltinType, val ast.Value) ast.Value {
	switch typ.Kind {
	case ast.BUILTIN_BOOL, ast.BUILTIN_UNTYPED_BOOL:
		return nil
	case ast.BUILTIN_UNTYPED_RUNE:
		return ast.Rune{Int: new(big.Int).Neg(val.(ast.Rune).Int)}
	case ast.BUILTIN_UNTYPED_INT:
		return ast.UntypedInt{Int: new(big.Int).Neg(val.(ast.UntypedInt).Int)}
	case ast.BUILTIN_UNTYPED_FLOAT:
		return ast.UntypedFloat{Float: new(big.Float).Neg(val.(ast.UntypedFloat).Float)}
	case ast.BUILTIN_FLOAT32, ast.BUILTIN_FLOAT64:
		return ast.Float(-float64(val.(ast.Float)))
	case ast.BUILTIN_UNTYPED_COMPLEX:
		v := val.(ast.UntypedComplex)
		return ast.UntypedComplex{
			Re: new(big.Float).Neg(v.Re),
			Im: new(big.Float).Neg(v.Im),
		}
	case ast.BUILTIN_COMPLEX64, ast.BUILTIN_COMPLEX128:
		v := val.(ast.Complex)
		return ast.Complex(complex(-real(v), -imag(v)))
	case ast.BUILTIN_STRING, ast.BUILTIN_UNTYPED_STRING:
		return nil
	default:
		if !typ.IsInteger() {
			panic("not reached")
		}
		if !typ.IsSigned() {
			return nil
		}
		return ast.Int(-int64(val.(ast.Int)))
	}
}

func complement(typ *ast.BuiltinType, val ast.Value) ast.Value {
	switch typ.Kind {
	case ast.BUILTIN_UNTYPED_INT:
		v := val.(ast.UntypedInt)
		return ast.UntypedInt{Int: new(big.Int).Not(v.Int)}
	case ast.BUILTIN_UNTYPED_RUNE:
		v := val.(ast.Rune)
		return ast.Rune{Int: new(big.Int).Not(v.Int)}
	default:
		if !typ.IsInteger() {
			return nil
		}
		v := val.(ast.Int)
		switch typ.Kind {
		case ast.BUILTIN_INT8:
			return ast.Int(^int8(v))
		case ast.BUILTIN_INT16:
			return ast.Int(^int16(v))
		case ast.BUILTIN_INT32:
			return ast.Int(^int32(v))
		case ast.BUILTIN_INT64, ast.BUILTIN_INT:
			return ast.Int(^int64(v))
		case ast.BUILTIN_UINT8:
			return ast.Int(^uint8(v))
		case ast.BUILTIN_UINT16:
			return ast.Int(^uint16(v))
		case ast.BUILTIN_UINT32:
			return ast.Int(^uint32(v))
		case ast.BUILTIN_UINT64, ast.BUILTIN_UINT, ast.BUILTIN_UINTPTR:
			return ast.Int(^uint64(v))
		default:
			panic("not reached")
		}
	}
}

func shift(
	x *ast.ConstValue, y *ast.ConstValue, op ast.Operation) (*ast.ConstValue, error) {

	// The shift count must be unsigned integer type. An untyped integer is
	// already converted to `uint64` by the expression verifier.
	if t := builtinType(y.Typ); t == nil || !t.IsInteger() || t.IsSigned() {
		return nil, &badShiftCountType{X: y}
	}
	// Get the shift count in a native uint.
	s := uint64(y.Value.(ast.Int))
	if s > MaxShift {
		return nil, &badShiftCountValue{X: s}
	}
	// Perform the shift.
	var res ast.Value
	switch t := builtinType(x.Typ); t.Kind {
	case ast.BUILTIN_UNTYPED_INT:
		v := x.Value.(ast.UntypedInt)
		if op == ast.SHL {
			res = ast.UntypedInt{Int: new(big.Int).Lsh(v.Int, uint(s))}
		} else {
			res = ast.UntypedInt{Int: new(big.Int).Rsh(v.Int, uint(s))}
		}
		return &ast.ConstValue{Off: x.Off, Typ: ast.BuiltinUntypedInt, Value: res}, nil
	case ast.BUILTIN_UNTYPED_FLOAT:
		v := x.Value.(ast.UntypedFloat)
		u, a := v.Int(new(big.Int))
		if a != big.Exact {
			return nil, &badOperandValue{Op: op, X: x}
		}
		if op == ast.SHL {
			res = ast.UntypedInt{Int: u.Lsh(u, uint(s))}
		} else {
			res = ast.UntypedInt{Int: u.Rsh(u, uint(s))}
		}
		return &ast.ConstValue{Off: x.Off, Typ: ast.BuiltinUntypedInt, Value: res}, nil
	case ast.BUILTIN_UNTYPED_COMPLEX:
		v := x.Value.(ast.UntypedComplex)
		if z, a := v.Im.Int64(); a != big.Exact || z != 0 {
			return nil, &badOperandValue{Op: op, X: x}
		}
		u, a := v.Re.Int(new(big.Int))
		if a != big.Exact {
			return nil, &badOperandValue{Op: op, X: x}
		}
		if op == ast.SHL {
			res = ast.UntypedInt{Int: u.Lsh(u, uint(s))}
		} else {
			res = ast.UntypedInt{Int: u.Rsh(u, uint(s))}
		}
		return &ast.ConstValue{Off: x.Off, Typ: ast.BuiltinUntypedInt, Value: res}, nil
	case ast.BUILTIN_UNTYPED_RUNE:
		v := x.Value.(ast.Rune)
		if op == ast.SHL {
			res = ast.Rune{Int: new(big.Int).Lsh(v.Int, uint(s))}
		} else {
			res = ast.Rune{Int: new(big.Int).Rsh(v.Int, uint(s))}
		}
		return &ast.ConstValue{Off: x.Off, Typ: ast.BuiltinUntypedRune, Value: res}, nil
	default:
		if !t.IsInteger() {
			return nil, &badOperandType{Op: op, Type: "integer type", X: x}
		}
		v := x.Value.(ast.Int)
		if op == ast.SHL {
			res = ast.Int(v << s)
		} else if builtinType(x.Typ).IsSigned() {
			res = ast.Int(int64(v) >> s)
		} else {
			res = ast.Int(v >> s)
		}
		return &ast.ConstValue{Off: x.Off, Typ: x.Typ, Value: res}, nil

	}
}

type badShiftCountType struct {
	X *ast.ConstValue
}

func (e *badShiftCountType) Error() string {
	return fmt.Sprintf("shift count must be unsigned integer (`%s` given)",
		e.X.TypeString())
}

type badShiftCountValue struct {
	X uint64
}

func (e *badShiftCountValue) Error() string {
	return fmt.Sprintf("shift count too big: %d", e.X)
}

type badOperandValue struct {
	Op ast.Operation
	X  *ast.ConstValue
}

func (e *badOperandValue) Error() string {
	return fmt.Sprintf("invalid operand to `%s`: `%s`", e.Op, e.X)
}

type badOperandType struct {
	Op   ast.Operation
	Type string
	X    *ast.ConstValue
}

func (e *badOperandType) Error() string {
	return fmt.Sprintf("invalid operand to `%s`: operand must have %s (`%s` given)",
		e.Op, e.Type, e.X.TypeString())
}

func untypedConvPanic(ast.Value) ast.Value { panic("not reached") }

func untypedConvIntToRune(x ast.Value) ast.Value {
	return ast.Rune{Int: new(big.Int).Set(x.(ast.UntypedInt).Int)}
}

func bigIntToFloat(x *big.Int) *big.Float {
	v := new(big.Float).SetPrec(ast.UNTYPED_FLOAT_PRECISION).SetMode(big.ToNearestEven)
	return v.SetInt(x)
}

func untypedConvIntToFloat(x ast.Value) ast.Value {
	return ast.UntypedFloat{Float: bigIntToFloat(x.(ast.UntypedInt).Int)}
}

func untypedConvIntToComplex(x ast.Value) ast.Value {
	return ast.UntypedComplex{
		Re: bigIntToFloat(x.(ast.UntypedInt).Int),
		Im: new(big.Float).
			SetPrec(ast.UNTYPED_FLOAT_PRECISION).
			SetMode(big.ToNearestEven),
	}
}

func untypedConvRuneToFloat(x ast.Value) ast.Value {
	return ast.UntypedFloat{Float: bigIntToFloat(x.(ast.Rune).Int)}
}

func untypedConvRuneToComplex(x ast.Value) ast.Value {
	return ast.UntypedComplex{
		Re: bigIntToFloat(x.(ast.Rune).Int),
		Im: new(big.Float).
			SetPrec(ast.UNTYPED_FLOAT_PRECISION).
			SetMode(big.ToNearestEven),
	}
}

func untypedConvFloatToComplex(x ast.Value) ast.Value {
	return ast.UntypedComplex{
		Re: new(big.Float).Set(x.(ast.UntypedFloat).Float),
		Im: new(big.Float).
			SetPrec(ast.UNTYPED_FLOAT_PRECISION).
			SetMode(big.ToNearestEven),
	}
}

type untypedKind uint

const (
	_UNTYPED_INT untypedKind = iota
	_UNTYPED_RUNE
	_UNTYPED_FLOAT
	_UNTYPED_COMPLEX
	_UNTYPED_BOOL
	_UNTYPED_STRING
)

func untypedToKind(x ast.Value) untypedKind {
	switch x.(type) {
	case ast.UntypedInt:
		return _UNTYPED_INT
	case ast.Rune:
		return _UNTYPED_RUNE
	case ast.UntypedFloat:
		return _UNTYPED_FLOAT
	case ast.UntypedComplex:
		return _UNTYPED_COMPLEX
	case ast.Bool:
		return _UNTYPED_BOOL
	case ast.String:
		return _UNTYPED_STRING
	default:
		panic("not reached")
	}
}

var untypedConvTable = [...][4]func(ast.Value) ast.Value{
	{
		untypedConvPanic,
		untypedConvIntToRune, untypedConvIntToFloat, untypedConvIntToComplex,
	},
	{
		untypedConvPanic, untypedConvPanic,
		untypedConvRuneToFloat, untypedConvRuneToComplex,
	},
	{
		untypedConvPanic, untypedConvPanic, untypedConvPanic,
		untypedConvFloatToComplex,
	},
}

func untypedConvert(x ast.Value, y ast.Value) (untypedKind, ast.Value, ast.Value) {
	i, j := untypedToKind(x), untypedToKind(y)
	if i == j {
		return i, x, y
	}
	swap := false
	if i > j {
		swap = true
		i, j = j, i
		x, y = y, x
	}
	x = untypedConvTable[i][j](x)
	if swap {
		x, y = y, x
	}
	return j, x, y
}

func compare(x *ast.ConstValue, y *ast.ConstValue, op ast.Operation) (ast.Value, error) {
	// If the constants are typed, they must have the same type.
	if t := builtinType(x.Typ); t != nil && !t.IsUntyped() {
		if t != builtinType(y.Typ) {
			return nil, &mismatchedTypes{Op: op, X: x, Y: y}
		}
		return compareTyped(t, x.Value, y.Value, op)
	} else {
		// If the operands are untyped, they must either both be bool,
		// or strings, or use the type later in the sequence int, rune, float,
		// complex, by converting one of the operands to the type of the
		// other.
		var (
			t    untypedKind
			u, v ast.Value
		)
		switch x.Value.(type) {
		case ast.Bool:
			if _, ok := y.Value.(ast.Bool); !ok {
				return nil, &mismatchedTypes{Op: op, X: x, Y: y}
			}
			t, u, v = _UNTYPED_BOOL, x.Value, y.Value
		case ast.String:
			if _, ok := y.Value.(ast.String); !ok {
				return nil, &mismatchedTypes{Op: op, X: x, Y: y}
			}
			t, u, v = _UNTYPED_STRING, x.Value, y.Value
		default:
			t, u, v = untypedConvert(x.Value, y.Value)
		}
		return compareUntyped(t, u, v, op)
	}
}

func compareTyped(
	typ *ast.BuiltinType, x ast.Value, y ast.Value, op ast.Operation) (ast.Value, error) {

	c := 0
	switch typ.Kind {
	case ast.BUILTIN_BOOL:
		u, v := x.(ast.Bool), y.(ast.Bool)
		switch op {
		case ast.EQ:
			return ast.Bool(u == v), nil
		case ast.NE:
			return ast.Bool(u != v), nil
		default:
			return nil, &invalidOperation{Op: op, Type: typ}
		}
	case ast.BUILTIN_UINT8, ast.BUILTIN_UINT16, ast.BUILTIN_UINT32,
		ast.BUILTIN_UINT64, ast.BUILTIN_UINT, ast.BUILTIN_UINTPTR:
		u, v := x.(ast.Int), y.(ast.Int)
		if u < v {
			c = -1
		} else if u > v {
			c = 1
		}
	case ast.BUILTIN_INT8, ast.BUILTIN_INT16, ast.BUILTIN_INT32,
		ast.BUILTIN_INT64, ast.BUILTIN_INT:
		u, v := int64(x.(ast.Int)), int64(y.(ast.Int))
		if u < v {
			c = -1
		} else if u > v {
			c = 1
		}
	case ast.BUILTIN_FLOAT32, ast.BUILTIN_FLOAT64:
		u, v := x.(ast.Float), y.(ast.Float)
		if u < v {
			c = -1
		} else if u > v {
			c = 1
		}
	case ast.BUILTIN_COMPLEX64, ast.BUILTIN_COMPLEX128:
		u, v := x.(ast.Complex), y.(ast.Complex)
		switch op {
		case ast.EQ:
			return ast.Bool(u == v), nil
		case ast.NE:
			return ast.Bool(u != v), nil
		default:
			return nil, &invalidOperation{Op: op, Type: typ}
		}
	case ast.BUILTIN_STRING:
		u, v := x.(ast.String), y.(ast.String)
		if u < v {
			c = -1
		} else if u > v {
			c = 1
		}
	default:
		panic("not reached")
	}
	return cmpToBool(c, op), nil
}

func compareUntyped(
	t untypedKind, x ast.Value, y ast.Value, op ast.Operation) (ast.Value, error) {

	c := 0
	switch t {
	case _UNTYPED_INT:
		x, y := x.(ast.UntypedInt), y.(ast.UntypedInt)
		c = x.Cmp(y.Int)
	case _UNTYPED_RUNE:
		x, y := x.(ast.Rune), y.(ast.Rune)
		c = x.Cmp(y.Int)
	case _UNTYPED_FLOAT:
		x, y := x.(ast.UntypedFloat), y.(ast.UntypedFloat)
		c = x.Cmp(y.Float)
	case _UNTYPED_COMPLEX:
		if op != ast.EQ && op != ast.NE {
			return nil, &invalidOperation{Op: op, Type: ast.BuiltinUntypedComplex}
		}
		x, y := x.(ast.UntypedComplex), y.(ast.UntypedComplex)
		u, v := x.Re.Cmp(y.Re), x.Im.Cmp(y.Im)
		if op == ast.EQ {
			return ast.Bool(u == 0 && v == 0), nil
		} else {
			return ast.Bool(u != 0 || v != 0), nil
		}
	case _UNTYPED_BOOL:
		if op != ast.EQ && op != ast.NE {
			return nil, &invalidOperation{Op: op, Type: ast.BuiltinUntypedBool}
		}
		x, y := x.(ast.Bool), y.(ast.Bool)
		if op == ast.EQ {
			return ast.Bool(x == y), nil
		} else {
			return ast.Bool(x != y), nil
		}
	case _UNTYPED_STRING:
		x, y := x.(ast.String), y.(ast.String)
		if x < y {
			c = -1
		} else if x > y {
			c = 1
		}
	default:
		panic("not reached")
	}
	return cmpToBool(c, op), nil
}

func cmpToBool(c int, op ast.Operation) ast.Bool {
	switch op {
	case ast.LT:
		return ast.Bool(c == -1)
	case ast.GT:
		return ast.Bool(c == 1)
	case ast.EQ:
		return ast.Bool(c == 0)
	case ast.NE:
		return ast.Bool(c != 0)
	case ast.LE:
		return ast.Bool(c <= 0)
	case ast.GE:
		return ast.Bool(c >= 0)
	default:
		panic("not reached")
	}
}

type invalidOperation struct {
	Op   ast.Operation
	Type *ast.BuiltinType
}

func (e *invalidOperation) Error() string {
	return fmt.Sprintf("operation `%s` not supported for `%s`", e.Op, e.Type)
}

type mismatchedTypes struct {
	Op   ast.Operation
	X, Y *ast.ConstValue
}

func (e *mismatchedTypes) Error() string {
	return fmt.Sprintf("invalid operation `%s`: mismatched types `%s` and `%s`",
		e.Op, e.X.TypeString(), e.Y.TypeString())
}

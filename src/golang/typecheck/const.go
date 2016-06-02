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
	ZeroInt   = big.Int{}
	ZeroFloat = big.Float{}
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

func convert(dst, src *ast.BuiltinType, val ast.Value) ast.Value {
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

func Convert(dst ast.Type, cst *ast.ConstValue) (*ast.ConstValue, error) {
	if d := builtinType(dst); d != nil {
		if v := convert(d, builtinType(cst.Typ), cst.Value); v != nil {
			return &ast.ConstValue{Typ: dst, Value: v}, nil
		}
	}
	return nil, &badConstConversion{Dst: dst, Src: cst}
}

func ToInt(c *ast.ConstValue) (int64, error) {
	v := convert(ast.BuiltinInt, builtinType(c.Typ), c.Value)
	if v == nil {
		return 0, &badConstConversion{Dst: ast.BuiltinInt, Src: c}
	}
	return int64(v.(ast.Int)), nil
}

func ToUint(c *ast.ConstValue) (uint64, error) {
	v := convert(ast.BuiltinUint, builtinType(c.Typ), c.Value)
	if v == nil {
		return 0, &badConstConversion{Dst: ast.BuiltinUint, Src: c}
	}
	return uint64(v.(ast.Int)), nil
}

func ToInt64(c *ast.ConstValue) (int64, error) {
	v := convert(ast.BuiltinInt64, builtinType(c.Typ), c.Value)
	if v == nil {
		return 0, &badConstConversion{Dst: ast.BuiltinInt64, Src: c}
	}
	return int64(v.(ast.Int)), nil
}

func ToUint64(c *ast.ConstValue) (uint64, error) {
	v := convert(ast.BuiltinUint64, builtinType(c.Typ), c.Value)
	if v == nil {
		return 0, &badConstConversion{Dst: ast.BuiltinUint64, Src: c}
	}
	return uint64(v.(ast.Int)), nil
}

func ToFloat(c *ast.ConstValue) (float64, error) {
	v := convert(ast.BuiltinFloat64, builtinType(c.Typ), c.Value)
	if v == nil {
		return 0, &badConstConversion{Dst: ast.BuiltinFloat64, Src: c}
	}
	return float64(v.(ast.Float)), nil
}

func ToComplex(c *ast.ConstValue) (complex128, error) {
	v := convert(ast.BuiltinComplex128, builtinType(c.Typ), c.Value)
	if v == nil {
		return 0, &badConstConversion{Dst: ast.BuiltinComplex128, Src: c}
	}
	return complex128(v.(ast.Complex)), nil
}

func ToString(c *ast.ConstValue) (string, error) {
	v := convert(ast.BuiltinString, builtinType(c.Typ), c.Value)
	if v == nil {
		return "", &badConstConversion{Dst: ast.BuiltinString, Src: c}
	}
	return string(v.(ast.String)), nil
}

func Minus(cst *ast.ConstValue) (*ast.ConstValue, error) {
	typ, val := builtinType(cst.Typ), cst.Value
	var res ast.Value
	switch typ.Kind {
	case ast.BUILTIN_BOOL, ast.BUILTIN_UNTYPED_BOOL:
		res = nil
	case ast.BUILTIN_UNTYPED_RUNE:
		res = ast.Rune{Int: new(big.Int).Neg(val.(ast.Rune).Int)}
	case ast.BUILTIN_UNTYPED_INT:
		res = ast.UntypedInt{Int: new(big.Int).Neg(val.(ast.UntypedInt).Int)}
	case ast.BUILTIN_UNTYPED_FLOAT:
		res = ast.UntypedFloat{Float: new(big.Float).Neg(val.(ast.UntypedFloat).Float)}
	case ast.BUILTIN_FLOAT32, ast.BUILTIN_FLOAT64:
		res = ast.Float(-float64(val.(ast.Float)))
	case ast.BUILTIN_UNTYPED_COMPLEX:
		v := val.(ast.UntypedComplex)
		res = ast.UntypedComplex{
			Re: new(big.Float).Neg(v.Re),
			Im: new(big.Float).Neg(v.Im),
		}
	case ast.BUILTIN_COMPLEX64, ast.BUILTIN_COMPLEX128:
		v := val.(ast.Complex)
		res = ast.Complex(complex(-real(v), -imag(v)))
	case ast.BUILTIN_STRING, ast.BUILTIN_UNTYPED_STRING:
		res = nil
	default:
		if !typ.IsInteger() {
			panic("not reached")
		}
		if !typ.IsSigned() {
			res = nil
		} else {
			res = ast.Int(-int64(val.(ast.Int)))
		}
	}
	if res == nil {
		return nil, &badOperandType{Op: '-', Type: "arithmetic type", X: cst}
	}
	return &ast.ConstValue{Typ: cst.Typ, Value: res}, nil
}

func Complement(cst *ast.ConstValue) (*ast.ConstValue, error) {
	typ, val := builtinType(cst.Typ), cst.Value
	var res ast.Value
	switch typ.Kind {
	case ast.BUILTIN_UNTYPED_INT:
		v := val.(ast.UntypedInt)
		res = ast.UntypedInt{Int: new(big.Int).Not(v.Int)}
	case ast.BUILTIN_UNTYPED_RUNE:
		v := val.(ast.Rune)
		res = ast.Rune{Int: new(big.Int).Not(v.Int)}
	default:
		if !typ.IsInteger() {
			res = nil
		} else {
			v := val.(ast.Int)
			switch typ.Kind {
			case ast.BUILTIN_INT8:
				res = ast.Int(^int8(v))
			case ast.BUILTIN_INT16:
				res = ast.Int(^int16(v))
			case ast.BUILTIN_INT32:
				res = ast.Int(^int32(v))
			case ast.BUILTIN_INT64, ast.BUILTIN_INT:
				res = ast.Int(^int64(v))
			case ast.BUILTIN_UINT8:
				res = ast.Int(^uint8(v))
			case ast.BUILTIN_UINT16:
				res = ast.Int(^uint16(v))
			case ast.BUILTIN_UINT32:
				res = ast.Int(^uint32(v))
			case ast.BUILTIN_UINT64, ast.BUILTIN_UINT, ast.BUILTIN_UINTPTR:
				res = ast.Int(^uint64(v))
			default:
				panic("not reached")
			}
		}
	}
	if res == nil {
		return nil, &badOperandType{Op: '-', Type: "integral type", X: cst}
	}
	return &ast.ConstValue{Typ: cst.Typ, Value: res}, nil
}

func Not(cst *ast.ConstValue) (*ast.ConstValue, error) {
	typ, val := builtinType(cst.Typ), cst.Value
	if typ.Kind != ast.BUILTIN_BOOL && typ.Kind != ast.BUILTIN_UNTYPED_BOOL {
		return nil, &badOperandType{Op: '!', Type: "boolean type", X: cst}
	}
	return &ast.ConstValue{Typ: cst.Typ, Value: ast.Bool(!val.(ast.Bool))}, nil
}

func Shift(
	x *ast.ConstValue, y *ast.ConstValue, op ast.Operation) (*ast.ConstValue, error) {

	// Get the shift count in a native uint.
	v := convert(ast.BuiltinUint64, builtinType(y.Typ), y.Value)
	if v == nil {
		return nil, &badShiftCount{X: y}
	}
	s := uint64(v.(ast.Int))
	if s > MaxShift {
		return nil, &bigShiftCount{X: s}
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

type untypedKind int

const (
	_UNTYPED_ERR untypedKind = iota - 1
	_UNTYPED_INT
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
	if i > _UNTYPED_COMPLEX || j > _UNTYPED_COMPLEX {
		return _UNTYPED_ERR, nil, nil
	}
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

func Compare(
	x *ast.ConstValue, y *ast.ConstValue, op ast.Operation) (*ast.ConstValue, error) {

	var res ast.Value
	var err error
	// If the constants are typed, they must have the same type.
	if t := builtinType(x.Typ); t != nil && !t.IsUntyped() {
		if x.Typ != y.Typ {
			return nil, &mismatchedTypes{Op: op, X: x, Y: y}
		}
		res, err = compareTyped(t, x.Value, y.Value, op)
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
			if t == _UNTYPED_ERR {
				return nil, &mismatchedTypes{Op: op, X: x, Y: y}
			}
		}
		res, err = compareUntyped(t, u, v, op)
	}
	if err != nil {
		return nil, err
	}
	return &ast.ConstValue{Typ: ast.BuiltinUntypedBool, Value: res}, nil
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

var untypedType = [...]*ast.BuiltinType{
	ast.BuiltinUntypedInt,
	ast.BuiltinUntypedRune,
	ast.BuiltinUntypedFloat,
	ast.BuiltinUntypedComplex,
	ast.BuiltinUntypedBool,
	ast.BuiltinUntypedString,
}

func Add(x *ast.ConstValue, y *ast.ConstValue) (*ast.ConstValue, error) {
	// If the constants are typed, they must have the same type.
	if t := builtinType(x.Typ); t != nil && !t.IsUntyped() {
		if x.Typ != y.Typ {
			return nil, &mismatchedTypes{Op: '+', X: x, Y: y}
		}
		if t.Kind == ast.BUILTIN_BOOL {
			return nil, &invalidOperation{Op: '+', Type: t}
		}
		return &ast.ConstValue{Typ: x.Typ, Value: addTyped(t, x.Value, y.Value)}, nil
	}

	// If the operands are untyped, they must either both strings, or use the
	// type later in the sequence int, rune, float, complex, by converting one
	// of the operands to the type of the other.
	var t untypedKind
	var u, v ast.Value
	switch x.Value.(type) {
	case ast.Bool:
		return nil, &invalidOperation{Op: '+', Type: ast.BuiltinUntypedBool}
	case ast.String:
		if _, ok := y.Value.(ast.String); !ok {
			return nil, &mismatchedTypes{Op: '+', X: x, Y: y}
		}
		t, u, v = _UNTYPED_STRING, x.Value, y.Value
	default:
		t, u, v = untypedConvert(x.Value, y.Value)
		if t == _UNTYPED_ERR {
			return nil, &mismatchedTypes{Op: '+', X: x, Y: y}
		}
	}
	return &ast.ConstValue{Typ: untypedType[t], Value: addUntyped(t, u, v)}, nil
}

func addTyped(typ *ast.BuiltinType, x ast.Value, y ast.Value) ast.Value {
	switch typ.Kind {
	case ast.BUILTIN_UINT8, ast.BUILTIN_UINT16, ast.BUILTIN_UINT32,
		ast.BUILTIN_UINT64, ast.BUILTIN_UINT, ast.BUILTIN_UINTPTR:
		u, v := x.(ast.Int), y.(ast.Int)
		return ast.Int(u + v)
	case ast.BUILTIN_INT8, ast.BUILTIN_INT16, ast.BUILTIN_INT32,
		ast.BUILTIN_INT64, ast.BUILTIN_INT:
		u, v := int64(x.(ast.Int)), int64(y.(ast.Int))
		return ast.Int(u + v)
	case ast.BUILTIN_FLOAT32:
		u, v := float32(x.(ast.Float)), float32(y.(ast.Float))
		return ast.Float(u + v)
	case ast.BUILTIN_FLOAT64:
		u, v := x.(ast.Float), y.(ast.Float)
		return ast.Float(u + v)
	case ast.BUILTIN_COMPLEX64:
		u, v := complex64(x.(ast.Complex)), complex64(y.(ast.Complex))
		return ast.Complex(u + v)
	case ast.BUILTIN_COMPLEX128:
		u, v := x.(ast.Complex), y.(ast.Complex)
		return ast.Complex(u + v)
	case ast.BUILTIN_STRING:
		u, v := x.(ast.String), y.(ast.String)
		return ast.String(u + v)
	default:
		panic("not reached")
	}
}

func addUntyped(t untypedKind, x ast.Value, y ast.Value) ast.Value {
	switch t {
	case _UNTYPED_INT:
		x, y := x.(ast.UntypedInt), y.(ast.UntypedInt)
		return ast.UntypedInt{Int: new(big.Int).Add(x.Int, y.Int)}
	case _UNTYPED_RUNE:
		x, y := x.(ast.Rune), y.(ast.Rune)
		return ast.Rune{Int: new(big.Int).Add(x.Int, y.Int)}
	case _UNTYPED_FLOAT:
		x, y := x.(ast.UntypedFloat), y.(ast.UntypedFloat)
		return ast.UntypedFloat{Float: new(big.Float).Add(x.Float, y.Float)}
	case _UNTYPED_COMPLEX:
		x, y := x.(ast.UntypedComplex), y.(ast.UntypedComplex)
		return ast.UntypedComplex{
			Re: new(big.Float).Add(x.Re, y.Re),
			Im: new(big.Float).Add(x.Im, y.Im),
		}
	case _UNTYPED_STRING:
		x, y := x.(ast.String), y.(ast.String)
		return ast.String(x + y)
	default:
		panic("not reached")
	}
}

func Sub(x *ast.ConstValue, y *ast.ConstValue) (*ast.ConstValue, error) {
	// If the constants are typed, they must have the same type.
	if t := builtinType(x.Typ); t != nil && !t.IsUntyped() {
		if x.Typ != y.Typ {
			return nil, &mismatchedTypes{Op: '-', X: x, Y: y}
		}
		if t.Kind == ast.BUILTIN_BOOL || t.Kind == ast.BUILTIN_STRING {
			return nil, &invalidOperation{Op: '-', Type: t}
		}
		return &ast.ConstValue{Typ: x.Typ, Value: subTyped(t, x.Value, y.Value)}, nil
	}

	// If the operands are untyped, use the type later in the sequence int,
	// rune, float, complex, by converting one of the operands to the type of
	// the other.
	var t untypedKind
	var u, v ast.Value
	switch x.Value.(type) {
	case ast.Bool:
		return nil, &invalidOperation{Op: '-', Type: ast.BuiltinUntypedBool}
	case ast.String:
		return nil, &invalidOperation{Op: '-', Type: ast.BuiltinUntypedString}
	default:
		t, u, v = untypedConvert(x.Value, y.Value)
		if t == _UNTYPED_ERR {
			return nil, &mismatchedTypes{Op: '-', X: x, Y: y}
		}
	}
	return &ast.ConstValue{Typ: untypedType[t], Value: subUntyped(t, u, v)}, nil
}

func subTyped(typ *ast.BuiltinType, x ast.Value, y ast.Value) ast.Value {
	switch typ.Kind {
	case ast.BUILTIN_UINT8, ast.BUILTIN_UINT16, ast.BUILTIN_UINT32,
		ast.BUILTIN_UINT64, ast.BUILTIN_UINT, ast.BUILTIN_UINTPTR:
		u, v := x.(ast.Int), y.(ast.Int)
		return ast.Int(u - v)
	case ast.BUILTIN_INT8, ast.BUILTIN_INT16, ast.BUILTIN_INT32,
		ast.BUILTIN_INT64, ast.BUILTIN_INT:
		u, v := int64(x.(ast.Int)), int64(y.(ast.Int))
		return ast.Int(u - v)
	case ast.BUILTIN_FLOAT32:
		u, v := float32(x.(ast.Float)), float32(y.(ast.Float))
		return ast.Float(u - v)
	case ast.BUILTIN_FLOAT64:
		u, v := x.(ast.Float), y.(ast.Float)
		return ast.Float(u - v)
	case ast.BUILTIN_COMPLEX64:
		u, v := complex64(x.(ast.Complex)), complex64(y.(ast.Complex))
		return ast.Complex(u - v)
	case ast.BUILTIN_COMPLEX128:
		u, v := x.(ast.Complex), y.(ast.Complex)
		return ast.Complex(u - v)
	default:
		panic("not reached")
	}
}

func subUntyped(t untypedKind, x ast.Value, y ast.Value) ast.Value {
	switch t {
	case _UNTYPED_INT:
		x, y := x.(ast.UntypedInt), y.(ast.UntypedInt)
		return ast.UntypedInt{Int: new(big.Int).Sub(x.Int, y.Int)}
	case _UNTYPED_RUNE:
		x, y := x.(ast.Rune), y.(ast.Rune)
		return ast.Rune{Int: new(big.Int).Sub(x.Int, y.Int)}
	case _UNTYPED_FLOAT:
		x, y := x.(ast.UntypedFloat), y.(ast.UntypedFloat)
		return ast.UntypedFloat{Float: new(big.Float).Sub(x.Float, y.Float)}
	case _UNTYPED_COMPLEX:
		x, y := x.(ast.UntypedComplex), y.(ast.UntypedComplex)
		return ast.UntypedComplex{
			Re: new(big.Float).Sub(x.Re, y.Re),
			Im: new(big.Float).Sub(x.Im, y.Im),
		}
	default:
		panic("not reached")
	}
}

func Mul(x *ast.ConstValue, y *ast.ConstValue) (*ast.ConstValue, error) {
	// If the constants are typed, they must have the same type.
	if t := builtinType(x.Typ); t != nil && !t.IsUntyped() {
		if x.Typ != y.Typ {
			return nil, &mismatchedTypes{Op: '*', X: x, Y: y}
		}
		if t.Kind == ast.BUILTIN_BOOL || t.Kind == ast.BUILTIN_STRING {
			return nil, &invalidOperation{Op: '*', Type: t}
		}
		return &ast.ConstValue{Typ: x.Typ, Value: mulTyped(t, x.Value, y.Value)}, nil
	}

	// If the operands are untyped, use the type later in the sequence int,
	// rune, float, complex, by converting one of the operands to the type of
	// the other.
	var t untypedKind
	var u, v ast.Value
	switch x.Value.(type) {
	case ast.Bool:
		return nil, &invalidOperation{Op: '*', Type: ast.BuiltinUntypedBool}
	case ast.String:
		return nil, &invalidOperation{Op: '*', Type: ast.BuiltinUntypedString}
	default:
		t, u, v = untypedConvert(x.Value, y.Value)
		if t == _UNTYPED_ERR {
			return nil, &mismatchedTypes{Op: '*', X: x, Y: y}
		}
	}
	return &ast.ConstValue{Typ: untypedType[t], Value: mulUntyped(t, u, v)}, nil
}

func mulTyped(typ *ast.BuiltinType, x ast.Value, y ast.Value) ast.Value {
	switch typ.Kind {
	case ast.BUILTIN_UINT8, ast.BUILTIN_UINT16, ast.BUILTIN_UINT32,
		ast.BUILTIN_UINT64, ast.BUILTIN_UINT, ast.BUILTIN_UINTPTR:
		u, v := x.(ast.Int), y.(ast.Int)
		return ast.Int(u * v)
	case ast.BUILTIN_INT8, ast.BUILTIN_INT16, ast.BUILTIN_INT32,
		ast.BUILTIN_INT64, ast.BUILTIN_INT:
		u, v := int64(x.(ast.Int)), int64(y.(ast.Int))
		return ast.Int(u * v)
	case ast.BUILTIN_FLOAT32:
		u, v := float32(x.(ast.Float)), float32(y.(ast.Float))
		return ast.Float(u * v)
	case ast.BUILTIN_FLOAT64:
		u, v := x.(ast.Float), y.(ast.Float)
		return ast.Float(u * v)
	case ast.BUILTIN_COMPLEX64:
		u, v := complex64(x.(ast.Complex)), complex64(y.(ast.Complex))
		return ast.Complex(u * v)
	case ast.BUILTIN_COMPLEX128:
		u, v := x.(ast.Complex), y.(ast.Complex)
		return ast.Complex(u * v)
	default:
		panic("not reached")
	}
}

func mulUntyped(t untypedKind, x ast.Value, y ast.Value) ast.Value {
	switch t {
	case _UNTYPED_INT:
		x, y := x.(ast.UntypedInt), y.(ast.UntypedInt)
		return ast.UntypedInt{Int: new(big.Int).Mul(x.Int, y.Int)}
	case _UNTYPED_RUNE:
		x, y := x.(ast.Rune), y.(ast.Rune)
		return ast.Rune{Int: new(big.Int).Mul(x.Int, y.Int)}
	case _UNTYPED_FLOAT:
		x, y := x.(ast.UntypedFloat), y.(ast.UntypedFloat)
		return ast.UntypedFloat{Float: new(big.Float).Mul(x.Float, y.Float)}
	case _UNTYPED_COMPLEX:
		x, y := x.(ast.UntypedComplex), y.(ast.UntypedComplex)
		t0 := new(big.Float).Mul(x.Re, y.Re)
		t1 := new(big.Float).Mul(x.Im, y.Im)
		re := new(big.Float).Sub(t0, t1)
		t0.Mul(x.Re, y.Im)
		t1.Mul(y.Re, x.Im)
		im := new(big.Float).Add(t0, t1)
		return ast.UntypedComplex{Re: re, Im: im}
	default:
		panic("not reached")
	}
}

func Div(x *ast.ConstValue, y *ast.ConstValue) (*ast.ConstValue, error) {
	// If the constants are typed, they must have the same type.
	if t := builtinType(x.Typ); t != nil && !t.IsUntyped() {
		if x.Typ != y.Typ {
			return nil, &mismatchedTypes{Op: '/', X: x, Y: y}
		}
		if t.Kind == ast.BUILTIN_BOOL || t.Kind == ast.BUILTIN_STRING {
			return nil, &invalidOperation{Op: '/', Type: t}
		}
		v, err := divTyped(t, x.Value, y.Value)
		if err != nil {
			return nil, err
		}
		return &ast.ConstValue{Typ: x.Typ, Value: v}, nil
	}

	// If the operands are untyped, use the type later in the sequence int,
	// rune, float, complex, by converting one of the operands to the type of
	// the other.
	var t untypedKind
	var u, v ast.Value
	switch x.Value.(type) {
	case ast.Bool:
		return nil, &invalidOperation{Op: '/', Type: ast.BuiltinUntypedBool}
	case ast.String:
		return nil, &invalidOperation{Op: '/', Type: ast.BuiltinUntypedString}
	default:
		t, u, v = untypedConvert(x.Value, y.Value)
		if t == _UNTYPED_ERR {
			return nil, &mismatchedTypes{Op: '/', X: x, Y: y}
		}
	}
	v, err := divUntyped(t, u, v)
	if err != nil {
		return nil, err
	}
	return &ast.ConstValue{Typ: untypedType[t], Value: v}, nil
}

func divTyped(typ *ast.BuiltinType, x ast.Value, y ast.Value) (ast.Value, error) {
	if typ.IsInteger() {
		u, v := x.(ast.Int), y.(ast.Int)
		if v == 0 {
			return nil, &divisionByZero{}
		}
		if typ.IsSigned() {
			return ast.Int(int64(u) / int64(v)), nil
		} else {
			return ast.Int(u / v), nil
		}
	}
	switch typ.Kind {
	case ast.BUILTIN_FLOAT32:
		u, v := float32(x.(ast.Float)), float32(y.(ast.Float))
		if v == 0 {
			return nil, &divisionByZero{}
		}
		return ast.Float(u / v), nil
	case ast.BUILTIN_FLOAT64:
		u, v := x.(ast.Float), y.(ast.Float)
		if v == 0 {
			return nil, &divisionByZero{}
		}
		return ast.Float(u / v), nil
	case ast.BUILTIN_COMPLEX64:
		u, v := complex64(x.(ast.Complex)), complex64(y.(ast.Complex))
		a, b, c, d := real(u), imag(u), real(v), imag(v)
		z := c*c + d*d
		if z == 0 {
			return nil, &divisionByZero{}
		}
		re := float64((a*c + b*d) / z)
		im := float64((b*c - a*d) / z)
		return ast.Complex(complex(re, im)), nil
	case ast.BUILTIN_COMPLEX128:
		u, v := x.(ast.Complex), y.(ast.Complex)
		a, b, c, d := real(u), imag(u), real(v), imag(v)
		z := c*c + d*d
		if z == 0 {
			return nil, &divisionByZero{}
		}
		re := (a*c + b*d) / z
		im := (b*c - a*d) / z
		return ast.Complex(complex(re, im)), nil
	default:
		panic("not reached")
	}
}

func divUntyped(t untypedKind, x ast.Value, y ast.Value) (ast.Value, error) {
	switch t {
	case _UNTYPED_INT:
		x, y := x.(ast.UntypedInt), y.(ast.UntypedInt)
		if y.Int.Cmp(&ZeroInt) == 0 {
			return nil, &divisionByZero{}
		}
		return ast.UntypedInt{Int: new(big.Int).Div(x.Int, y.Int)}, nil
	case _UNTYPED_RUNE:
		x, y := x.(ast.Rune), y.(ast.Rune)
		if y.Int.Cmp(&ZeroInt) == 0 {
			return nil, &divisionByZero{}
		}
		return ast.Rune{Int: new(big.Int).Div(x.Int, y.Int)}, nil
	case _UNTYPED_FLOAT:
		x, y := x.(ast.UntypedFloat), y.(ast.UntypedFloat)
		if y.Float.Cmp(&ZeroFloat) == 0 {
			return nil, &divisionByZero{}
		}
		return ast.UntypedFloat{Float: new(big.Float).Quo(x.Float, y.Float)}, nil
	case _UNTYPED_COMPLEX:
		x, y := x.(ast.UntypedComplex), y.(ast.UntypedComplex)
		a, b, c, d := x.Re, x.Im, y.Re, y.Im
		t0 := new(big.Float).Mul(c, c)
		t1 := new(big.Float).Mul(d, d)
		z := new(big.Float).Add(t0, t1)
		if z.Cmp(&ZeroFloat) == 0 {
			return nil, &divisionByZero{}
		}
		t0.Mul(a, c)
		t1.Mul(b, d)
		re := new(big.Float).Add(t0, t1)
		re.Quo(re, z)
		t0.Mul(b, c)
		t1.Mul(a, d)
		im := t0.Sub(t0, t1).Quo(t0, z)
		return ast.UntypedComplex{Re: re, Im: im}, nil
	default:
		panic("not reached")
	}
}

func Rem(x *ast.ConstValue, y *ast.ConstValue) (*ast.ConstValue, error) {
	// If the constants are typed, they must have the same type.
	tx := builtinType(x.Typ)
	if !tx.IsUntyped() {
		if x.Typ != y.Typ {
			return nil, &mismatchedTypes{Op: '%', X: x, Y: y}
		}
		// Only integer operands.
		if !tx.IsInteger() {
			return nil, &invalidOperation{Op: '%', Type: tx}
		}
		u, v := x.Value.(ast.Int), y.Value.(ast.Int)
		if v == 0 {
			return nil, &divisionByZero{}
		}
		if tx.IsSigned() {
			return &ast.ConstValue{Typ: x.Typ, Value: ast.Int(int64(u) % int64(v))}, nil
		} else {
			return &ast.ConstValue{Typ: x.Typ, Value: ast.Int(u % v)}, nil
		}
	}

	// Only untyped integers and runes allowed.
	var u, v *big.Int
	switch {
	case tx == ast.BuiltinUntypedInt:
		u = x.Value.(ast.UntypedInt).Int
	case tx == ast.BuiltinUntypedRune:
		u = x.Value.(ast.Rune).Int
	default:
		return nil, &invalidOperation{Op: '%', Type: tx}
	}
	ty := builtinType(y.Typ)
	switch {
	case ty == ast.BuiltinUntypedInt:
		v = y.Value.(ast.UntypedInt).Int
	case ty == ast.BuiltinUntypedRune:
		v = y.Value.(ast.Rune).Int
		tx = ast.BuiltinUntypedRune
	default:
		return nil, &invalidOperation{Op: '%', Type: ty}
	}
	if v.Cmp(&ZeroInt) == 0 {
		return nil, &divisionByZero{}
	}
	var res ast.Value
	r := new(big.Int).Rem(u, v)
	if tx == ast.BuiltinUntypedRune {
		res = ast.Rune{Int: r}
	} else {
		res = ast.UntypedInt{Int: r}
	}
	return &ast.ConstValue{Typ: tx, Value: res}, nil
}

func Bit(op ast.Operation, x *ast.ConstValue, y *ast.ConstValue) (*ast.ConstValue, error) {
	// If the constants are typed, they must have the same type.
	tx := builtinType(x.Typ)
	if !tx.IsUntyped() {
		if x.Typ != y.Typ {
			return nil, &mismatchedTypes{Op: op, X: x, Y: y}
		}
		// Only integer operands.
		if !tx.IsInteger() {
			return nil, &invalidOperation{Op: op, Type: tx}
		}
		v := bitTyped(op, x.Value.(ast.Int), y.Value.(ast.Int))
		return &ast.ConstValue{Typ: x.Typ, Value: v}, nil
	}

	// Only untyped integers and runes allowed.
	var u, v *big.Int
	switch {
	case tx == ast.BuiltinUntypedInt:
		u = x.Value.(ast.UntypedInt).Int
	case tx == ast.BuiltinUntypedRune:
		u = x.Value.(ast.Rune).Int
	default:
		return nil, &invalidOperation{Op: op, Type: tx}
	}
	ty := builtinType(y.Typ)
	switch {
	case ty == ast.BuiltinUntypedInt:
		v = y.Value.(ast.UntypedInt).Int
	case ty == ast.BuiltinUntypedRune:
		v = y.Value.(ast.Rune).Int
		tx = ast.BuiltinUntypedRune
	default:
		return nil, &invalidOperation{Op: op, Type: ty}
	}
	r := bitUntyped(op, u, v)
	var res ast.Value
	if tx == ast.BuiltinUntypedRune {
		res = ast.Rune{Int: r}
	} else {
		res = ast.UntypedInt{Int: r}
	}
	return &ast.ConstValue{Typ: tx, Value: res}, nil
}

func bitTyped(op ast.Operation, x ast.Int, y ast.Int) ast.Int {
	switch op {
	case '&':
		return x & y
	case '|':
		return x | y
	case '^':
		return x ^ y
	case ast.ANDN:
		return x &^ y
	default:
		panic("not reached")
	}
}

func bitUntyped(op ast.Operation, x *big.Int, y *big.Int) *big.Int {
	switch op {
	case '&':
		return new(big.Int).And(x, y)
	case '|':
		return new(big.Int).Or(x, y)
	case '^':
		return new(big.Int).Xor(x, y)
	case ast.ANDN:
		return new(big.Int).AndNot(x, y)
	default:
		panic("not reached")
	}
}

func Logical(
	op ast.Operation, x *ast.ConstValue, y *ast.ConstValue) (*ast.ConstValue, error) {

	tx, ty := builtinType(x.Typ), builtinType(y.Typ)
	if tx.Kind != ast.BUILTIN_BOOL && tx.Kind != ast.BUILTIN_UNTYPED_BOOL {
		return nil, &invalidOperation{Op: op, Type: tx}
	}
	if ty.Kind != ast.BUILTIN_BOOL && ty.Kind != ast.BUILTIN_UNTYPED_BOOL {
		return nil, &invalidOperation{Op: op, Type: ty}
	}
	if x.Typ != y.Typ {
		return nil, &mismatchedTypes{Op: op, X: x, Y: y}
	}
	var v bool
	if op == ast.AND {
		v = bool(x.Value.(ast.Bool)) && bool(y.Value.(ast.Bool))
	} else {
		v = bool(x.Value.(ast.Bool)) || bool(y.Value.(ast.Bool))
	}
	return &ast.ConstValue{Typ: x.Typ, Value: ast.Bool(v)}, nil
}

func Real(x *ast.ConstValue) (*ast.ConstValue, error) {
	tx := builtinType(x.Typ)
	switch tx.Kind {
	case ast.BUILTIN_COMPLEX64:
		v := float32(real(x.Value.(ast.Complex)))
		return &ast.ConstValue{Typ: ast.BuiltinFloat32, Value: ast.Float(v)}, nil
	case ast.BUILTIN_COMPLEX128:
		v := real(x.Value.(ast.Complex))
		return &ast.ConstValue{Typ: ast.BuiltinFloat64, Value: ast.Float(v)}, nil
	case ast.BUILTIN_UNTYPED_COMPLEX:
		re := new(big.Float).Copy(x.Value.(ast.UntypedComplex).Re)
		return &ast.ConstValue{
			Typ:   ast.BuiltinUntypedFloat,
			Value: ast.UntypedFloat{Float: re}}, nil
	case ast.BUILTIN_UNTYPED_FLOAT:
		re := x.Value.(ast.UntypedFloat).Float
		return &ast.ConstValue{
			Typ:   tx,
			Value: ast.UntypedFloat{Float: re}}, nil
	case ast.BUILTIN_UNTYPED_INT:
		re := bigIntToFloat(x.Value.(ast.UntypedInt).Int)
		return &ast.ConstValue{
			Typ:   ast.BuiltinUntypedFloat,
			Value: ast.UntypedFloat{Float: re}}, nil
	case ast.BUILTIN_UNTYPED_RUNE:
		re := bigIntToFloat(x.Value.(ast.Rune).Int)
		return &ast.ConstValue{
			Typ:   ast.BuiltinUntypedFloat,
			Value: ast.UntypedFloat{Float: re}}, nil
	default:
		return nil, &invalidOperation{Op: ast.REAL, Type: tx}
	}
}

func Imag(x *ast.ConstValue) (*ast.ConstValue, error) {
	tx := builtinType(x.Typ)
	switch tx.Kind {
	case ast.BUILTIN_COMPLEX64:
		v := float32(imag(x.Value.(ast.Complex)))
		return &ast.ConstValue{Typ: ast.BuiltinFloat32, Value: ast.Float(v)}, nil
	case ast.BUILTIN_COMPLEX128:
		v := imag(x.Value.(ast.Complex))
		return &ast.ConstValue{Typ: ast.BuiltinFloat64, Value: ast.Float(v)}, nil
	case ast.BUILTIN_UNTYPED_COMPLEX:
		im := new(big.Float).Copy(x.Value.(ast.UntypedComplex).Im)
		return &ast.ConstValue{
			Typ:   ast.BuiltinUntypedFloat,
			Value: ast.UntypedFloat{Float: im}}, nil
	case ast.BUILTIN_UNTYPED_FLOAT:
		zero := new(big.Float).
			SetPrec(ast.UNTYPED_FLOAT_PRECISION).
			SetMode(big.ToNearestEven)
		return &ast.ConstValue{
			Typ:   tx,
			Value: ast.UntypedFloat{Float: zero}}, nil
	case ast.BUILTIN_UNTYPED_INT:
		zero := new(big.Float).
			SetPrec(ast.UNTYPED_FLOAT_PRECISION).
			SetMode(big.ToNearestEven)
		return &ast.ConstValue{
			Typ:   ast.BuiltinUntypedFloat,
			Value: ast.UntypedFloat{Float: zero}}, nil
	case ast.BUILTIN_UNTYPED_RUNE:
		zero := new(big.Float).
			SetPrec(ast.UNTYPED_FLOAT_PRECISION).
			SetMode(big.ToNearestEven)
		return &ast.ConstValue{
			Typ:   ast.BuiltinUntypedFloat,
			Value: ast.UntypedFloat{Float: zero}}, nil
	default:
		return nil, &invalidOperation{Op: ast.IMAG, Type: tx}
	}
}

func Complex(x *ast.ConstValue, y *ast.ConstValue) (*ast.ConstValue, error) {
	tx, ty := builtinType(x.Typ), builtinType(y.Typ)
	if !tx.IsUntyped() {
		// If the operands are typed, they should be of the same type, and
		// either `float32` or `float64`
		if tx != ast.BuiltinFloat32 && tx != ast.BuiltinFloat64 {
			return nil, &invalidOperation{Op: ast.CMPLX, Type: tx}
		}
		if ty != ast.BuiltinFloat32 && ty != ast.BuiltinFloat64 {
			return nil, &invalidOperation{Op: ast.CMPLX, Type: ty}
		}
		if tx != ty {
			return nil, &mismatchedTypes{Op: ast.CMPLX, X: x, Y: y}
		}
		u, v := x.Value.(ast.Float), y.Value.(ast.Float)
		typ := ast.BuiltinComplex128
		if tx == ast.BuiltinFloat32 {
			typ = ast.BuiltinComplex64
		}
		return &ast.ConstValue{Typ: typ, Value: ast.Complex(complex(u, v))}, nil
	}

	// If the operands are untyped, they should be of a numeric type, and, if
	// complex, the imaginary part must be zero.
	var u *big.Float
	switch tx.Kind {
	case ast.BUILTIN_UNTYPED_INT:
		u = bigIntToFloat(x.Value.(ast.UntypedInt).Int)
	case ast.BUILTIN_UNTYPED_RUNE:
		u = bigIntToFloat(x.Value.(ast.Rune).Int)
	case ast.BUILTIN_UNTYPED_FLOAT:
		u = new(big.Float).Copy(x.Value.(ast.UntypedFloat).Float)
	case ast.BUILTIN_UNTYPED_COMPLEX:
		c := x.Value.(ast.UntypedComplex)
		if c.Im.Cmp(&ZeroFloat) != 0 {
			return nil, &badOperandValue{Op: ast.CMPLX, X: x}
		}
		u = new(big.Float).Copy(c.Re)
	default:
		return nil, &invalidOperation{Op: ast.CMPLX, Type: tx}
	}
	var v *big.Float
	switch ty.Kind {
	case ast.BUILTIN_UNTYPED_INT:
		v = bigIntToFloat(y.Value.(ast.UntypedInt).Int)
	case ast.BUILTIN_UNTYPED_RUNE:
		v = bigIntToFloat(y.Value.(ast.Rune).Int)
	case ast.BUILTIN_UNTYPED_FLOAT:
		v = new(big.Float).Copy(y.Value.(ast.UntypedFloat).Float)
	case ast.BUILTIN_UNTYPED_COMPLEX:
		c := y.Value.(ast.UntypedComplex)
		if c.Im.Cmp(&ZeroFloat) != 0 {
			return nil, &badOperandValue{Op: ast.CMPLX, X: y}
		}
		v = new(big.Float).Copy(c.Re)
	default:
		return nil, &invalidOperation{Op: ast.CMPLX, Type: ty}
	}
	return &ast.ConstValue{
		Typ: ast.BuiltinUntypedComplex, Value: ast.UntypedComplex{Re: u, Im: v}}, nil
}

// The badConversion error is returned when the destination type cannot
// represent the value of the converted constant.
type badConstConversion struct {
	Dst ast.Type
	Src *ast.ConstValue
}

func (e *badConstConversion) Error() string {
	return fmt.Sprintf("%s (`%s`) cannot be converted to `%s`",
		e.Src, e.Src.TypeString(), e.Dst)
}

type badShiftCount struct {
	X *ast.ConstValue
}

func (e *badShiftCount) Error() string {
	return fmt.Sprintf("invalid shift count `%s` (`%s`)", e.X, e.X.Typ)
}

type bigShiftCount struct {
	X uint64
}

func (e *bigShiftCount) Error() string {
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

type divisionByZero struct{}

func (e *divisionByZero) Error() string {
	return "division by zero"
}

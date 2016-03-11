package typecheck

import (
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

func convertUntypedInt(dst *ast.BuiltinType, val ast.UntypedInt) ast.Value {
	if dst.IsInteger() || dst.Kind == ast.BUILTIN_STRING {
		// If the destination type is integral, a successful conversion it not
		// possible if the untyped constant is outside the range of (u)int64.
		i := val.Int
		if i.Cmp(MinInt64) < 0 || i.Cmp(MaxUint64) > 0 {
			if dst.Kind == ast.BUILTIN_STRING {
				return ast.String(0xfffd)
			} else {
				return nil
			}
		}
		if i.Sign() < 0 {
			return convertInt(dst, ast.BuiltinInt64, uint64(i.Int64()))
		} else {
			return convertInt(dst, ast.BuiltinUint64, i.Uint64())
		}
	}
	switch dst.Kind {
	case ast.BUILTIN_BOOL:
		return nil
	case ast.BUILTIN_COMPLEX64:
		fallthrough
	case ast.BUILTIN_FLOAT32:
		f := new(big.Float).SetInt(val.Int)
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
		f := new(big.Float).SetInt(val.Int)
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
	switch v := val.(type) {
	case ast.Bool:
		res = convertBool(dst, v)
	case ast.Rune:
		res = convertInt(dst, ast.BuiltinInt32, uint64(v))
	case ast.UntypedInt:
		res = convertUntypedInt(dst, v)
	case ast.Int:
		res = convertInt(dst, src, uint64(v))
	case ast.UntypedFloat:
		if dst.IsInteger() {
			if v.IsInt() {
				i, _ := v.Int(nil)
				res = convertUntypedInt(dst, ast.UntypedInt{Int: i})
			}
		} else {
			x, _ := v.Float64()
			if !math.IsInf(x, 0) {
				res = convertFloat(dst, x)
			}
		}
	case ast.Float:
		res = convertFloat(dst, float64(v))
	case ast.UntypedComplex:
		if dst.IsInteger() {
			if v.Re.IsInt() && v.Im.Sign() == 0 {
				i, _ := v.Re.Int(nil)
				res = convertUntypedInt(dst, ast.UntypedInt{Int: i})
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
	case ast.Complex:
		re, im := real(v), imag(v)
		res = convertComplex(dst, re, im)
	case ast.String:
		if dst.Kind == ast.BUILTIN_STRING {
			res = v
		} else {
			res = nil
		}
	default:
		panic("not reached")
	}
	return res
}

func toInt(c *ast.ConstValue) (int64, bool) {
	src := builtinType(c.Typ)
	if src == nil && c.Typ != nil {
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
	if src == nil && c.Typ != nil {
		return 0, false
	}
	v := convertConst(ast.BuiltinFloat64, src, c.Value)
	if v == nil {
		return 0, false
	}
	return float64(v.(ast.Float)), true
}

func minus(typ *ast.BuiltinType, val ast.Value) ast.Value {
	switch v := val.(type) {
	case ast.Bool:
		return nil
	case ast.Rune:
		return ast.Rune(-int32(v))
	case ast.UntypedInt:
		return ast.UntypedInt{Int: new(big.Int).Neg(v.Int)}
	case ast.Int:
		if !typ.IsSigned() {
			return nil
		}
		return ast.Int(-int64(v))
	case ast.UntypedFloat:
		return ast.UntypedFloat{Float: new(big.Float).Neg(v.Float)}
	case ast.Float:
		return ast.Float(-float64(v))
	case ast.UntypedComplex:
		return ast.UntypedComplex{
			Re: new(big.Float).Neg(v.Re),
			Im: new(big.Float).Neg(v.Im),
		}
	case ast.Complex:
		return ast.Complex(complex(-real(v), -imag(v)))
	case ast.String:
		return nil
	default:
		panic("not reached")
	}
}

func complement(typ *ast.BuiltinType, val ast.Value) ast.Value {
	switch v := val.(type) {
	case ast.Int:
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
	case ast.UntypedInt:
		return ast.UntypedInt{Int: new(big.Int).Not(v.Int)}
	case ast.Rune:
		return ^v
	default:
		return nil
	}
}

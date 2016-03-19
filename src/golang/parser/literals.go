package parser

import (
	"golang/ast"
	"math/big"
	"unicode/utf8"
)

// Converts the bytes of a (perhap raw) string literal to a Go string. The
// literal value must have been obtained from and checked for correctness by
// the scanner and contains double or single quotation marks.
func String(lit []byte) ast.String {
	s := ""
	if lit[0] == '"' {
		s = readString(lit)
	} else {
		s = readRawString(lit)
	}
	return ast.String(s)
}

func isOctDigit(ch byte) bool {
	return '0' <= ch && ch <= '7'
}

func hexValue(ch byte) uint {
	switch {
	case '0' <= ch && ch <= '9':
		return uint(ch - '0')
	case 'a' <= ch && ch <= 'f':
		return 10 + uint(ch-'a')
	case 'A' <= ch && ch <= 'F':
		return 10 + uint(ch-'A')
	default:
		panic("not reached")
	}
}

// Reads three octal digits.
func readOct(lit []byte) rune {
	r := hexValue(lit[0])
	ch := lit[1]
	r = r<<3 | hexValue(ch)
	ch = lit[2]
	r = r<<3 | hexValue(ch)
	return rune(r)
}

// Reads N hexadecimal digits, return the value as a rune
func readHex(lit []byte, n uint) rune {
	var i, r uint = 0, 0
	for i < n {
		r = r<<4 + hexValue(lit[i])
		i++
	}
	if i == n {
		return rune(r)
	}
	panic("not reached")
}

// Scan an escape sequence. The parameter STR indicates if we are processing a
// string (or a rune) literal. Return the decoded rune in R. If the escape
// sequence was a two-digit hexadecimal sequence or a three-digit octal
// sequence, return true in the return B. This case needs to be distinguished
// when processing string literals.
func readEscape(lit []byte, str bool) (rune, bool, uint) {
	var r rune
	ch := lit[0]
	if isOctDigit(ch) {
		return readOct(lit), true, 3
	}
	switch {
	case ch == 'x':
		return readHex(lit[1:], 2), true, 3
	case ch == 'u':
		return readHex(lit[1:], 4), false, 5
	case ch == 'U':
		return readHex(lit[1:], 8), false, 9
	case ch == 'a':
		r = '\x07'
	case ch == 'b':
		r = '\x08'
	case ch == 'f':
		r = '\x0c'
	case ch == 'n':
		r = '\x0a'
	case ch == 'r':
		r = '\x0d'
	case ch == 't':
		r = '\x09'
	case ch == 'v':
		r = '\x0b'
	case ch == '\\':
		r = '\x5c'
	case !str && ch == '\'':
		r = '\x27'
	case str && ch == '"':
		r = '\x22'
	default:
		panic("not reached")
	}

	return r, false, 1
}

func readString(lit []byte) string {
	var buf []byte
	lit = lit[1:]
	for len(lit) > 1 {
		ch, n := utf8.DecodeRune(lit)
		lit = lit[n:]
		if ch == '\\' {
			c, b, n := readEscape(lit, true)
			lit = lit[n:]
			if b {
				buf = append(buf, byte(c))
				continue
			}
			ch = c
		}
		var tmp [4]byte
		n = utf8.EncodeRune(tmp[:], ch)
		buf = append(buf, tmp[:n]...)
	}
	return string(buf)
}

func readRawString(lit []byte) string {
	lit = lit[1:]
	var buf []byte
	for len(lit) > 1 {
		ch, n := utf8.DecodeRune(lit)
		lit = lit[n:]
		if ch != '\n' {
			var tmp [4]byte
			n = utf8.EncodeRune(tmp[:], ch)
			buf = append(buf, tmp[:n]...)
		}
	}
	return string(buf)
}

// Converts the bytes of a rune literal to a Go `rune` value. The
// literal value must have been obtained from and checked for correctness by
// the scanner.
func Rune(b []byte) ast.Rune {
	r, _ := utf8.DecodeRune(b)
	return ast.Rune{Int: big.NewInt(int64(r))}
}

var bigDigit = [...]*big.Int{
	big.NewInt(0), big.NewInt(1), big.NewInt(2), big.NewInt(3),
	big.NewInt(4), big.NewInt(5), big.NewInt(6), big.NewInt(7),
	big.NewInt(8), big.NewInt(9), big.NewInt(10), big.NewInt(11),
	big.NewInt(12), big.NewInt(13), big.NewInt(14), big.NewInt(15),
}

// Converts the bytes of an integer literal to an untyped Go integer
// constant. The literal value must have been obtained from and checked for
// correctness by the scanner.
func Int(b []byte) ast.UntypedInt {
	if b[0] == '0' && len(b) > 1 {
		if b[1] == 'x' || b[1] == 'X' {
			return ast.UntypedInt{Int: readInt(b[2:], 4)}
		} else {
			return ast.UntypedInt{Int: readInt(b[1:], 3)}
		}
	}
	return ast.UntypedInt{Int: readDecInt(b)}
}

func readInt(b []byte, n uint) *big.Int {
	x := big.NewInt(0)
	for _, d := range b {
		x.Lsh(x, n)
		x.Add(x, bigDigit[hexValue(d)])
	}
	return x
}

func readDecInt(b []byte) *big.Int {
	x := big.NewInt(0)
	for _, d := range b {
		x.Mul(x, bigDigit[10])
		x.Add(x, bigDigit[hexValue(d)])
	}
	return x
}

// Converts the bytes of a float literal to an untyped Go floating point
// constant. The literal value must have been obtained from and checked for
// correctness by the scanner.
func Float(b []byte) (ast.UntypedFloat, error) {
	x, _, err := big.ParseFloat(string(b), 0, 256, big.ToNearestEven)
	return ast.UntypedFloat{Float: x}, err
}

// Converts the bytes of a imaginary literal to an untyped Go complex floating
// point constant. The literal value must have been obtained from and checked
// for correctness by the scanner.
func Imaginary(b []byte) (ast.UntypedComplex, error) {
	x, _, err := big.ParseFloat(string(b), 0, 256, big.ToNearestEven)
	return ast.UntypedComplex{Re: big.NewFloat(0.0), Im: x}, err
}

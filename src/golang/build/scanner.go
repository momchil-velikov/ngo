package build

import (
	"bytes"
	"fmt"
	"io"
	"unicode"
	"unicode/utf8"
)

const (
	tEOF    = 0
	tERROR  = 1
	tLPAREN = '('
	tRPAREN = ')'
	tDOT    = '.'
	tSEMI   = ';'

	tPACKAGE = 256 + iota
	tID
	tIMPORT
	tSTRING
	tLINE_COMMENT
	tBLOCK_COMMENT
)

type token struct {
	Kind, Line, Col uint
	Value           string
}

func (t *token) String() string {
	switch t.Kind {
	case tEOF:
		return "<eof>"
	case tERROR:
		return "<error>"
	case tLPAREN:
		return "("
	case tRPAREN:
		return ")"
	case tDOT:
		return "."
	case tSEMI:
		return ";"
	case tPACKAGE:
		return "package"
	case tID:
		return fmt.Sprintf("<id:%s>", t.Value)
	case tIMPORT:
		return "import"
	case tSTRING:
		return fmt.Sprintf("<string:%s>", t.Value)
	case tLINE_COMMENT:
		return fmt.Sprintf("<line comment:%s>", t.Value)
	case tBLOCK_COMMENT:
		return fmt.Sprintf("<line comment:%s>", t.Value)
	default:
		panic("invalid token")
	}
}

// The build scanner is a simplified Go scanner, sufficient for parsing
// package import clauses.
type scanner struct {
	name      string        // Name of the source, e.g. a file name
	rd        io.RuneReader // Input stream
	err       error         // Current error
	rn        rune          // Current rune
	ln, col   uint          // Current rune line and column
	needSemi  bool          // emit semicolon at next newline
	forceSemi bool          // emit semicolon as the next token
}

func (s *scanner) init(name string, rd io.RuneReader) *scanner {
	s.name = name
	s.rd = rd
	s.needSemi = false
	s.forceSemi = false
	s.next()
	s.ln = 1
	s.col = 1
	return s
}

func (s *scanner) next() {
	if s.rn == '\n' {
		s.ln++
		s.col = 1
	} else {
		s.col++
	}
	s.rn, _, s.err = s.rd.ReadRune()
	if s.err != nil {
		if s.err == io.EOF {
			s.rn = tEOF
			s.err = nil
		} else {
			s.rn = tERROR
		}
	} else if s.rn == utf8.RuneError {
		s.rn = tERROR
		s.err = s.error("invalid UTF8 encoding")
	}
}

type scannerError struct {
	name    string
	ln, col uint
	msg     string
}

func (e *scannerError) Error() string {
	return fmt.Sprintf("%s:%d:%d: %s", e.name, e.ln, e.col, e.msg)
}

func (s *scanner) error(msg string) error {
	return &scannerError{name: s.name, ln: s.ln, col: s.col, msg: msg}
}

// Scans until end of line and returns '\n' on success. Returns tERROR on I/O
// error or invalid utf8 encoding, or tEOF if end of source reached before the
// comment ends.
func (s *scanner) scanLineComment() (uint, string) {
	var buf []rune
	for s.rn != tEOF && s.rn != tERROR && s.rn != '\n' {
		buf = append(buf, s.rn)
		s.next()
	}
	if s.rn == '\n' || s.rn == tEOF {
		s.next()
		return '\n', string(buf)
	}
	return tERROR, ""
}

// Scans until the end of a block comment. Returns a newline if the comment
// contains any newlines or a space, otherwise. Returns tERROR on I/O error or
// invalid utf8 encoding, or tEOF if end of source reached before the comment
// ends.
func (s *scanner) scanBlockComment() (uint, string) {
	t := uint(' ')
	var buf []rune
	for s.rn != tEOF && s.rn != tERROR {
		if s.rn == '\n' {
			t = '\n'
		}
		if s.rn == '*' {
			s.next()
			if s.rn == '/' {
				break
			}
			buf = append(buf, '*')
		} else {
			buf = append(buf, s.rn)
			s.next()
		}
	}

	if s.rn == '/' {
		s.next()
		return t, string(buf)
	}
	if s.rn == tEOF {
		s.err = s.error("EOF in comment")
	}
	return tERROR, ""
}

// Reads three octal digits.
func (s *scanner) readOct() (uint, bool) {
	r := hexValue(s.rn)
	s.next()
	if isOctDigit(s.rn) {
		r = r<<3 | hexValue(s.rn)
		s.next()
		if isOctDigit(s.rn) {
			r = r<<3 | hexValue(s.rn)
			s.next()
			return r, true
		}
	}
	s.err = s.error("invalid octal escape sequence")
	return tERROR, false
}

// Reads N hexadecimal digits.
func (s *scanner) readHex(n uint) (uint, bool) {
	var i, r uint = 0, 0
	for i < n && isHexDigit(s.rn) {
		r = r<<4 + hexValue(s.rn)
		s.next()
		i++
	}
	if i == n {
		return r, true
	}
	s.err = s.error("invalid hex escape sequence")
	return tERROR, false
}

// Scan an escape sequence.  Return the decoded rune in R. If the escape
// sequence was a two-digit hexadecimal sequence or a three-digit octal
// sequence, return true in the return B.
func (s *scanner) scanEscape() (uint, bool, bool) {
	var r uint
	if isOctDigit(s.rn) {
		r, ok := s.readOct()
		return r, true, ok
	}
	switch s.rn {
	case 'x':
		s.next()
		r, ok := s.readHex(2)
		return r, true, ok
	case 'u':
		s.next()
		r, ok := s.readHex(4)
		return r, false, ok
	case 'U':
		s.next()
		r, ok := s.readHex(8)
		return r, false, ok
	case 'a':
		r = '\x07'
	case 'b':
		r = '\x08'
	case 'f':
		r = '\x0c'
	case 'n':
		r = '\x0a'
	case 'r':
		r = '\x0d'
	case 't':
		r = '\x09'
	case 'v':
		r = '\x0b'
	case '\\':
		r = '\x5c'
	case '"':
		r = '\x22'
	default:
		s.err = s.error("invalid escape sequence")
		return tERROR, true, false
	}
	s.next()
	return r, false, true
}

// Scans a string literal.
func (s *scanner) scanString() (uint, string) {
	buf := bytes.Buffer{}
	for s.rn != '"' && s.rn != '\n' && s.rn != tEOF && s.rn != tERROR {
		r := s.rn
		s.next()
		if r == '\\' {
			v, b, ok := s.scanEscape()
			if !ok {
				return tERROR, ""
			}
			if b {
				buf.WriteByte(byte(v))
				continue
			}
			r = rune(v)
		}
		buf.WriteRune(r)
	}

	if s.rn == '"' {
		s.next()
		return tSTRING, string(buf.Bytes())
	}
	if s.rn == '\n' || s.rn == tEOF {
		s.err = s.error("unterminated string literal")
	}
	return tERROR, ""
}

func (s *scanner) scanRawString() (uint, string) {
	var buf []rune
	for s.rn != '`' && s.rn != tEOF && s.rn != tERROR {
		r := s.rn
		s.next()
		if r != '\r' {
			buf = append(buf, r)
		}
	}
	if s.rn == '`' {
		s.next()
		return tSTRING, string(buf)
	}
	if s.rn == tEOF {
		s.err = s.error("unterminated raw string literal")
	}
	return tERROR, ""
}

// Scan an identifier. Upon entry the current rune is the first letter of the
// idenfifier.
func (s *scanner) scanIdent() string {
	var buf []rune
	for isLetter(s.rn) || isDigit(s.rn) {
		buf = append(buf, s.rn)
		s.next()
	}
	return string(buf)
}

// Main scanner routine. Get the next token.
func (s *scanner) Get() (token, error) {
	// Emit a forced semicolon after block comment
	if s.forceSemi {
		s.forceSemi = false
		s.needSemi = false
		return token{';', s.ln, s.col, ""}, nil
	}

	// Skip over whitespace and insert semicolons
	for {
		if s.rn == tEOF {
			if s.needSemi {
				s.needSemi = false
				return token{';', s.ln, s.col, ""}, nil
			} else {
				return token{tEOF, s.ln, s.col, ""}, nil
			}
		}

		if !isWhitespace(s.rn) {
			break
		}

		if s.rn == '\n' {
			if s.needSemi {
				s.needSemi = false
				return token{';', s.ln, s.col, ""}, nil
			}
		}
		s.next()
	}

	ln, col := s.ln, s.col
	needSemi := s.needSemi
	s.needSemi = false
	switch s.rn {
	case tERROR:
		return token{tERROR, ln, col, ""}, s.err
	case '/': // /*, //
		s.next()
		if s.rn == '/' {
			s.next()
			if r, v := s.scanLineComment(); r == tERROR {
				return token{tERROR, ln, col, ""}, s.err
			} else {
				s.needSemi = needSemi
				return token{tLINE_COMMENT, ln, col, v}, nil
			}
		} else if s.rn == '*' {
			s.next()
			if r, v := s.scanBlockComment(); r == tERROR {
				return token{tERROR, ln, col, ""}, s.err
			} else {
				s.needSemi = needSemi
				if needSemi && r == '\n' {
					// Block comments, which contain one or more newlines act
					// as a newline. Since we are returning the comment itself
					// now, on the next token force emit a semicolon.
					s.forceSemi = true
				}
				return token{tBLOCK_COMMENT, ln, col, v}, nil
			}
		}
		return token{tERROR, ln, col, ""}, s.error("invalid token")
	case '(', '.', ';':
		r := s.rn
		s.next()
		return token{uint(r), ln, col, ""}, nil
	case ')':
		s.next()
		s.needSemi = true
		return token{')', ln, col, ""}, nil
	case '"':
		// Scan string literal
		s.next()
		t, v := s.scanString()
		if t == tSTRING {
			s.needSemi = true
			return token{tSTRING, ln, col, v}, nil
		} else {
			return token{t, ln, col, ""}, s.err
		}
	case '`':
		s.next()
		// Scan raw string literal
		t, v := s.scanRawString()
		if t == tSTRING {
			s.needSemi = true
			return token{tSTRING, ln, col, v}, nil
		} else {
			return token{t, ln, col, ""}, s.err
		}
	}

	if isLetter(s.rn) {
		id := s.scanIdent()
		if id == "import" {
			return token{tIMPORT, ln, col, ""}, nil
		}
		if id == "package" {
			return token{tPACKAGE, ln, col, ""}, nil
		}
		s.needSemi = true
		return token{tID, ln, col, id}, nil
	}

	return token{tERROR, ln, col, ""}, s.error("invalid source character")
}

func isWhitespace(ch rune) bool {
	switch ch {
	case ' ', '\t', '\n', '\r':
		return true
	default:
		return false
	}
}

func isLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' ||
		'A' <= ch && ch <= 'Z' ||
		ch == '_' || ch >= 0x80 && unicode.IsLetter(ch)
}

func isOctDigit(ch rune) bool {
	return '0' <= ch && ch <= '7'
}

func isHexDigit(ch rune) bool {
	return '0' <= ch && ch <= '9' || 'a' <= ch && ch <= 'f' || 'A' <= ch && ch <= 'F'
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9' || ch >= 0x80 && unicode.IsDigit(ch)
}

func hexValue(ch rune) uint {
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

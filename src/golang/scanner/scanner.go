package scanner

import (
	"fmt"
	"unicode/utf8"
)

type Scanner struct {
	Name     string // Name of the source, e.g. a file name
	Err      error
	srcmap   SourceMap
	Value    []byte // Value of the last token
	TOff     int    // Byte offset of the last token
	src      []byte // source characters
	off      int    // Curent byte offset
	lineOff  int    // Current line offset
	needSemi bool
}

func New(name string, src string) *Scanner {
	s := &Scanner{}
	s.Init(name, []byte(src))
	return s
}

func (s *Scanner) Init(name string, src []byte) {
	s.Name = name
	s.src = src
}

func (s *Scanner) Position(off int) (int, int) {
	if off >= s.lineOff {
		return s.srcmap.LineCount() + 1, off - s.lineOff + 1
	} else {
		return s.srcmap.Position(off)
	}
}

func (s *Scanner) error(msg string) uint {
	var line []byte
	off := s.lineOff
	for r, sz := s.peekAt(off); r != NL && r != EOF; r, sz = s.peekAt(off) {
		off += sz
	}

	if s.lineOff < s.TOff {
		line = append(line, s.src[s.lineOff:s.TOff]...)
	}
	line = append(line, " ->|"...)

	line = append(line, s.src[s.TOff:s.off]...)

	line = append(line, "|<- "...)
	line = append(line, s.src[s.off:off]...)
	ln, pos := s.Position(s.off)
	s.Err = makeError(s.Name, ln, pos, msg, string(line))
	return ERROR
}

// Return the rune at the given offset
func (s *Scanner) peekAt(off int) (rune, int) {
	if off < 0 || off >= len(s.src) {
		return EOF, 0
	} else {
		return utf8.DecodeRune(s.src[off:])
	}
}

// Return the rune at the current offset
func (s *Scanner) peek() rune {
	r, _ := s.peekAt(s.off)
	return r
}

// Position at the next codepoint
func (s *Scanner) next() rune {
	r, sz := s.peekAt(s.off)
	if sz > 0 {
		s.off += sz
		if r == NL {
			s.srcmap.addLine(s.off - s.lineOff)
			s.lineOff = s.off
		}
	}
	return r
}

// Return the one byte rune at the given offset
func (s *Scanner) peekCharAt(off int) rune {
	if off >= len(s.src) {
		return EOF
	} else {
		return rune(s.src[off])
	}
}

// Return the one byte rune at the current offset
func (s *Scanner) peekChar() rune {
	return s.peekCharAt(s.off)
}

func (s *Scanner) nextChar() rune {
	r := s.peekChar()
	if r != EOF {
		s.off++
		if r == NL {
			s.srcmap.addLine(s.off - s.lineOff)
		}
	}
	return r
}

// Skip until end of line. Return NL or EOF if end of source reached
// before the comment ends.
func (s *Scanner) skipLineComment() rune {
	r := s.next()
	for r != EOF && r != NL {
		r = s.next()
	}
	return r
}

// Skip until the end of a block comment. Return EOF if end of source reached
// before the comment ends. Return NL if the comment contains newlines or
// SPACE otherwise.
func (s *Scanner) skipBlockComment() rune {
	line := s.srcmap.LineCount() + 1

	var r rune
	for r = s.next(); r != EOF; r = s.next() {
		if r == '*' {
			r = s.peek()
			if r == '/' {
				s.nextChar()
				break
			}
		}
	}

	if r == EOF {
		return EOF
	} else {
		if line < s.srcmap.LineCount()+1 {
			return NL
		} else {
			return SPACE
		}
	}
}

// Scan an identifier. Upon entry the source offset is positioned at the first
// letter of the idenfifier.
func (s *Scanner) scanIdent() []byte {
	off := s.off
	r := s.peek()
	for isLetter(r) || isDigit(r) {
		s.next()
		r = s.peek()
	}
	return s.src[off:s.off]
}

func (s *Scanner) scanHex() {
	for ch := s.peekChar(); isHexDigit(ch); ch = s.peekChar() {
		s.nextChar()
	}
}

func (s *Scanner) scanOct() {
	for ch := s.peekChar(); isOctDigit(ch); ch = s.peekChar() {
		s.nextChar()
	}
}

func (s *Scanner) scanDec() {
	for ch := s.peekChar(); isDecDigit(ch); ch = s.peekChar() {
		s.nextChar()
	}
}

func (s *Scanner) scanFraction(start int) (uint, []byte) {
	ch := s.peekChar()
	if ch == '.' {
		s.nextChar()
		ch = s.peekChar()
		if isDecDigit(ch) {
			s.scanDec()
			ch = s.peekChar()
		}
	}

	if ch == 'e' || ch == 'E' {
		s.nextChar()
		ch = s.peekChar()
		if ch == '+' || ch == '-' {
			s.nextChar()
			ch = s.peekChar()
		}
		if isDecDigit(ch) {
			s.scanDec()
			ch = s.peekChar()
		} else {
			return s.error("Invalid exponent"), nil
		}
	}

	if ch == 'i' {
		s.nextChar()
		return IMAGINARY, s.src[start : s.off-1]
	} else {
		return FLOAT, s.src[start:s.off]
	}
}

// Scan a numeric literal.
func (s *Scanner) scanNumeric() (uint, []byte) {
	start := s.off
	r := s.peekChar()

	switch {
	case r == '0':
		s.nextChar()
		r = s.peekChar()
		if r == 'x' || r == 'X' {
			// scan hexadecimal
			s.nextChar()
			s.scanHex()
			if s.off-start > 2 {
				return INTEGER, s.src[start:s.off]
			} else {
				return s.error("Invalid hex literal"), nil
			}
		} else {
			// scan octal or floating
			d := false
			s.scanOct()
			r = s.peekChar()
			if r == '8' || r == '9' {
				d = true
				s.scanDec()
				r = s.peekChar()
			}
			if r == 'i' {
				s.nextChar()
				return IMAGINARY, s.src[start : s.off-1]
			} else if r == '.' || r == 'E' || r == 'e' {
				return s.scanFraction(start)
			} else if d {
				return s.error("Invalid octal literal"), nil
			} else {
				return INTEGER, s.src[start:s.off]
			}
		}

	case r == '.':
		return s.scanFraction(start)

	case isDecDigit(r):
		s.scanDec()
		r = s.peekChar()
		if r == '.' || r == 'e' || r == 'E' {
			return s.scanFraction(start)
		} else if r == 'i' {
			s.nextChar()
			return IMAGINARY, s.src[start : s.off-1]
		} else {
			return INTEGER, s.src[start:s.off]
		}

	default:
		panic("should not reach here")
	}
}

// Scans N hexadecimal characters. Returns true if exactly N characters
// scanned.
func (s *Scanner) scanHexN(n uint) bool {
	i := uint(0)
	for ch := s.peekChar(); i < n && isHexDigit(ch); i++ {
		s.nextChar()
		ch = s.peekChar()
	}

	if i == n {
		return true
	} else {
		s.error("Invalid hex escape sequence")
		return false
	}
}

// Read N octal characters, return the value as a rune
func (s *Scanner) scanOctN(n uint) bool {
	i := uint(0)
	for ch := s.peekChar(); i < n && isOctDigit(ch); i++ {
		s.nextChar()
		ch = s.peekChar()
	}

	if i == n {
		return true
	} else {
		s.error("Invalid octal escape sequence")
		return false
	}
}

// Scans an escape sequence. The parameter STR indicates if we are processing
// a string (or a rune) literal.
func (s *Scanner) scanEscape(str bool) bool {
	ch := s.peekChar()
	if isOctDigit(ch) {
		return s.scanOctN(3)
	}
	s.nextChar()
	ok := false
	switch ch {
	case 'x':
		ok = s.scanHexN(2)
	case 'u':
		ok = s.scanHexN(4)
	case 'U':
		ok = s.scanHexN(8)
	case 'a', 'b', 'f', 'n', 'r', 't', 'v', '\\':
		ok = true
	default:
		if !str && ch == '\'' || str && ch == '"' {
			ok = true
		} else {
			s.error("Invalid escape sequence")
		}
	}
	return ok
}

// Scan a rune literal, including the apostrophes. Return a string, containing a
// single rune.
func (s *Scanner) scanRune() (uint, []byte) {
	s.nextChar() // skip leading apostrophe
	start := s.off
	ch := s.peekChar()

	if ch == '\'' {
		s.nextChar()
		return s.error("Empty rune literal"), nil
	}

	if ch == EOF {
		return s.error("EOF in rune literal"), nil
	}

	if ch == '\\' {
		s.nextChar()
		if ok := s.scanEscape(false); !ok {
			return ERROR, nil
		}
	} else {
		s.next()
	}

	ch = s.peekChar()
	if ch != '\'' {
		return s.error("Unterminated rune literal"), nil
	}

	s.nextChar()
	return RUNE, s.src[start : s.off-1]
}

// Scans a string literal. Returns raw source bytes including the quotation
// marks.
func (s *Scanner) scanString() (uint, []byte) {
	start := s.off
	s.nextChar()

	var ch rune
	for ch = s.peek(); ch != '"' && ch != EOF && ch != NL; ch = s.peek() {
		s.next()
		if ch == '\\' {
			if ok := s.scanEscape(true); !ok {
				return ERROR, nil
			}
		}
	}

	if ch == '"' {
		s.nextChar()
		return STRING, s.src[start:s.off]
	} else {
		return s.error("Unterminated string literal"), nil
	}
}

func (s *Scanner) scanRawString() (uint, []byte) {
	start := s.off
	s.nextChar()

	ch := s.peek()
	for ch != '`' && ch != EOF {
		s.next()
		ch = s.peek()
	}

	if ch == '`' {
		s.nextChar()
		return STRING, s.src[start:s.off]
	} else {
		return s.error("Unterminated raw string literal"), nil
	}
}

// Main scanner routine. Get the next token.
func (s *Scanner) Get() uint {
	s.Err = nil

L:
	needSemi := s.needSemi

	if s.off == len(s.src) {
		if needSemi {
			s.needSemi = false
			return ';'
		} else {
			return EOF
		}
	}

	r := s.peek()

	s.TOff = s.off
	if isWhitespace(r) {
		s.next()
		if r == NL {
			if needSemi {
				s.needSemi = false
				return ';'
			}
		}
		goto L
	}

	s.needSemi = false

	switch r {
	case '+': // +, +=, ++
		s.nextChar()
		r = s.peekChar()
		if r == '=' {
			s.nextChar()
			return PLUS_ASSIGN
		} else if r == '+' {
			s.nextChar()
			s.needSemi = true
			return INC
		}
		return '+'
	case '-': // -, -=, --
		s.nextChar()
		r = s.peekChar()
		if r == '=' {
			s.nextChar()
			return MINUS_ASSIGN
		} else if r == '-' {
			s.nextChar()
			s.needSemi = true
			return DEC
		}
		return '-'
	case '*': // *, *=
		s.nextChar()
		r = s.peekChar()
		if r == '=' {
			s.nextChar()
			return MUL_ASSIGN
		}
		return '*'
	case '/': // /, /=, /*, //
		s.nextChar()
		r = s.peekChar()
		if r == '=' {
			s.nextChar()
			return DIV_ASSIGN
		} else if r == '/' {
			s.nextChar()
			r = s.skipLineComment()
			if r == EOF {
				return s.error("EOF in comment")
			}
			if r == NL && needSemi {
				return ';'
			}
			goto L
		} else if r == '*' {
			s.nextChar()
			r = s.skipBlockComment()
			if r == EOF {
				return s.error("EOF in comment")
			} else if r == NL && needSemi {
				return ';'
			}
			s.needSemi = needSemi
			goto L
		}
		return '/'
	case '%': // %, %=
		s.nextChar()
		r = s.peekChar()
		if r == '=' {
			s.nextChar()
			return REM_ASSIGN
		}
		return '%'
	case '&': // &, &=, &^, &^=, &&
		s.nextChar()
		r = s.peekChar()
		if r == '=' {
			s.nextChar()
			return AND_ASSIGN
		} else if r == '^' {
			s.nextChar()
			r = s.peekChar()
			if r == '=' {
				s.nextChar()
				return ANDN_ASSIGN
			}
			return ANDN
		} else if r == '&' {
			s.nextChar()
			return AND
		}
		return '&'
	case '|': // |, |=, ||
		s.nextChar()
		r = s.peekChar()
		if r == '=' {
			s.nextChar()
			return OR_ASSIGN
		} else if r == '|' {
			s.nextChar()
			return OR
		}
		return '|'
	case '^': // ^, ^=
		s.nextChar()
		r = s.peekChar()
		if r == '=' {
			s.nextChar()
			return XOR_ASSIGN
		}
		return '^'
	case '<': // <, <<, <<=, <=, <-
		s.nextChar()
		r = s.peekChar()
		if r == '<' {
			s.nextChar()
			r = s.peekChar()
			if r == '=' {
				s.nextChar()
				return SHL_ASSIGN
			}
			return SHL
		} else if r == '=' {
			s.nextChar()
			return LE
		} else if r == '-' {
			s.nextChar()
			return RECV
		}
		return '<'
	case '>': // >, >>, >>=, >=
		s.nextChar()
		r = s.peekChar()
		if r == '>' {
			s.nextChar()
			r = s.peekChar()
			if r == '=' {
				s.nextChar()
				return SHR_ASSIGN
			}
			return SHR
		} else if r == '=' {
			s.nextChar()
			return GE
		}
		return '>'
	case '=': // =, ==
		s.nextChar()
		r = s.peekChar()
		if r == '=' {
			s.nextChar()
			return EQ
		}
		return '='
	case '!': // !, !=
		s.nextChar()
		r = s.peekChar()
		if r == '=' {
			s.nextChar()
			return NE
		}
		return '!'
	case '(', '[', '{', ',', ';':
		s.nextChar()
		return uint(r)
	case ')', ']', '}':
		s.nextChar()
		s.needSemi = true
		return uint(r)
	case '.': // ., ..., .<fraction>
		r = s.peekCharAt(s.off + 1)
		if r == '.' && s.peekCharAt(s.off+2) == '.' {
			s.nextChar()
			s.nextChar()
			s.nextChar()
			return DOTS
		} else if isDecDigit(r) {
			t, v := s.scanNumeric()
			s.Value = v
			s.needSemi = true
			return t
		} else {
			s.nextChar()
			return '.'
		}
	case ':': // :, :=
		s.nextChar()
		r = s.peekChar()
		if r == '=' {
			s.nextChar()
			return DEFINE
		}
		return ':'
	case '\'':
		// Scan  rune literal
		t, v := s.scanRune()
		if t == RUNE {
			s.Value = v
			s.needSemi = true
			return RUNE
		} else {
			return t
		}
	case '"':
		// Scan string literal
		t, v := s.scanString()
		if t == STRING {
			s.Value = v
			s.needSemi = true
			return STRING
		} else {
			return t
		}
	case '`':
		// Scan raw string literal
		t, v := s.scanRawString()
		if t == STRING {
			s.Value = v
			s.needSemi = true
			return STRING
		} else {
			return t
		}
	}

	if isLetter(r) {
		id := s.scanIdent()
		kw, ok := keywords[string(id)]
		if ok {
			switch kw {
			case BREAK, CONTINUE, FALLTHROUGH, RETURN:
				s.needSemi = true
			}
			return kw
		} else {
			s.Value = id
			s.needSemi = true
			return ID
		}
	} else if isDecDigit(r) {
		// Scan numeric literal
		t, v := s.scanNumeric()
		s.needSemi = true
		s.Value = v
		return t
	} else {
		s.next()
		msg := fmt.Sprintf("Invalid character: |%c|", r)
		return s.error(msg)
	}
}

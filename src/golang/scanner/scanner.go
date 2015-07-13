package scanner

import (
	"fmt"
	"unicode/utf8"
)

type Scanner struct {
	Name              string // Name of the source, e.g. a file name
	Err               error
	Value             string // Value of the last token
	TOff, TLine, TPos int    // Byte offset, line number, line position of the last token
	src               string // source characters
	off, line, pos    int    // Curent byte offset, line number and line position
	lineOff           int    // Current line offset
	needSemi          bool
}

func New(name string, src string) *Scanner {
	s := &Scanner{}
	s.Init(name, src)
	return s
}

func (s *Scanner) Init(name string, src string) {
	s.Name = name
	s.src = src
	s.line = 1
	s.pos = 1
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
	s.Err = makeError(s.Name, s.line, s.pos, msg, string(line))
	return ERROR
}

// Return the rune at the given offset
func (s *Scanner) peekAt(off int) (rune, int) {
	if off < 0 || off >= len(s.src) {
		return EOF, 0
	} else {
		return utf8.DecodeRuneInString(s.src[off:])
	}
}

// Return the rune at the current offset
func (s *Scanner) peek() (rune, int) {
	return s.peekAt(s.off)
}

// Position at the next codepoint
func (s *Scanner) next() rune {
	r, sz := s.peek()
	if sz > 0 {
		s.off += sz
		s.pos++
		if r == NL {
			s.pos = 1
			s.line++
			s.lineOff = s.off
		}
	}
	return r
}

// Return the one byte rune at the current offset
func (s *Scanner) peekChar() rune {
	if s.off == len(s.src) {
		return EOF
	} else {
		return rune(s.src[s.off])
	}
}

func (s *Scanner) nextChar() rune {
	r := s.peekChar()
	if r != EOF {
		s.off++
		s.pos++
		if r == NL {
			s.pos = 1
			s.line++
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
	line := s.line

	var r rune
	for r = s.next(); r != EOF; r = s.next() {
		if r == '*' {
			r, _ = s.peek()
			if r == '/' {
				s.nextChar()
				break
			}
		}
	}

	if r == EOF {
		return EOF
	} else {
		if line < s.line {
			return NL
		} else {
			return SPACE
		}
	}
}

// Scan an identifier. Upon entry the source offset is positioned at the first
// letter of the idenfifier.
func (s *Scanner) scanIdent() string {
	off := s.off
	r, _ := s.peek()
	for isLetter(r) || isDigit(r) {
		s.next()
		r, _ = s.peek()
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

func (s *Scanner) scanFraction(start int) (uint, string) {
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
			return s.error("Invalid exponent"), ""
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
func (s *Scanner) scanNumeric() (uint, string) {
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
				return s.error("Invalid hex literal"), ""
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
				return s.error("Invalid octal literal"), ""
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
		} else {
			return INTEGER, s.src[start:s.off]
		}

	default:
		panic("should not reach here")
	}
}

// Read N hexadecimal characters, return the value as a rune
func (s *Scanner) readHex(n uint) rune {
	var i, r uint = 0, 0
	for ch := s.peekChar(); i < n && isHexDigit(ch); i++ {
		r = r<<4 + hexValue(ch)
		s.nextChar()
		ch = s.peekChar()
	}

	if i == n {
		return rune(r)
	} else {
		s.error("Invalid hex escape sequence")
		return utf8.RuneError
	}
}

// Read N octal characters, return the value as a rune
func (s *Scanner) readOct(n uint) rune {
	var i, r uint = 0, 0
	for ch := s.peekChar(); i < n && isOctDigit(ch); i++ {
		r = r<<3 + hexValue(ch)
		s.nextChar()
		ch = s.peekChar()
	}

	if i == n {
		return rune(r)
	} else {
		s.error("Invalid octal escape sequence")
		return utf8.RuneError
	}
}

// Scan an escape sequence. The parameter STR indicates if we are processing a
// string (or a rune) literal. Return the decoded rune in R. If the escape
// sequence was a two-digit hexadecimal sequence of three-digit octal sequence,
// return true in the return B. This case needs to be distinguished when
// processing string literals.
func (s *Scanner) scanEscape(str bool) (r rune, b bool, ok bool) {
	ch := s.peekChar()

	if isOctDigit(ch) {
		b = true
		r = s.readOct(3)
		ok = r != utf8.RuneError
		return
	}

	s.nextChar()
	b, ok = false, true
	switch {
	case ch == 'x':
		r, b = s.readHex(2), true
		ok = r != utf8.RuneError
	case ch == 'u':
		r = s.readHex(4)
	case ch == 'U':
		r = s.readHex(8)
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
		s.error("Invalid escape sequence")
		return utf8.RuneError, false, false
	}

	return
}

// Scan a rune literal, including the apostrophes. Return a string, containing a
// single rune.
func (s *Scanner) scanRune() (uint, string) {
	s.nextChar() // skip leading apostrophe
	ch := s.peekChar()

	if ch == '\'' {
		s.nextChar()
		return s.error("Empty rune literal"), ""
	}

	if ch == EOF {
		return s.error("EOF in rune literal"), ""
	}

	var r rune
	if ch == '\\' {
		s.nextChar()
		rr, _, ok := s.scanEscape(false)
		if !ok {
			return ERROR, ""
		}
		r = rr
	} else {
		r = s.next()
	}

	ch = s.peekChar()
	if ch != '\'' {
		return s.error("Unterminated rune literal"), ""
	}

	s.nextChar()
	return RUNE, string(r)
}

func (s *Scanner) scanString() (uint, string) {
	s.nextChar() // skip leading quotation mark

	var ch rune
	var str []byte
	for ch, _ = s.peek(); ch != '"' && ch != EOF && ch != NL; ch, _ = s.peek() {
		s.next()
		if ch == '\\' {
			c, b, ok := s.scanEscape(true)
			if !ok {
				return ERROR, ""
			}
			if b {
				str = append(str, byte(c))
				continue
			}
			ch = c
		}
		var tmp [4]byte
		sz := utf8.EncodeRune(tmp[:], ch)
		str = append(str, tmp[:sz]...)
	}

	if ch == '"' {
		s.nextChar()
		return STRING, string(str)
	} else {
		return s.error("Unterminated string literal"), ""
	}
}

func (s *Scanner) scanRawString() (uint, string) {
	s.nextChar() // skip leading back apostrophe

	var ch rune
	var str []byte
	for ch, _ = s.peek(); ch != '`' && ch != EOF; ch, _ = s.peek() {
		if ch != CR {
			var tmp [4]byte
			sz := utf8.EncodeRune(tmp[:], ch)
			str = append(str, tmp[:sz]...)
		}
		s.next()
	}

	if ch == '`' {
		s.nextChar()
		return STRING, string(str)
	} else {
		return s.error("Unterminated raw string literal"), ""
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

	r, _ := s.peek()

	s.TOff, s.TLine, s.TPos = s.off, s.line, s.pos
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
	case '.': // ., ...
		s.nextChar()
		r = s.peekChar()
		if r == '.' {
			s.nextChar()
			r = s.peekChar()
			if r == '.' {
				s.nextChar()
				return DOTS
			}
			s.off--
			s.pos--
		} else if isDecDigit(r) {
			// Scan numeric literal
			s.off--
			t, v := s.scanNumeric()
			s.Value = v
			s.needSemi = true
			return t
		}
		return '.'
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
		kw, ok := keywords[id]
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

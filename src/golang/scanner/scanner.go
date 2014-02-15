package scanner

import (
    "fmt"
    "unicode/utf8"
)

type Scanner struct {
    Name string // Name of the source, e.g. a file name

    Err               error
    Value             string // Value of the last token
    TOff, TLine, TPos int    // Byte offset, line number, line position of the last token

    src            string // source characters
    off, line, pos int    // Curent byte offset, line number and line position
    line_off       int    // Current line offset
    need_semicolon bool
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
    off := s.line_off
    for r, sz := s.peek_at(off); r != NL && r != EOF; r, sz = s.peek_at(off) {
        off += sz
    }

    if s.line_off < s.TOff {
        line = append(line, s.src[s.line_off:s.TOff]...)
    }
    line = append(line, " ->|"...)

    line = append(line, s.src[s.TOff:s.off]...)

    line = append(line, "|<- "...)
    line = append(line, s.src[s.off:off]...)
    s.Err = make_error(s.Name, s.line, s.pos, msg, string(line))
    return ERROR
}

// Return the rune at the given offset
func (s *Scanner) peek_at(off int) (rune, int) {
    if off < 0 || off >= len(s.src) {
        return EOF, 0
    } else {
        return utf8.DecodeRuneInString(s.src[off:])
    }
}

// Return the rune at the current offset
func (s *Scanner) peek() (rune, int) {
    return s.peek_at(s.off)
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
            s.line_off = s.off
        }
    }
    return r
}

// Return the one byte rune at the current offset
func (s *Scanner) peek_char() rune {
    if s.off == len(s.src) {
        return EOF
    } else {
        return rune(s.src[s.off])
    }
}

func (s *Scanner) next_char() rune {
    r := s.peek_char()
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
func (s *Scanner) skip_line_comment() rune {
    r := s.next()
    for r != EOF && r != NL {
        r = s.next()
    }
    return r
}

// Skip until the end of a block comment. Return EOF if end of source reached
// before the comment ends. Return NL if the comment contains newlines or
// SPACE otherwise.
func (s *Scanner) skip_block_comment() rune {
    line := s.line

    var r rune
    for r = s.next(); r != EOF; r = s.next() {
        if r == '*' {
            r, _ = s.peek()
            if r == '/' {
                s.next_char()
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
func (s *Scanner) scan_ident() string {
    off := s.off
    r, _ := s.peek()
    for is_letter(r) || is_digit(r) {
        s.next()
        r, _ = s.peek()
    }
    return s.src[off:s.off]
}

func (s *Scanner) scan_hex() {
    for ch := s.peek_char(); is_hex_digit(ch); ch = s.peek_char() {
        s.next_char()
    }
}

func (s *Scanner) scan_oct() {
    for ch := s.peek_char(); is_oct_digit(ch); ch = s.peek_char() {
        s.next_char()
    }
}

func (s *Scanner) scan_dec() {
    for ch := s.peek_char(); is_dec_digit(ch); ch = s.peek_char() {
        s.next_char()
    }
}

func (s *Scanner) scan_fract(start int) (uint, string) {
    ch := s.peek_char()
    if ch == '.' {
        s.next_char()
        ch = s.peek_char()
        if is_dec_digit(ch) {
            s.scan_dec()
            ch = s.peek_char()
        }
    }

    if ch == 'e' || ch == 'E' {
        s.next_char()
        ch = s.peek_char()
        if ch == '+' || ch == '-' {
            s.next_char()
            ch = s.peek_char()
        }
        if is_dec_digit(ch) {
            s.scan_dec()
            ch = s.peek_char()
        } else {
            s.next()
            return s.error("Invalid exponent"), ""
        }
    }

    if ch == 'i' {
        s.next_char()
        return IMAGINARY, s.src[start : s.off-1]
    } else {
        return FLOAT, s.src[start:s.off]
    }
}

// Scan a numeric literal.
func (s *Scanner) scan_numeric() (uint, string) {
    start := s.off
    r := s.peek_char()

    switch {
    case r == '0':
        s.next_char()
        r = s.peek_char()
        if r == 'x' || r == 'X' {
            // scan hexadecimal
            s.next_char()
            s.scan_hex()
            if s.off-start > 2 {
                return INTEGER, s.src[start:s.off]
            } else {
                return s.error("Invalid hex literal"), ""
            }
        } else {
            // scan octal or floating
            have_decimals := false
            s.scan_oct()
            r = s.peek_char()
            if r == '8' || r == '9' {
                have_decimals = true
                s.scan_dec()
                r = s.peek_char()
            }
            if r == 'i' {
                s.next_char()
                return IMAGINARY, s.src[start : s.off-1]
            } else if r == '.' || r == 'E' || r == 'e' {
                return s.scan_fract(start)
            } else if have_decimals {
                return s.error("Invalid octal literal"), ""
            } else {
                return INTEGER, s.src[start:s.off]
            }
        }

    case r == '.':
        return s.scan_fract(start)

    case is_dec_digit(r):
        s.scan_dec()
        r = s.peek_char()
        if r == '.' || r == 'e' || r == 'E' {
            return s.scan_fract(start)
        } else {
            return INTEGER, s.src[start:s.off]
        }

    default:
        panic("should not reach here")
    }
}

// Read N hexadecimal characters, return the value as a rune
func (s *Scanner) read_hex(n uint) rune {
    var i, r uint = 0, 0
    for ch := s.peek_char(); i < n && is_hex_digit(ch); i++ {
        r = r<<4 + hex_value(ch)
        s.next_char()
        ch = s.peek_char()
    }

    if i == n {
        return rune(r)
    } else {
        s.error("Invalid hex escape sequence")
        return utf8.RuneError
    }
}

// Read N octal characters, return the value as a rune
func (s *Scanner) read_oct(n uint) rune {
    var i, r uint = 0, 0
    for ch := s.peek_char(); i < n && is_oct_digit(ch); i++ {
        r = r<<3 + hex_value(ch)
        s.next_char()
        ch = s.peek_char()
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
func (s *Scanner) scan_escape(str bool) (r rune, b bool, ok bool) {
    ch := s.peek_char()

    if is_oct_digit(ch) {
        b = true
        r = s.read_oct(3)
        ok = r != utf8.RuneError
        return
    }

    s.next_char()
    b, ok = false, true
    switch {
    case ch == 'x':
        r, b = s.read_hex(2), true
        ok = r != utf8.RuneError
    case ch == 'u':
        r = s.read_hex(4)
    case ch == 'U':
        r = s.read_hex(8)
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
func (s *Scanner) scan_rune() (uint, string) {
    s.next_char() // skip leading apostrophe
    ch := s.peek_char()

    if ch == '\'' {
        s.next_char()
        return s.error("Empty rune literal"), ""
    }

    if ch == EOF {
        return s.error("EOF in rune literal"), ""
    }

    var r rune
    if ch == '\\' {
        s.next_char()
        rr, _, ok := s.scan_escape(false)
        if !ok {
            return ERROR, ""
        }
        r = rr
    } else {
        r = s.next()
    }

    ch = s.peek_char()
    if ch != '\'' {
        return s.error("Unterminated rune literal"), ""
    }

    s.next_char()
    return RUNE, string(r)
}

func (s *Scanner) scan_string() (uint, string) {
    s.next_char() // skip leading quotation mark

    var ch rune
    var str []byte
    for ch, _ = s.peek(); ch != '"' && ch != EOF && ch != NL; ch, _ = s.peek() {
        s.next()
        if ch == '\\' {
            c, b, ok := s.scan_escape(true)
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
        s.next_char()
        return STRING, string(str)
    } else {
        return s.error("Unterminated string literal"), ""
    }
}

func (s *Scanner) scan_raw_string() (uint, string) {
    s.next_char() // skip leading back apostrophe

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
        s.next_char()
        return STRING, string(str)
    } else {
        return s.error("Unterminated raw string literal"), ""
    }
}

// Main scanner routine. Get the next token.
func (s *Scanner) Get() uint {
    s.Err = nil

L:
    need_semicolon := s.need_semicolon

    if s.off == len(s.src) {
        if need_semicolon {
            s.need_semicolon = false
            return ';'
        } else {
            return EOF
        }
    }

    r, _ := s.peek()

    s.TOff, s.TLine, s.TPos = s.off, s.line, s.pos
    if is_whitespace(r) {
        s.next()
        if r == NL {
            if need_semicolon {
                s.need_semicolon = false
                return ';'
            }
        }
        goto L
    }

    s.need_semicolon = false

    switch r {
    case '+': // +, +=, ++
        s.next_char()
        r = s.peek_char()
        if r == '=' {
            s.next_char()
            return PLUS_ASSIGN
        } else if r == '+' {
            s.next_char()
            s.need_semicolon = true
            return INC
        }
        return '+'
    case '-': // -, -=, --
        s.next_char()
        r = s.peek_char()
        if r == '=' {
            s.next_char()
            return MINUS_ASSIGN
        } else if r == '-' {
            s.next_char()
            s.need_semicolon = true
            return DEC
        }
        return '-'
    case '*': // *, *=
        s.next_char()
        r = s.peek_char()
        if r == '=' {
            s.next_char()
            return MUL_ASSIGN
        }
        return '*'
    case '/': // /, /=, /*, //
        s.next_char()
        r = s.peek_char()
        if r == '=' {
            s.next_char()
            return DIV_ASSIGN
        } else if r == '/' {
            s.next_char()
            r = s.skip_line_comment()
            if r == EOF {
                return s.error("EOF in comment")
            }
            if r == NL && need_semicolon {
                return ';'
            }
            goto L
        } else if r == '*' {
            s.next_char()
            r = s.skip_block_comment()
            if r == EOF {
                return s.error("EOF in comment")
            } else if r == NL && need_semicolon {
                return ';'
            }
            goto L
        }
        return '/'
    case '%': // %, %=
        s.next_char()
        r = s.peek_char()
        if r == '=' {
            s.next_char()
            return REM_ASSIGN
        }
        return '%'
    case '&': // &, &=, &^, &^=, &&
        s.next_char()
        r = s.peek_char()
        if r == '=' {
            s.next_char()
            return AND_ASSIGN
        } else if r == '^' {
            s.next_char()
            r = s.peek_char()
            if r == '=' {
                s.next_char()
                return ANDN_ASSIGN
            }
            return ANDN
        } else if r == '&' {
            s.next_char()
            return AND
        }
        return '&'
    case '|': // |, |=, ||
        s.next_char()
        r = s.peek_char()
        if r == '=' {
            s.next_char()
            return OR_ASSIGN
        } else if r == '|' {
            s.next_char()
            return OR
        }
        return '|'
    case '^': // ^, ^=
        s.next_char()
        r = s.peek_char()
        if r == '=' {
            s.next_char()
            return XOR_ASSIGN
        }
        return '^'
    case '<': // <, <<, <<=, <=, <-
        s.next_char()
        r = s.peek_char()
        if r == '<' {
            s.next_char()
            r = s.peek_char()
            if r == '=' {
                s.next_char()
                return SHL_ASSIGN
            }
            return SHL
        } else if r == '=' {
            s.next_char()
            return LE
        } else if r == '-' {
            s.next_char()
            return RECV
        }
        return '<'
    case '>': // >, >>, >>=, >=
        s.next_char()
        r = s.peek_char()
        if r == '>' {
            s.next_char()
            r = s.peek_char()
            if r == '=' {
                s.next_char()
                return SHR_ASSIGN
            }
            return SHR
        } else if r == '=' {
            s.next_char()
            return GE
        }
        return '>'
    case '=': // =, ==
        s.next_char()
        r = s.peek_char()
        if r == '=' {
            s.next_char()
            return EQ
        }
        return '='
    case '!': // !, !=
        s.next_char()
        r = s.peek_char()
        if r == '=' {
            s.next_char()
            return NEQ
        }
        return '!'
    case '(', '[', '{', ',', ';':
        s.next_char()
        return uint(r)
    case ')', ']', '}':
        s.next_char()
        s.need_semicolon = true
        return uint(r)
    case '.': // ., ...
        s.next_char()
        r = s.peek_char()
        if r == '.' {
            s.next_char()
            r = s.peek_char()
            if r == '.' {
                s.next_char()
                return DOTS
            }
            s.off--
            s.pos--
        } else if is_dec_digit(r) {
            // Scan numeric literal
            s.off--
            t, v := s.scan_numeric()
            s.Value = v
            s.need_semicolon = true
            return t
        }
        return '.'
    case ':': // :, :=
        s.next_char()
        r = s.peek_char()
        if r == '=' {
            s.next_char()
            return DEFINE
        }
        return ':'
    case '\'':
        // Scan  rune literal
        t, v := s.scan_rune()
        if t == RUNE {
            s.Value = v
            s.need_semicolon = true
            return RUNE
        } else {
            return t
        }
    case '"':
        // Scan string literal
        t, v := s.scan_string()
        if t == STRING {
            s.Value = v
            s.need_semicolon = true
            return STRING
        } else {
            return t
        }
    case '`':
        // Scan raw string literal
        t, v := s.scan_raw_string()
        if t == STRING {
            s.Value = v
            s.need_semicolon = true
            return STRING
        } else {
            return t
        }
    }

    if is_letter(r) {
        id := s.scan_ident()
        kw, ok := keywords[id]
        if ok {
            switch kw {
            case BREAK, CONTINUE, FALLTHROUGH, RETURN:
                s.need_semicolon = true
            }
            return kw
        } else {
            s.Value = id
            s.need_semicolon = true
            return ID
        }
    } else if is_dec_digit(r) {
        // Scan numeric literal
        t, v := s.scan_numeric()
        s.need_semicolon = true
        s.Value = v
        return t
    } else {
        s.next()
        msg := fmt.Sprintf("Invalid character: |%c|", r)
        return s.error(msg)
    }
}

package scanner

import (
	"io/ioutil"
	"os"
	"testing"
)

func expectToken(t *testing.T, tok uint, exp uint) bool {
	if tok != exp {
		t.Errorf("Expected %s, got %s", TokenNames[exp], TokenNames[tok])
		return false
	} else {
		return true
	}
}

func TestEmpty(t *testing.T) {
	const tokens = ""

	s := New("basic tokens", tokens)
	if s.Get() != EOF {
		t.Fail()
	}
}

func testOps(t *testing.T, input string, tokens []uint) {
	s := New("ops", input)
	for i := range tokens {
		tok := s.Get()
		expectToken(t, tok, tokens[i])
	}
}

func TestOperations1(t *testing.T) {
	input := "+ += ++ - -= -- * *= / /= % %= & &= &^ &^= && | |= || ^ ^="
	tokens := [...]uint{'+', PLUS_ASSIGN, INC, '-', MINUS_ASSIGN, DEC, '*',
		MUL_ASSIGN, '/', DIV_ASSIGN, '%', REM_ASSIGN, '&', AND_ASSIGN, ANDN,
		ANDN_ASSIGN, AND, '|', OR_ASSIGN, OR, '^', XOR_ASSIGN}
	testOps(t, input, tokens[:])
}

func TestOperations2(t *testing.T) {
	input := "< <= << <<= <- > >> >>= >= = == ! != ( ) { } [ ] . ... .. : :="
	tokens := [...]uint{'<', LE, SHL, SHL_ASSIGN, RECV, '>', SHR, SHR_ASSIGN,
		GE, '=', EQ, '!', NE, '(', ')', '{', '}', '[', ']', '.', DOTS, '.',
		'.', ':', DEFINE}
	testOps(t, input, tokens[:])
}

func TestLineComment1(t *testing.T) {
	input := `// allalaa`

	s := New("line comments", input)
	tok := s.Get()
	if expectToken(t, tok, ERROR) {
		t.Log(s.Err)
	}
}

func TestLineComment2(t *testing.T) {
	input := `// allalaa
  a`

	s := New("line comments", input)
	tok := s.Get()
	expectToken(t, tok, LINE_COMMENT)
	tok = s.Get()
	expectToken(t, tok, ID)
}

func TestLineComment3(t *testing.T) {
	input := `a // allalaa
  b`

	s := New("line comments", input)
	tok := s.Get()
	expectToken(t, tok, ID)
	tok = s.Get()
	expectToken(t, tok, LINE_COMMENT)
	tok = s.Get()
	expectToken(t, tok, ';')
	tok = s.Get()
	expectToken(t, tok, ID)
}

func TestLineComment4(t *testing.T) {
	file, err := os.Open("test-invalid-encoding-1.txt")
	if err != nil {
		t.Error(err)
	}
	defer file.Close()
	input, err := ioutil.ReadAll(file)
	if err != nil {
		t.Error(err)
	}
	s := New("line comments", string(input))
	tok := s.Get()
	if expectToken(t, tok, ERROR) {
		t.Log(s.Err)
	}
}

func TestBlockComment1(t *testing.T) {
	input := `a /* allalaa`

	s := New("block comments", input)
	tok := s.Get()
	expectToken(t, tok, ID)
	tok = s.Get()
	if expectToken(t, tok, ERROR) {
		t.Log(s.Err)
	}
}

func TestBlockComment2(t *testing.T) {
	input := `/* allalaa
    alala`

	s := New("block comments", input)
	tok := s.Get()
	if expectToken(t, tok, ERROR) {
		t.Log(s.Err)
	}
}

func TestBlockComment3(t *testing.T) {
	input := `/* allalaa
    alala */`

	s := New("block comments", input)
	tok := s.Get()
	expectToken(t, tok, BLOCK_COMMENT)
	tok = s.Get()
	expectToken(t, tok, EOF)
}

func TestBlockComment4(t *testing.T) {
	input := `a + /* allalaa
    alala */ b`

	s := New("block comments", input)
	tokens := [...]uint{ID, '+', BLOCK_COMMENT, ID, ';', EOF}
	for _, exp := range tokens {
		tok := s.Get()
		expectToken(t, tok, exp)
	}
}

func TestBlockComment5(t *testing.T) {
	input := `a + /* allalaaalala */ b`

	s := New("block comments", input)
	tokens := [...]uint{ID, '+', BLOCK_COMMENT, ID, ';', EOF}
	for _, exp := range tokens {
		tok := s.Get()
		expectToken(t, tok, exp)
	}
}

func TestBlockComment6(t *testing.T) {
	input := `a /* allalaa
    alala */ b`

	s := New("block comments", input)
	tokens := [...]uint{ID, BLOCK_COMMENT, ';', ID, ';', EOF}
	for _, exp := range tokens {
		tok := s.Get()
		expectToken(t, tok, exp)
	}
}

func TestBlockComment7(t *testing.T) {
	input := `a /* x
*//* y
*/b`

	s := New("block comments", input)
	tokens := [...]uint{ID, BLOCK_COMMENT, ';', BLOCK_COMMENT, ID, ';', EOF}
	for _, exp := range tokens {
		tok := s.Get()
		expectToken(t, tok, exp)
	}
}

func TestBlockComment8(t *testing.T) {
	file, err := os.Open("test-invalid-encoding.txt")
	if err != nil {
		t.Error(err)
	}
	defer file.Close()
	input, err := ioutil.ReadAll(file)
	if err != nil {
		t.Error(err)
	}
	s := New("block comments", string(input))
	tok := s.Get()
	if expectToken(t, tok, ERROR) {
		t.Log(s.Err)
	}
}

func TestSemicolon1(t *testing.T) {
	const tokens = `a
    +
    12
    12.0
    0i
    'γ'
    "zz"
    )
    ]
    }
   ++
   --
   break
   continue
   fallthrough
   return
    `

	var expect = []uint{ID, ';', '+', INTEGER, ';', FLOAT, ';', IMAGINARY, ';',
		RUNE, ';', STRING, ';', ')', ';', ']', ';', '}', ';', INC, ';', DEC,
		';', BREAK, ';', CONTINUE, ';', FALLTHROUGH, ';', RETURN, ';',
	}

	s := New("semicolons", tokens)
	for _, exp := range expect {
		tok := s.Get()
		expectToken(t, tok, exp)
	}

	tok := s.Get()
	expectToken(t, tok, EOF)
}

func testString(t *testing.T, input string, value string) {
	s := New("strings", input)
	tok := s.Get()
	if tok != STRING {
		t.Errorf("Expected string, got %s", TokenNames[tok])
	}
	str := s.Value
	tok = s.Get()
	if tok != ';' {
		t.Errorf("Expected ';', got %s", TokenNames[tok])
	}
	tok = s.Get()
	if tok != EOF {
		t.Errorf("Expected EOF, got %s", TokenNames[tok])
	}

	if string(str) != value {
		t.Errorf("Incorrect string value, expected |%s|, got |%s|", value, str)
	}
}

func TestStringBasic(t *testing.T) {
	input := [...]string{`""`, `"\t"`, `"abc"`, `"ab\taßäख🃑\r\txy"`}
	for i := range input {
		testString(t, input[i], input[i])
	}
}

func TestStringEscapes(t *testing.T) {
	const input = `"\041 \x20 \u0916 \U0001F0D1 ' \""`
	testString(t, input, input)
}

func TestStringErrors(t *testing.T) {
	input := [...]string{`"a`, `"lala
    lala"`, `"\xyz"`}

	for _, in := range input {
		s := New("string escape errors", in)
		tok := s.Get()
		if expectToken(t, tok, ERROR) {
			t.Log(s.Err)
		}
	}
}

func testRunes(t *testing.T, runes string, values []string) {
	s := New("runes", runes)

	tok := s.Get()
	for _, val := range values {
		if tok != RUNE {
			t.Errorf("Expected rune, got %s", TokenNames[tok])
			if tok == ERROR {
				t.Errorf("error is: %s", s.Err)
			}
		} else {
			if string(s.Value) != val {
				t.Errorf("Expected value |%s|, got |%s|", val, s.Value)
			}
		}

		tok = s.Get()
	}
}

func TestRunesBasic(t *testing.T) {
	const runes = `' ' 'a' 'ß' 'ä' 'ख' '🃑'`
	values := [...]string{" ", "a", "ß", "ä", "ख", "🃑"}

	testRunes(t, runes, values[:])
}

func TestRunesHexOctEscapes(t *testing.T) {
	runes := `'\041' '\x20' '\u0916' '\U0001F0d1'`
	values := [...]string{`\041`, `\x20`, `\u0916`, `\U0001F0d1`}
	testRunes(t, runes, values[:])
}

func TestRunesEscapes(t *testing.T) {
	const runes = ` '\a' '\b' '\f' '\n' '\r' '\t' '\v' '\\' '\'' `
	values := [...]string{`\a`, `\b`, `\f`, `\n`, `\r`, `\t`, `\v`, `\\`, `\'`}
	// \"   U+0022 double quote  (valid escape only within string literals)
	testRunes(t, runes, values[:])
}

func TestRuneEscapeErrors(t *testing.T) {
	input := [...]string{`'\049'`, `'\u091g'`, `'\w0916'`, `''`, `'`, `'a `}

	for _, in := range input {
		s := New("rune escape errors", in)
		tok := s.Get()
		if expectToken(t, tok, ERROR) {
			t.Log(s.Err)
		}
	}
}

func TestRuneEncodeError(t *testing.T) {
	input := []byte{'\'', 0x81, '\''}
	s := New("rune escape errors", string(input))
	tok := s.Get()
	if expectToken(t, tok, ERROR) {
		t.Log(s.Err)
	}
}

func TestIntLiterals1(t *testing.T) {
	input := "0x0 0x0123456789abcdefABCDEF 0xy 0x"
	values := [...]string{"0x0", "0x0123456789abcdefABCDEF"}

	s := New("hex numeric", input)
	tok := s.Get()
	expectToken(t, tok, INTEGER)
	if string(s.Value) != values[0] {
		t.Errorf("Expected token value %s, got %s", values[0], s.Value)
	}

	tok = s.Get()
	expectToken(t, tok, INTEGER)
	if string(s.Value) != values[1] {
		t.Errorf("Expected token value %s, got %s", values[0], s.Value)
	}

	tok = s.Get()
	if expectToken(t, tok, ERROR) {
		t.Log(s.Err)
	}

	tok = s.Get()
	expectToken(t, tok, ID)

	tok = s.Get()
	if expectToken(t, tok, ERROR) {
		t.Log(s.Err)
	}

	tok = s.Get()
	expectToken(t, tok, ';')

	tok = s.Get()
	expectToken(t, tok, EOF)
}

func TestIntLiterals2(t *testing.T) {
	input := "01 01234567 0a 01289"
	values := [...]string{"01", "01234567", "0", "a"}

	s := New("oct numeric", input)
	tok := s.Get()
	expectToken(t, tok, INTEGER)
	if string(s.Value) != values[0] {
		t.Errorf("Expected token value %s, got %s", values[0], s.Value)
	}

	tok = s.Get()
	expectToken(t, tok, INTEGER)
	if string(s.Value) != values[1] {
		t.Errorf("Expected token value %s, got %s", values[1], s.Value)
	}

	tok = s.Get()
	expectToken(t, tok, INTEGER)
	if string(s.Value) != values[2] {
		t.Errorf("Expected token value %s, got %s", values[2], s.Value)
	}

	tok = s.Get()
	expectToken(t, tok, ID)

	tok = s.Get()
	if expectToken(t, tok, ERROR) {
		t.Log(s.Err)
	}

	tok = s.Get()
	expectToken(t, tok, ';')

	tok = s.Get()
	expectToken(t, tok, EOF)
}

func TestIntLiterals3(t *testing.T) {
	input := "123456789a"
	values := [...]string{"123456789", "a"}

	s := New("int", input)
	tok := s.Get()
	expectToken(t, tok, INTEGER)
	if string(s.Value) != values[0] {
		t.Errorf("Expected token value %s, got %s", values[0], s.Value)
	}

	tok = s.Get()
	expectToken(t, tok, ID)

	tok = s.Get()
	expectToken(t, tok, ';')

	tok = s.Get()
	expectToken(t, tok, EOF)
}

func TestFloatLiterals(t *testing.T) {
	input := ".1 .1e+12 .1e12 .12e-12 1.1 012.1e+12 019.1e12 910.12e-12"
	values := [...]string{".1", ".1e+12", ".1e12", ".12e-12", "1.1",
		"012.1e+12", "019.1e12", "910.12e-12"}

	s := New("float", input)
	for i := range values {
		tok := s.Get()
		expectToken(t, tok, FLOAT)
		if string(s.Value) != values[i] {
			t.Errorf("Expected token value %s, got %s", values[i], s.Value)
		}
	}

	tok := s.Get()
	expectToken(t, tok, ';')

	tok = s.Get()
	expectToken(t, tok, EOF)
}

func TestFloatLiteralErrors(t *testing.T) {
	input := ".1eΣ"

	s := New("float", input)
	tokens := [...]uint{ERROR, ID, ';', EOF}
	for _, exp := range tokens {
		tok := s.Get()
		expectToken(t, tok, exp)
		if tok == ERROR {
			t.Log(s.Err)
		}
	}
}

func TestImaginaryLiterals(t *testing.T) {
	input := "0i 0129i .1i .1e+12i .1e12i .12e-12i 1.1i 012.1e+12i 019.1e12i 910.12e-12i"
	values := [...]string{"0", "0129", ".1", ".1e+12", ".1e12", ".12e-12",
		"1.1", "012.1e+12", "019.1e12", "910.12e-12"}

	s := New("float", input)
	for i := range values {
		tok := s.Get()
		expectToken(t, tok, IMAGINARY)
		if string(s.Value) != values[i] {
			t.Errorf("Expected token value %s, got %s", values[i], s.Value)
		}
	}

	tok := s.Get()
	expectToken(t, tok, ';')

	tok = s.Get()
	expectToken(t, tok, EOF)
}

func TestKeywords(t *testing.T) {
	input := "break case chan const continue default defer else fallthrough for func go goto if import interface map package range return select struct switch type var"
	tokens := [...]uint{BREAK, CASE, CHAN, CONST, CONTINUE, DEFAULT, DEFER, ELSE,
		FALLTHROUGH, FOR, FUNC, GO, GOTO, IF, IMPORT, INTERFACE, MAP, PACKAGE,
		RANGE, RETURN, SELECT, STRUCT, SWITCH, TYPE, VAR}

	s := New("keywords", input)
	for _, exp := range tokens {
		tok := s.Get()
		expectToken(t, tok, exp)
	}

	tok := s.Get()
	expectToken(t, tok, EOF)
}

func TestRawString(t *testing.T) {
	file, err := os.Open("test-raw-literal.txt")
	if err != nil {
		t.Error(err)
	}

	src, err := ioutil.ReadAll(file)
	if err != nil {
		t.Error(err)
	}
	t.Log("src=", src)
	s := New("raw strings", string(src))
	tok := s.Get()
	if !expectToken(t, tok, STRING) && tok == ERROR {
		t.Error(s.Err)
	}

	const value = "`a\nb\nc`"

	if string(s.Value) != value {
		t.Errorf("Expected value ->|%s|<-\ngot ->|%s|<- ", value, s.Value)
	}

	tok = s.Get()
	if expectToken(t, tok, ERROR) {
		t.Log(s.Err)
	}

	file.Close()
}

func TestIdent(t *testing.T) {
	input := "_ a _a Ab_C φγ ڦ Պէ абгдежЗИЙКлмноп םשׁ ばぽ 괇괚"
	values := [...]string{"_", "a", "_a", "Ab_C", "φγ", "ڦ", "Պէ",
		"абгдежЗИЙКлмноп", "םשׁ", "ばぽ", "괇괚"}

	s := New("idents", input)
	for _, exp := range values {
		tok := s.Get()
		expectToken(t, tok, ID)
		if string(s.Value) != exp {
			t.Errorf("Expected value ->|%s|<-, got ->|%s|<- ", exp, s.Value)
		}
	}

	tok := s.Get()
	expectToken(t, tok, ';')

	tok = s.Get()
	expectToken(t, tok, EOF)
}

// String and rune literals may contain Unicode replacement char, a.k.a
// RuneError
func TestRegression20140119143717(t *testing.T) {
	input := `'\uFFFD'` // correctly encoded RuneError

	s := New("regression20140119", input)
	tok := s.Get()
	if !expectToken(t, tok, RUNE) {
		t.Log(s.Err)
	}

	tok = s.Get()
	expectToken(t, tok, ';')

	tok = s.Get()
	expectToken(t, tok, EOF)
}

func TestRegession20140119165305(t *testing.T) {
	input := `/***/`
	s := New("regression20140119", input)
	tok := s.Get()
	expectToken(t, tok, BLOCK_COMMENT)
	tok = s.Get()
	expectToken(t, tok, EOF)
}

func TestRegression20140121222305(t *testing.T) {
	input := `x⊛y`
	s := New("regression20140121", input)
	tok := s.Get()
	expectToken(t, tok, ID)

	tok = s.Get()
	if expectToken(t, tok, ERROR) {
		t.Log(s.Err)
	}
}

func TestBug20150722T232927(t *testing.T) {
	src := `x /* */
            +1`
	exp := []uint{ID, BLOCK_COMMENT, ';', '+', INTEGER, ';'}
	s := New("Bug20150722T232927.go", src)
	for i := range exp {
		tok := s.Get()
		expectToken(t, tok, exp[i])
	}
	tok := s.Get()
	expectToken(t, tok, EOF)
}

func TestBug20150725T222756(t *testing.T) {
	src := `1i 11i 0i`
	exp := []uint{IMAGINARY, IMAGINARY, IMAGINARY, ';'}
	s := New("Bug20150725T222756.go", src)
	for i := range exp {
		tok := s.Get()
		expectToken(t, tok, exp[i])
	}
	tok := s.Get()
	expectToken(t, tok, EOF)
}

func TestSourceMap(t *testing.T) {
	src := `a b /* alala */
c
// alalal
d e
// alalalal
// fofofof
f
`
	offs := []int{0, 2, 4, 4, 16, 16, 18, 28, 30, 30, 32, 44, 55, 55}
	s := New("test-position.go", src)
	i := 0
	tok := s.Get()
	for tok != ERROR && tok != EOF {
		if s.TOff != offs[i] {
			t.Error("wrong token offset: token:", TokenNames[tok], "off:", s.TOff)
		}
		tok = s.Get()
		i++
	}

	off := 0
	for i := 0; i < s.SrcMap.LineCount(); i++ {
		o, len := s.SrcMap.LineExtent(i)
		if off != o {
			t.Error("line ", i+1, "expected at offset", off, "got", o, "instead")
		}
		off += len
	}
	if off != s.SrcMap.Size() {
		t.Error("expected source size", off, "got", s.SrcMap.Size(), "instead")
	}
}

func TestBug20150809T112324(t *testing.T) {
	src := `package main`
	s := New("test-position.go", src)
	tok := s.Get()
	for tok != EOF {
		tok = s.Get()
	}

	if len(s.SrcMap.line) != 1 {
		t.Error("wrong count of lines in source map")
	}
}

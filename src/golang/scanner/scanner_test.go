package scanner

import (
	"io/ioutil"
	"os"
	"testing"
)

func expect_token(t *testing.T, tok uint, exp uint) bool {
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

func test_ops(t *testing.T, input string, tokens []uint) {
	s := New("ops", input)
	for i, _ := range tokens {
		tok := s.Get()
		expect_token(t, tok, tokens[i])
	}
}

func TestOperations1(t *testing.T) {
	input := "+ += ++ - -= -- * *= / /= % %= & &= &^ &^= && | |= || ^ ^="
	tokens := [...]uint{'+', PLUS_ASSIGN, INC, '-', MINUS_ASSIGN, DEC, '*',
		MUL_ASSIGN, '/', DIV_ASSIGN, '%', REM_ASSIGN, '&', AND_ASSIGN, ANDN,
		ANDN_ASSIGN, AND, '|', OR_ASSIGN, OR, '^', XOR_ASSIGN}
	test_ops(t, input, tokens[:])
}

func TestOperations2(t *testing.T) {
	input := "< <= << <<= <- > >> >>= >= = == ! != ( ) { } [ ] . ... .. : :="
	tokens := [...]uint{'<', LE, SHL, SHL_ASSIGN, RECV, '>', SHR, SHR_ASSIGN,
		GE, '=', EQ, '!', NEQ, '(', ')', '{', '}', '[', ']', '.', DOTS, '.',
		'.', ':', DEFINE}
	test_ops(t, input, tokens[:])
}
func TestLineComment1(t *testing.T) {
	input := `// allalaa`

	s := New("line comments", input)
	tok := s.Get()
	if expect_token(t, tok, ERROR) {
		t.Log(s.Err)
	}
}

func TestLineComment2(t *testing.T) {
	input := `// allalaa
  a`

	s := New("line comments", input)
	tok := s.Get()
	expect_token(t, tok, ID)
}

func TestLineComment3(t *testing.T) {
	input := `a // allalaa
  b`

	s := New("line comments", input)
	tok := s.Get()
	expect_token(t, tok, ID)

	tok = s.Get()
	expect_token(t, tok, ';')

	tok = s.Get()
	expect_token(t, tok, ID)
}

func TestBlockComment1(t *testing.T) {
	input := `a /* allalaa`

	s := New("block comments", input)
	tok := s.Get()
	expect_token(t, tok, ID)
	tok = s.Get()
	if expect_token(t, tok, ERROR) {
		t.Log(s.Err)
	}
}

func TestBlockComment2(t *testing.T) {
	input := `/* allalaa
    alala`

	s := New("block comments", input)
	tok := s.Get()
	if expect_token(t, tok, ERROR) {
		t.Log(s.Err)
	}
}

func TestBlockComment3(t *testing.T) {
	input := `/* allalaa
    alala */`

	s := New("block comments", input)
	tok := s.Get()
	expect_token(t, tok, EOF)
}

func TestBlockComment4(t *testing.T) {
	input := `a + /* allalaa
    alala */ b`

	s := New("block comments", input)
	tok := s.Get()
	expect_token(t, tok, ID)
	tok = s.Get()
	expect_token(t, tok, '+')
	tok = s.Get()
	expect_token(t, tok, ID)
	tok = s.Get()
	expect_token(t, tok, EOF)
}

func TestBlockComment5(t *testing.T) {
	input := `a + /* allalaaalala */ b`

	s := New("block comments", input)
	tok := s.Get()
	expect_token(t, tok, ID)
	tok = s.Get()
	expect_token(t, tok, '+')
	tok = s.Get()
	expect_token(t, tok, ID)
	tok = s.Get()
	expect_token(t, tok, EOF)
}

func TestBlockComment6(t *testing.T) {
	input := `a /* allalaa
    alala */ b`

	s := New("block comments", input)
	tok := s.Get()
	expect_token(t, tok, ID)
	tok = s.Get()
	expect_token(t, tok, ';')
	tok = s.Get()
	expect_token(t, tok, ID)
	tok = s.Get()
	expect_token(t, tok, EOF)
}

// func TestBasicTokens1(t *testing.T) {
// 	// const tokens = `> <- ßeta _
// 	//    ;`
// 	const tokens = `> <- ßeta _
//     ;`
// 	var expect = []uint{GT, RECV, ID, ID, ';', ';'}

// 	s := New("basic tokens", tokens)
// 	for i, tok := 0, s.Get(); tok != EOF; i++ {
// 		if i >= len(expect) {
// 			t.Errorf("EOF not found, found token %s instead", TokenNames[tok])
// 		} else if tok != expect[i] {
// 			t.Errorf("failed to recognize token '%s'", TokenNames[expect[i]])
// 		}
// 		tok = s.Get()
// 	}
// }

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
		expect_token(t, tok, exp)
	}

	tok := s.Get()
	expect_token(t, tok, EOF)
}

func test_string(t *testing.T, input string, value string) {
	s := New("strings", input)
	tok := s.Get()
	if tok != STRING {
		t.Errorf("Expected string, got %s", TokenNames[tok])
	}
	str := s.Value
	tok = s.Get()
	if tok != EOF {
		t.Errorf("Expected EOF, got %s", TokenNames[tok])
	}

	if str != value {
		t.Errorf("Incorrect string value, expected |%s|, got |%s|", value, str)
	}
}

func TestStringBasic(t *testing.T) {
	input := [...]string{`""`, `"\t"`, `"abc"`, `"ab\taßäख🃑\r\txy"`}
	values := [...]string{"", "\t", "abc", "ab\taßäख🃑\r\txy"}
	for i, _ := range input {
		test_string(t, input[i], values[i])
	}
}

func TestStringEscapes(t *testing.T) {
	const input = `"\041 \x20 \u0916 \U0001F0D1 ' \""`
	value := "! \x20 ख 🃑 ' \""
	test_string(t, input, value)
}

func TestStringErrors(t *testing.T) {
	input := [...]string{`"a`, `"lala
    lala"`, `"\xyz"`}

	for _, in := range input {
		s := New("string escape errors", in)
		tok := s.Get()
		if expect_token(t, tok, ERROR) {
			t.Log(s.Err)
		}
	}
}

func test_runes(t *testing.T, runes string, values []string) {
	s := New("runes", runes)

	for i, tok := 0, s.Get(); tok != EOF; i++ {
		if i >= len(values) {
			t.Fatalf("EOF not found, list of values exhausted")
		}

		if tok != RUNE {
			t.Errorf("Expected rune, got %s", TokenNames[tok])
			if tok == ERROR {
				t.Errorf("error is: %s", s.Err)
			}
		} else {
			if s.Value != values[i] {
				t.Errorf("Expected value |%s|, got |%s|", values[i], s.Value)
			}
		}

		tok = s.Get()
	}
}

func TestRunesBasic(t *testing.T) {
	const runes = `' ' 'a' 'ß' 'ä' 'ख' '🃑'`
	values := [...]string{" ", "a", "ß", "ä", "ख", "🃑"}

	test_runes(t, runes, values[:])
}

func TestRunesHexOctEscapes(t *testing.T) {
	const runes = `'\041' '\x20' '\u0916' '\U0001F0d1'`
	values := [...]string{"!", " ", "ख", "🃑"}

	test_runes(t, runes, values[:])
}

func TestRunesEscapes(t *testing.T) {
	const runes = ` '\a' '\b' '\f' '\n' '\r' '\t' '\v' '\\' '\'' `
	values := [...]string{"\a", "\b", "\f", "\n", "\r", "\t", "\v", "\\", "'"}

	// \"   U+0022 double quote  (valid escape only within string literals)

	test_runes(t, runes, values[:])
}

func TestRuneEscapeErrors(t *testing.T) {
	input := [...]string{`'\049'`, `'\u091g'`, `'\w0916'`, `''`, `'`, `'a `}

	for _, in := range input {
		s := New("rune escape errors", in)
		tok := s.Get()
		if expect_token(t, tok, ERROR) {
			t.Log(s.Err)
		}
	}
}

func TestIntLiterals1(t *testing.T) {
	input := "0x0 0x0123456789abcdefABCDEF 0xy 0x"
	values := [...]string{"0x0", "0x0123456789abcdefABCDEF"}

	s := New("hex numeric", input)
	tok := s.Get()
	expect_token(t, tok, INTEGER)
	if s.Value != values[0] {
		t.Errorf("Expected token value %s, got %s", values[0], s.Value)
	}

	tok = s.Get()
	expect_token(t, tok, INTEGER)
	if s.Value != values[1] {
		t.Errorf("Expected token value %s, got %s", values[0], s.Value)
	}

	tok = s.Get()
	if expect_token(t, tok, ERROR) {
		t.Log(s.Err)
	}

	tok = s.Get()
	expect_token(t, tok, ID)

	tok = s.Get()
	if expect_token(t, tok, ERROR) {
		t.Log(s.Err)
	}

	tok = s.Get()
	expect_token(t, tok, EOF)
}

func TestIntLiterals2(t *testing.T) {
	input := "01 01234567 0a 01289"
	values := [...]string{"01", "01234567", "0", "a"}

	s := New("oct numeric", input)
	tok := s.Get()
	expect_token(t, tok, INTEGER)
	if s.Value != values[0] {
		t.Errorf("Expected token value %s, got %s", values[0], s.Value)
	}

	tok = s.Get()
	expect_token(t, tok, INTEGER)
	if s.Value != values[1] {
		t.Errorf("Expected token value %s, got %s", values[1], s.Value)
	}

	tok = s.Get()
	expect_token(t, tok, INTEGER)
	if s.Value != values[2] {
		t.Errorf("Expected token value %s, got %s", values[2], s.Value)
	}

	tok = s.Get()
	expect_token(t, tok, ID)

	tok = s.Get()
	if expect_token(t, tok, ERROR) {
		t.Log(s.Err)
	}

	tok = s.Get()
	expect_token(t, tok, EOF)
}

func TestIntLiterals3(t *testing.T) {
	input := "123456789a"
	values := [...]string{"123456789", "a"}

	s := New("int", input)
	tok := s.Get()
	expect_token(t, tok, INTEGER)
	if s.Value != values[0] {
		t.Errorf("Expected token value %s, got %s", values[0], s.Value)
	}

	tok = s.Get()
	expect_token(t, tok, ID)

	tok = s.Get()
	expect_token(t, tok, EOF)
}

func TestFloatLiterals(t *testing.T) {
	input := ".1 .1e+12 .1e12 .12e-12 1.1 012.1e+12 019.1e12 910.12e-12"
	values := [...]string{".1", ".1e+12", ".1e12", ".12e-12", "1.1",
		"012.1e+12", "019.1e12", "910.12e-12"}

	s := New("float", input)
	for i, _ := range values {
		tok := s.Get()
		expect_token(t, tok, FLOAT)
		if s.Value != values[i] {
			t.Errorf("Expected token value %s, got %s", values[i], s.Value)
		}
	}

	tok := s.Get()
	expect_token(t, tok, EOF)
}

func TestFloatLiteralErrors(t *testing.T) {
	input := ".1eΣ"

	s := New("float", input)
	for tok := s.Get(); tok != EOF; tok = s.Get() {
		if expect_token(t, tok, ERROR) {
			t.Log(s.Err)
		}
	}
}

func TestImaginaryLiterals(t *testing.T) {
	input := "0i 0129i .1i .1e+12i .1e12i .12e-12i 1.1i 012.1e+12i 019.1e12i 910.12e-12i"
	values := [...]string{"0", "0129", ".1", ".1e+12", ".1e12", ".12e-12",
		"1.1", "012.1e+12", "019.1e12", "910.12e-12"}

	s := New("float", input)
	for i, _ := range values {
		tok := s.Get()
		expect_token(t, tok, IMAGINARY)
		if s.Value != values[i] {
			t.Errorf("Expected token value %s, got %s", values[i], s.Value)
		}
	}

	tok := s.Get()
	expect_token(t, tok, EOF)
}

func TestKeywords(t *testing.T) {
	input := "break case chan const continue default defer else fallthrough for func go goto if import interface map package range return select struct switch type var"
	tokens := [...]uint{BREAK, CASE, CHAN, CONST, CONTINUE, DEFAULT, DEFER, ELSE,
		FALLTHROUGH, FOR, FUNC, GO, GOTO, IF, IMPORT, INTERFACE, MAP, PACKAGE,
		RANGE, RETURN, SELECT, STRUCT, SWITCH, TYPE, VAR}

	s := New("keywords", input)
	for _, exp := range tokens {
		tok := s.Get()
		expect_token(t, tok, exp)
	}

	tok := s.Get()
	expect_token(t, tok, EOF)
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
	if !expect_token(t, tok, STRING) && tok == ERROR {
		t.Error(s.Err)
	}

	const value = `a
b
c`

	if s.Value != value {
		t.Errorf("Expected value ->|%s|<-, got ->|%s|<- ", value, s.Value)
	}

	tok = s.Get()
	if expect_token(t, tok, ERROR) {
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
		expect_token(t, tok, ID)
		if s.Value != exp {
			t.Errorf("Expected value ->|%s|<-, got ->|%s|<- ", exp, s.Value)
		}
	}

	tok := s.Get()
	expect_token(t, tok, EOF)
}

func TestRegression20140119143717(t *testing.T) {
	input := `'\uFFFD'` // correctly encoded RuneError

	s := New("regression20140119", input)
	tok := s.Get()
	if !expect_token(t, tok, RUNE) {
		t.Log(s.Err)
	}

	tok = s.Get()
	expect_token(t, tok, EOF)
}

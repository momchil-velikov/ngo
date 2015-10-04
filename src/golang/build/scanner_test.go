package build

import (
	"bytes"
	"strings"
	"testing"
)

func testScanner(src string) scanner {
	s := scanner{}
	s.init("scanner-test.go", bytes.NewBufferString(src))
	return s
}

func testTokenSequence(t *testing.T, src string, exp []uint) {
	s := testScanner(src)
	tokens := []token{}
	token, _ := s.Get()
	for token.Kind != tEOF && token.Kind != tERROR {
		tokens = append(tokens, token)
		token, _ = s.Get()
	}
	if len(tokens) != len(exp) {
		t.Errorf("incorrected number of tokens: %d, shoild be %d", len(tokens), len(exp))
	}
	for i := range exp {
		if tokens[i].Kind != exp[i] {
			t.Error("incorrect token received:", tokens[i].String())
		}
	}
}

func TestScanner(t *testing.T) {
	testTokenSequence(
		t,
		"().;package a \n import \"a\" /* xx */ // foo",
		[]uint{'(', ')', '.', ';', tPACKAGE, tID, ';', tIMPORT, tSTRING, tBLOCK_COMMENT,
			tLINE_COMMENT, ';'},
	)
}

func TestScannerInvalidToken(t *testing.T) {
	s := testScanner("/ a")
	token, err := s.Get()
	if token.Kind != tERROR {
		t.Error("expected error token")
	}
	if err == nil || !strings.Contains(err.Error(), "invalid token") {
		t.Error("expected invalid token error")
	}

	s = testScanner("* a")
	token, err = s.Get()
	if token.Kind != tERROR {
		t.Error("expected error token")
	}
	if err == nil || !strings.Contains(err.Error(), "invalid source character") {
		t.Error("expected invalid source character error")
	}
}

func TestScannerString(t *testing.T) {
	s := testScanner(`"a\142\x63\u0064\U00000065\xE4\u00e4\a\b\f\n\r\t\v\\\""`)
	token, err := s.Get()
	if err != nil {
		t.Fatal(err)
	}
	if token.Kind != tSTRING {
		t.Error("unexpected token", token.String())
	}
	if token.Value != "abcde\xe4ä\a\b\f\n\r\t\v\\\"" {
		t.Error("unexpected token value:", token.Value)
	}
}

func TestScannerUnterminatedString(t *testing.T) {
	s := testScanner(`"a`)
	token, err := s.Get()
	if token.Kind != tERROR {
		t.Error("unexpected token", token.String())
	}
	if err == nil || !strings.Contains(err.Error(), "unterminated string literal") {
		t.Error("expected unterminated string literal error")
	}

	s = testScanner(`"a
`)
	token, err = s.Get()
	if token.Kind != tERROR {
		t.Error("unexpected token", token.String())
	}
	if err == nil || !strings.Contains(err.Error(), "unterminated string literal") {
		t.Error("expected unterminated string literal error")
	}
}

func TestScannerRawString(t *testing.T) {
	buf := []byte{'`', 'a', '\\', 'n', '\n', '\r', 'b', '`'}
	s := testScanner(string(buf))
	token, err := s.Get()
	if err != nil {
		t.Error(err)
	}
	if token.Kind != tSTRING {
		t.Error("unexpected token", token.String())
	}
	if token.Value != "a\\n\nb" {
		t.Error("unexpected token value:", token.Value)
	}
}

func TestScannerUnterminatedRawString(t *testing.T) {
	buf := []byte{'`', 'a'}
	s := testScanner(string(buf))
	token, err := s.Get()
	if token.Kind != tERROR {
		t.Error("unexpected token", token.String())
	}
	if err == nil || !strings.Contains(err.Error(), "unterminated raw string literal") {
		t.Error("expected unterminated raw string literal error")
	}

	s = testScanner(`"a
`)
	token, err = s.Get()
	if token.Kind != tERROR {
		t.Error("unexpected token", token.String())
	}
	if err == nil || !strings.Contains(err.Error(), "unterminated string literal") {
		t.Error("expected unterminated string literal error")
	}
}

func TestScannerBlockComment(t *testing.T) {
	testTokenSequence(
		t,
		`a /* ccc * ddd
*/ b`,
		[]uint{tID, tBLOCK_COMMENT, ';', tID, ';'},
	)
}

func TestScannerEOFInComment(t *testing.T) {
	s := testScanner(`/*`)
	_, err := s.Get()
	if err == nil || !strings.Contains(err.Error(), "EOF in comment") {
		t.Error("expected EOF-in-comment error")
	}
}

func TestScannerInvalidEscape(t *testing.T) {
	s := testScanner(`"\0z"`)
	_, err := s.Get()
	if err == nil || !strings.Contains(err.Error(), "invalid octal escape") {
		t.Error("expected invalid octal sequence error", err)
	}

	s = testScanner(`"\xz"`)
	_, err = s.Get()
	if err == nil || !strings.Contains(err.Error(), "invalid hex escape") {
		t.Error("expected invalid hex sequence error", err)
	}

	s = testScanner(`"\z"`)
	_, err = s.Get()
	if err == nil || !strings.Contains(err.Error(), "invalid escape") {
		t.Error("expected invalid escape sequence error", err)
	}
}

func TestScannerInvalidEncoding(t *testing.T) {
	buf := []byte{'/', '/', 224, 184, '\n'}
	s := testScanner(string(buf))
	token, err := s.Get()
	if token.Kind != tERROR {
		t.Error("expected error token")
	}
	if err == nil || !strings.Contains(err.Error(), "invalid UTF8") {
		t.Error("expected invalid UTF* encoding")
	}

	buf = []byte{224, 184, '\n'}
	s = testScanner(string(buf))
	token, err = s.Get()
	if token.Kind != tERROR {
		t.Error("expected error token")
	}
	if err == nil || !strings.Contains(err.Error(), "invalid UTF8") {
		t.Error("expected invalid UTF* encoding")
	}
}

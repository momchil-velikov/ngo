package build

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testParseFile(src string) (*File, error) {
	return parseFile("test-parser.go", bytes.NewBufferString(src))
}

func TestParser(t *testing.T) {
	src := `// comment 1
// comment 2
package a
// comment 3
import "b" /* comment 4 */
import (. "c"; _ "d"
        "e")
`
	f, err := testParseFile(src)
	if err != nil {
		t.Fatal(err)
	}
	if f.Package != "a" {
		t.Error("incorrect package name:", f.Package)
	}
	if len(f.Imports) != 4 {
		t.Error("must have exactly 4 imports")
	}
	is := []string{"b", "c", "d", "e"}
	for i := range is {
		if f.Imports[i] != is[i] {
			t.Errorf("incorrect import #%d: must be %s, but is %s",
				i, is[i], f.Imports[i])
		}
	}
}

func TestParserErrors(t *testing.T) {
	_, err := testParseFile("package a b")
	if err == nil || !strings.Contains(err.Error(), "newline to follow package") {
		t.Error("expected missing semicolon or newline after package clause error")
	}

	_, err = testParseFile(`package`)
	if err == nil || !strings.Contains(err.Error(), "followed by package name") {
		t.Error("expected missing package name error")
	}

	_, err = testParseFile(`package a
import "b" c`)
	if err == nil || !strings.Contains(err.Error(), "newline to follow import specification") {
		t.Error("expected missing semicolon or newline after import spec error")
	}

	_, err = testParseFile(`package a
import ("b" c`)
	if err == nil || !strings.Contains(err.Error(), "newline to follow import specification") {
		t.Error("expected missing semicolon or newline after import spec error")
	}

	_, err = testParseFile(`package a
import ("b" ;`)
	if err == nil || !strings.Contains(err.Error(), "missing closing parenthesis") {
		t.Error("expected missing closing parethesis error")
	}

	_, err = testParseFile(`package a
import (_ `)
	if err == nil || !strings.Contains(err.Error(), "missing import path") {
		t.Error("expected missing import path error", err)
	}

	_, err = testParseFile(`package a
import `)
	if err == nil || !strings.Contains(err.Error(), "missing import path") {
		t.Error("expected missing import path error", err)
	}
}

func TestParserParsePackage(t *testing.T) {
	d, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	gopath := filepath.Join(d, _TEST, "07")
	c, err := new(Config).Init(gopath, "linux", "amd64")
	if err != nil {
		t.Fatal(err)
	}
	dir := c.FindPackageDir("a")
	if len(dir) == 0 {
		t.Fatal("missing test package")
	}
	srcs, err := c.FindPackageFiles(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	_, err = parsePackage("a", dir, srcs)
	if err == nil ||
		!strings.Contains(err.Error(), "inconsistent package name: a, should be main") {
		t.Error("expected inconsistent package name error")
	}
}

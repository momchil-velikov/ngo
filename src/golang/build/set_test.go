package build

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestImportLoop1(t *testing.T) {
	d, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	gopath := filepath.Join(d, "test", "03")
	c, err := new(Config).Init(gopath, "linux", "amd64")

	_, err = c.CreateBuildSet("a", false)
	if err == nil || !strings.Contains(err.Error(), "a\" eventually imports itself") {
		t.Error("expected circular dependency error")
	} else {
		t.Log(err)
	}
}

func TestImportLoop2(t *testing.T) {
	d, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	gopath := filepath.Join(d, "test", "04")
	c, err := new(Config).Init(gopath, "linux", "amd64")

	_, err = c.CreateBuildSet("a", false)
	if err == nil || !strings.Contains(err.Error(), "a\" eventually imports itself") {
		t.Error("expected circular dependency error")
	} else {
		t.Log(err)
	}
}

func TestImportLoop3(t *testing.T) {
	d, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	gopath := filepath.Join(d, "test", "05")
	c, err := new(Config).Init(gopath, "linux", "amd64")

	_, err = c.CreateBuildSet("a", false)
	if err == nil || !strings.Contains(err.Error(), "a\" eventually imports itself") {
		t.Error("expected circular dependency error")
	} else {
		t.Log(err)
	}
}

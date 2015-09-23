package build

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfig(t *testing.T) {
	os.Setenv("GOOS", "darwin")
	os.Setenv("GOARCH", "arm")

	c, err := new(Config).Default()
	if err != nil {
		t.Fatal(err)
	}
	if c.OS != "darwin" {
		t.Errorf("unexpected OS: \"%s\"; must get \"darwin\" from environment", c.OS)
	}
	if c.Arch != "arm" {
		t.Errorf("unexpected CPU arch: \"%s\"; must get \"arm\" from environment", c.Arch)
	}

	c, err = new(Config).Init("", "freebsd", "mips")
	if err != nil {
		t.Fatal(err)
	}
	if c.OS != "freebsd" {
		t.Errorf("unexpected OS: \"%s\"; must get \"freebsd\" from args", c.OS)
	}
	if c.Arch != "mips" {
		t.Errorf("unexpected CPU arch: \"%s\"; must get \"mips\" from args", c.Arch)
	}

	c, err = new(Config).Init("", "dos", "amd64")
	if err == nil || err.Error() != "unrecognized OS: dos" {
		t.Error("must return invalid OS error")
	}

	c, err = new(Config).Init("", "linux", "z80")
	if err == nil || err.Error() != "unrecognized CPU arch: z80" {
		t.Error("must return invalid CPU arch error")
	}
}

func TestConfigPath(t *testing.T) {
	d, err := ioutil.TempDir("", "glang-test-")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(d)

	p1 := filepath.Join(d, "src1")
	if err := os.MkdirAll(filepath.Join(p1, "src", "pkg1"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(p1, "src", "pkg2", "pkg3"), 0700); err != nil {
		t.Fatal(err)
	}

	p2 := filepath.Join(d, "src2")
	if err := os.MkdirAll(filepath.Join(p2, "src", "pkg4"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(p2, "src", "pkg5", "pkg6"), 0700); err != nil {
		t.Fatal(err)
	}

	p3 := filepath.Join(d, "src3")
	if f, err := os.Create(p3); err == nil {
		defer f.Close()
	} else {
		t.Fatal(err)
	}

	gopath :=
		strings.Join([]string{p1, "", p2, "..", "./src", p3}, string(os.PathListSeparator))
	c, err := new(Config).Init(gopath, "linux", "amd64")
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Path) != 2 || c.Path[0] != p1 || c.Path[1] != p2 {
		t.Errorf("GOPATH expected to contain only \"%s\" and \"%s\"", p1, p2)
	}

	pkgs := []string{"pkg1", "pkg2", "pkg3", "pkg4", "pkg5", "pkg6"}
	exp := []string{p1, p1, "", p2, p2, ""}

	for i := range pkgs {
		s := c.FindPackageDir(pkgs[i])
		if len(s) == 0 && len(exp[i]) > 0 {
			t.Errorf("package \"%s\" not found; should be in \"%s\"", pkgs[i], exp[i])
		} else if len(s) > 0 && len(exp[i]) == 0 {
			t.Errorf("package \"%s\" should not be found", pkgs[i])
		} else if s != exp[i] {
			t.Errorf(
				"package \"%s\" found in \"%s\"; should be in \"%s\"",
				pkgs[i], s, exp[i],
			)
		}
	}

	pkg := "pkg2/pkg3"
	s := c.FindPackageDir(pkg)
	if len(s) == 0 {
		t.Errorf("package \"%s\" not found; should be in \"%s\"", pkg, p1)
	} else if s != p1 {
		t.Errorf(
			"package \"%s\" found in \"%s\"; should be in \"%s\"",
			pkg, s, p1,
		)
	}

	pkg = "pkg5/pkg6"
	s = c.FindPackageDir(pkg)
	if len(s) == 0 {
		t.Errorf("package \"%s\" not found; should be in \"%s\"", pkg, p2)
	} else if s != p2 {
		t.Errorf(
			"package \"%s\" found in \"%s\"; should be in \"%s\"",
			pkg, s, p2,
		)
	}
}

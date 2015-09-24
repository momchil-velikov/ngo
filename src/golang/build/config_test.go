package build

import (
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
	d, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	p1 := filepath.Join(d, "test", "01", "src1")
	p2 := filepath.Join(d, "test", "01", "src2")
	p3 := filepath.Join(d, "test", "01", "src3")
	gopath :=
		strings.Join([]string{p1, "", p2, ".", "./src", p3}, string(os.PathListSeparator))
	c, err := new(Config).Init(gopath, "linux", "amd64")
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Path) != 2 || c.Path[0] != p1 || c.Path[1] != p2 {
		t.Errorf("GOPATH expected to contain only \"%s\" and \"%s\"", p1, p2)
	}

	pkgs := []string{"pkg1", "pkg2", "pkg3", "pkg4", "pkg5", "pkg6"}
	exps := []string{
		filepath.Join(p1, "src", "pkg1"),
		filepath.Join(p1, "src", "pkg2"),
		"",
		filepath.Join(p2, "src", "pkg4"),
		filepath.Join(p2, "src", "pkg5"),
		"",
	}

	for i := range pkgs {
		s := c.FindPackageDir(pkgs[i])
		if len(s) == 0 && len(exps[i]) > 0 {
			t.Errorf("package \"%s\" not found; should be in \"%s\"", pkgs[i], exps[i])
		} else if len(s) > 0 && len(exps[i]) == 0 {
			t.Errorf("package \"%s\" should not be found", pkgs[i])
		} else if s != exps[i] {
			t.Errorf(
				"package \"%s\" found in \"%s\"; should be in \"%s\"",
				pkgs[i], s, exps[i],
			)
		}
	}

	pkg := "pkg2/pkg3"
	exp := filepath.Join(p1, "src", "pkg2", "pkg3")
	s := c.FindPackageDir(pkg)
	if len(s) == 0 {
		t.Errorf(
			"package \"%s\" not found; should be in \"%s\"",
			pkg, exp,
		)
	} else if s != exp {
		t.Errorf(
			"package \"%s\" found in \"%s\"; should be in \"%s\"",
			pkg, s, exp,
		)
	}

	pkg = "pkg5/pkg6"
	exp = filepath.Join(p2, "src", "pkg5", "pkg6")
	s = c.FindPackageDir(pkg)
	if len(s) == 0 {
		t.Errorf(
			"package \"%s\" not found; should be in \"%s\"",
			pkg, exp,
		)
	} else if s != exp {
		t.Errorf(
			"package \"%s\" found in \"%s\"; should be in \"%s\"",
			pkg, s, exp,
		)
	}
}

package build

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

const _TEST = "_test"

func TestConfig(t *testing.T) {
	os.Setenv("GOOS", "darwin")
	os.Setenv("GOARCH", "arm")
	defer os.Unsetenv("GOOS")
	defer os.Unsetenv("GOARCH")

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
	p1 := filepath.Join(d, _TEST, "01", "src1")
	p2 := filepath.Join(d, _TEST, "01", "src2")
	p3 := filepath.Join(d, _TEST, "01", "src3")
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

func checkSame(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

type filterTestCase struct {
	os, arch string
	test     bool
	names    []string
}

func TestConfigFilterNames(t *testing.T) {
	cases := []filterTestCase{
		{"linux", "amd64", false,
			[]string{"f2_linux_amd64.go", "f4_linux.go", "f6_amd64.go", "f8.go"}},
		{"linux", "amd64", true,
			[]string{"f1_linux_amd64_test.go", "f2_linux_amd64.go", "f3_linux_test.go",
				"f4_linux.go", "f5_amd64_test.go", "f6_amd64.go", "f7_test.go", "f8.go"}},
		{"freebsd", "amd64", false,
			[]string{"f12_freebsd.go", "f6_amd64.go", "f8.go"}},
		{"freebsd", "amd64", true,
			[]string{"f11_freebsd_test.go", "f12_freebsd.go", "f5_amd64_test.go",
				"f6_amd64.go", "f7_test.go", "f8.go"}},
		{"linux", "arm", false,
			[]string{"f14_arm.go", "f4_linux.go", "f8.go"}},
		{"linux", "arm", true,
			[]string{"f13_arm_test.go", "f14_arm.go", "f3_linux_test.go", "f4_linux.go",
				"f7_test.go", "f8.go"}},
		{"freebsd", "arm", false,
			[]string{"f10_freebsd_arm.go", "f12_freebsd.go", "f14_arm.go", "f8.go"}},
		{"freebsd", "arm", true,
			[]string{"f10_freebsd_arm.go", "f11_freebsd_test.go", "f12_freebsd.go",
				"f13_arm_test.go", "f14_arm.go", "f7_test.go", "f8.go",
				"f9_freebsd_arm_test.go"}},
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	for i := range cases {
		c, err :=
			new(Config).Init(filepath.Join(cwd, _TEST, "02"), cases[i].os, cases[i].arch)
		if err != nil {
			t.Fatal(err)
		}

		dir := c.FindPackageDir("pkg")
		if len(dir) == 0 {
			t.Fatal("package \"pkg\" not found")
		}

		names, err := c.FindPackageFiles(dir, cases[i].test)
		if err != nil {
			t.Fatal(err)
		}
		sort.Sort(sort.StringSlice(names))

		if !checkSame(names, cases[i].names) {
			t.Errorf("Wrong list of files for %s/%s/test=%v: %v",
				cases[i].os, cases[i].arch, cases[i].test, names)
		}
	}
}

func TestConfigBadDir(t *testing.T) {
	c, err := new(Config).Default()
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.FindPackageFiles("bad-dir-name", true)
	if err == nil {
		t.Error("expected missing directory error")
	}
}

func TestConfigImportLoop1(t *testing.T) {
	d, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	gopath := filepath.Join(d, _TEST, "03")
	c, err := new(Config).Init(gopath, "linux", "amd64")

	_, err = c.CreateBuildSet("a", false)
	if err == nil || !strings.Contains(err.Error(), "a\" eventually imports itself") {
		t.Error("expected circular dependency error")
	} else {
		t.Log(err)
	}
}

func TestConfigImportLoop2(t *testing.T) {
	d, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	gopath := filepath.Join(d, _TEST, "04")
	c, err := new(Config).Init(gopath, "linux", "amd64")

	_, err = c.CreateBuildSet("a", false)
	if err == nil || !strings.Contains(err.Error(), "a\" eventually imports itself") {
		t.Error("expected circular dependency error")
	} else {
		t.Log(err)
	}
}

func TestConfigImportLoop3(t *testing.T) {
	d, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	gopath := filepath.Join(d, _TEST, "05")
	c, err := new(Config).Init(gopath, "linux", "amd64")

	_, err = c.CreateBuildSet("a", false)
	if err == nil || !strings.Contains(err.Error(), "a\" eventually imports itself") {
		t.Error("expected circular dependency error")
	} else {
		t.Log(err)
	}
}
func TestConfigBuildSetOrder(t *testing.T) {
	d, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	gopath := filepath.Join(d, _TEST, "06")
	c, err := new(Config).Init(gopath, "linux", "amd64")

	pkgs, err := c.CreateBuildSet("a", false)
	if err != nil {
		t.Error(err)
	}

	for _, p := range pkgs {
		p.Mark = 0
	}
	no := 1
	for _, p := range pkgs {
		p.Mark = no
		no++
		s := fmt.Sprintf("%s(%d): ", p.Name, p.Mark)
		for _, d := range p.Imports {
			if d.Mark == 0 || d.Mark >= p.Mark {
				t.Errorf("invalid order: %s should precede %s\n", d.Name, p.Name)
			}
			s += fmt.Sprintf("%s(%d) ", d.Name, d.Mark)
		}
		t.Log(s)
	}
}

func TestConfigDirectives(t *testing.T) {
	d, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	gopath := filepath.Join(d, _TEST, "08")
	c, err := new(Config).Init(gopath, "android", "arm")
	if err != nil {
		t.Fatal(err)
	}
	pkgs, err := c.CreateBuildSet("a", false)
	if err != nil || len(pkgs) != 1 {
		t.Fatal("should return exactly one package")
	}
	pkg := pkgs[0]
	es := []string{"a.go", "c.go", "e.go", "g.go"}
	if len(pkg.Files) != len(es) {
		t.Error("incorrect number of files")
	}
	for i := range pkg.Files {
		name := filepath.Base(pkg.Files[i].Path)
		if !matchAny(name, es) {
			t.Errorf("file %s incorrecly included in build", name)
		}
	}
}

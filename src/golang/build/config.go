package build

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var (
	knownOS   = []string{"linux", "darwin", "freebsd", "netbsd", "openbsd", "android"}
	knownArch = []string{"amd64", "arm", "powerpc", "mips"}
)

type Config struct {
	Path []string
	Arch string
	OS   string
	Tags []string
}

func matchAny(s string, ss []string) bool {
	for i := range ss {
		if s == ss[i] {
			return true
		}
	}
	return false
}

func suffixMatchAny(s string, ss []string) string {
	for i := range ss {
		tag := "_" + ss[i]
		if strings.HasSuffix(s, tag) {
			return tag
		}
	}
	return ""
}

// Initializes a build configuraion from environment.
func (c *Config) Default() (*Config, error) {
	return c.Init("", "", "")
}

// Initializes a build configuration with the given GOPATH, GOOS and
// GOARCH. If either string is empty, take the value from the environment.
func (c *Config) Init(gopath string, goos string, goarch string) (*Config, error) {
	if len(gopath) == 0 {
		gopath = os.Getenv("GOPATH")
	}
	if len(goos) == 0 {
		goos = os.Getenv("GOOS")
	}
	if len(goos) > 0 && !matchAny(goos, knownOS) {
		return nil, errors.New("unrecognized OS: " + goos)
	}
	c.OS = goos
	if len(goarch) == 0 {
		goarch = os.Getenv("GOARCH")
	}
	if len(goarch) > 0 && !matchAny(goarch, knownArch) {
		return nil, errors.New("unrecognized CPU arch: " + goarch)
	}
	c.Arch = goarch
	ps := strings.Split(gopath, string(os.PathListSeparator))
	for i := range ps {
		if p := filepath.Clean(ps[i]); filepath.IsAbs(p) {
			if info, err := os.Stat(p); err == nil && info.IsDir() {
				c.Path = append(c.Path, p)
			}
		}
	}
	if len(c.OS) > 0 {
		c.Tags = append(c.Tags, c.OS)
		if c.OS == "android" {
			c.Tags = append(c.Tags, "linux")
		}
	}
	if len(c.Arch) > 0 {
		c.Tags = append(c.Tags, c.Arch)
	}
	return c, nil
}

// Finds the directory, which contains the source files of the given package.
func (c *Config) FindPackageDir(pkg string) string {
	for i := range c.Path {
		path := filepath.Join(c.Path[i], "src", pkg)
		if info, err := os.Stat(path); err == nil {
			if info.IsDir() {
				return path
			}
		}
	}
	return ""
}

func (c *Config) isAcceptableName(name string, test bool) bool {
	if name[0] == '.' || name[0] == '_' || filepath.Ext(name) != ".go" {
		return false
	}
	name = name[0 : len(name)-3]
	if !test && strings.HasSuffix(name, "_test") {
		return false
	}
	name = strings.TrimSuffix(name, "_test")
	if arch := suffixMatchAny(name, knownArch); len(arch) > 0 {
		if arch[1:] != c.Arch {
			return false
		}
		name = strings.TrimSuffix(name, arch)
	}
	if os := suffixMatchAny(name, knownOS); len(os) > 0 {
		if os[1:] != c.OS {
			return false
		}
	}
	return true
}

func (c *Config) filterByName(ns []string, test bool) []string {
	i, j := 0, 0
	for j < len(ns) {
		if c.isAcceptableName(ns[j], test) {
			ns[i] = ns[j]
			i++
		}
		j++
	}
	return ns[0:i]
}

func (c *Config) FindPackageFiles(dir string, test bool) ([]string, error) {
	infos, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(infos))
	for _, info := range infos {
		if !info.IsDir() {
			names = append(names, info.Name())
		}
	}
	return c.filterByName(names, test), nil
}

func (c *Config) evalTag(tag string) bool {
	neg := false
	if tag[0] == '!' {
		neg = true
		tag = tag[1:]
	}
	if tag == "ignore" {
		return neg
	}
	for _, t := range c.Tags {
		if tag == t {
			return !neg
		}
	}
	return neg
}

func (c *Config) evalBuildDirective(d string) bool {
	fs := strings.Fields(d)
	if fs[0] != "+build" {
		return true
	}
	fs = fs[1:]
	for _, f := range fs {
		ts := strings.Split(f, ",")
		value := true
		for _, t := range ts {
			value = value && c.evalTag(t)
		}
		if value {
			return true
		}
	}
	return false
}

func (c *Config) filterByTags(in []*File) []*File {
	out := []*File{}
L:
	for _, f := range in {
		if len(f.Comments) == 0 {
			out = append(out, f)
			continue
		}
		for _, d := range f.Comments {
			if !c.evalBuildDirective(d) {
				continue L
			}
		}
		out = append(out, f)
	}
	return out
}

func (c *Config) CreateBuildSet(path string, test bool) ([]*Package, error) {
	// Parse the initial package.
	dir := c.FindPackageDir(path)
	if len(dir) == 0 {
		return nil, errors.New(fmt.Sprintf("package %s not found", path))
	}
	srcs, err := c.FindPackageFiles(dir, test)
	if err != nil {
		return nil, err
	}
	pkg, err := parsePackage(dir, srcs)
	if err != nil {
		return nil, err
	}
	pkg.Files = c.filterByTags(pkg.Files)
	// Find all the dependency packages.
	err = c.findPackageDeps(pkg, map[string]*Package{dir: pkg})
	if err != nil {
		return nil, err
	}
	// Sort the packages in topological order.
	set, err := postorder(pkg, nil)
	if err != nil {
		return nil, err
	}
	n := len(set)
	i, j := 0, n-1
	for i < n/2 {
		set[i], set[j] = set[j], set[i]
		i++
		j--
	}
	return set, nil
}

func postorder(pkg *Package, set []*Package) ([]*Package, error) {
	var err error
	pkg.Mark = 1
	for _, p := range pkg.Imports {
		if p.Mark == 2 {
			continue
		}
		if p.Mark == 1 {
			msg := fmt.Sprintf("package \"%s\" eventually imports itself", p.Path)
			return nil, errors.New(msg)
		}
		set, err = postorder(p, set)
		if err != nil {
			return nil, err
		}
	}
	pkg.Mark = 2
	return append(set, pkg), nil
}

// Finds all the package dependencies, recursively.
func (c *Config) findPackageDeps(pkg *Package, found map[string]*Package) error {
	// Walk over all the import declarations in all the source files of the
	// package.
	for _, f := range pkg.Files {
		for _, path := range f.Imports {
			// Find the dependent package directory.
			dir := c.FindPackageDir(path)
			if len(dir) == 0 {
				return errors.New(fmt.Sprintf("package \"%s\" not found", path))
			}
			// Check if the package was already encountered.
			dep, ok := found[dir]
			if !ok {
				// Find the source files and parse the package.
				srcs, err := c.FindPackageFiles(dir, false)
				if err != nil {
					return err
				}
				dep, err = parsePackage(dir, srcs)
				if err != nil {
					return err
				}
				dep.Files = c.filterByTags(dep.Files)
				// Find the dependencies of the new package, recursively.
				found[dir] = dep
				err = c.findPackageDeps(dep, found)
				if err != nil {
					return err
				}
			}
			// Record the package dependency.
			pkg.Imports[dir] = dep
		}
	}
	return nil
}

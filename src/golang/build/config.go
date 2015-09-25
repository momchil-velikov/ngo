package build

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var (
	knownOS   = []string{"linux", "darwin", "freebsd", "netbsd", "openbsd"}
	knownArch = []string{"amd64", "arm", "powerpc", "mips"}
)

type configError string

func (err configError) Error() string {
	return string(err)
}

type Config struct {
	Path []string
	Arch string
	OS   string
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
	if !matchAny(goos, knownOS) {
		return nil, configError("unrecognized OS: " + goos)
	}
	c.OS = goos
	if len(goarch) == 0 {
		goarch = os.Getenv("GOARCH")
	}
	if !matchAny(goarch, knownArch) {
		return nil, configError("unrecognized CPU arch: " + goarch)
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
	if filepath.Ext(name) != ".go" {
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

func (c *Config) filterNames(ns []string, test bool) []string {
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
	return c.filterNames(names, test), nil
}

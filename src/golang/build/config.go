package build

import (
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

func (c *Config) Default() (*Config, error) {
	return c.Init("", "", "")
}

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

// Finds that component of GOPATH, which is the top level workspace directory
// of the given package.
func (c *Config) FindPackageDir(pkg string) string {
	for i := range c.Path {
		path := filepath.Join(c.Path[i], "src", pkg)
		if info, err := os.Stat(path); err == nil {
			if info.IsDir() {
				return c.Path[i]
			}
		}
	}
	return ""
}

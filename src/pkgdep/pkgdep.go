package main

import (
	"flag"
	"fmt"
	"golang/ast"
	"golang/build"
	"io"
	"os"
)

func writeDot(w io.Writer, name string, pkgs []*ast.Package) {
	for i := range pkgs {
		pkgs[i].Mark = i
	}
	fmt.Fprintf(w, "digraph \"%s\" {\n", name)
	for _, pkg := range pkgs {
		fmt.Fprintf(w, "n%d[label=%s]\n", pkg.Mark, pkg.Name)
		for _, dep := range pkg.Imports {
			fmt.Fprintf(w, "n%d -> n%d\n", pkg.Mark, dep.Mark)
		}
	}
	fmt.Fprintf(w, "}\n")
}

func pkgdepMain() int {
	path := flag.String("path", "", "Lookup path for packages (overrides GOPATH)")
	osys := flag.String("os", "", "Operating system family (overrides GOOS)")
	arch := flag.String("arch", "", "CPU family (overrides GOARCH)")
	out := flag.String("o", "", "Output file name")
	flag.Parse()

	w := os.Stdout
	if len(*out) > 0 {
		f, err := os.Create(*out)
		if err != nil {
			fmt.Fprintln(os.Stderr, *out, ":", err)
			return 1
		}
		defer f.Close()
		w = f
	}
	c, err := new(build.Config).Init(*path, *osys, *arch)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	for _, name := range flag.Args() {
		pkgs, err := c.CreateBuildSet(name, false)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		if len(pkgs) > 0 {
			writeDot(w, name, pkgs)
		}
	}
	return 0
}

func main() {
	os.Exit(pkgdepMain())
}

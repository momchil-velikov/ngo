package build

import (
	"fmt"
	"golang/ast"
	"golang/constexpr"
	"golang/parser"
)

func (c *Config) CreateBuildSet(path string, test bool) ([]*ast.Package, error) {
	// Parse the initial package.
	dir := c.FindPackageDir(path)
	if len(dir) == 0 {
		return nil, configError(fmt.Sprintf("package %s not found", path))
	}
	srcs, err := c.FindPackageFiles(dir, test)
	if err != nil {
		return nil, err
	}
	pkg, err := parser.ParsePackage(dir, srcs)
	if err != nil {
		return nil, err
	}
	// Find all the dependency packages.
	err = c.findPackageDeps(pkg, map[string]*ast.Package{dir: pkg})
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

func postorder(pkg *ast.Package, set []*ast.Package) ([]*ast.Package, error) {
	var err error
	pkg.Mark = 1
	for _, p := range pkg.Imports {
		if p.Mark == 2 {
			continue
		}
		if p.Mark == 1 {
			msg := fmt.Sprintf("package \"%s\" eventually imports itself", p.Path)
			return nil, configError(msg)
		}
		set, err = postorder(p, set)
		if err != nil {
			return nil, err
		}
	}
	pkg.Mark = 2
	return append(set, pkg), nil
}

func walkImports(decls []ast.Decl, fn func(*ast.Import) error) error {
	for _, d := range decls {
		switch i := d.(type) {
		case *ast.DeclGroup:
			return walkImports(i.Decls, fn)
		case *ast.Import:
			if e := fn(i); e != nil {
				return e
			}
		default:
			panic("invalid import type")
		}
	}
	return nil
}

// Finds all the package dependencies, recursively.
func (c *Config) findPackageDeps(pkg *ast.Package, found map[string]*ast.Package) error {
	// Walk over all the import declarations in all the source files of the
	// package.
	for _, f := range pkg.Files {
		err := walkImports(f.Imports, func(i *ast.Import) error {
			// Find the dependent package directory.
			path := constexpr.String(i.Path)
			dir := c.FindPackageDir(path)
			if len(dir) == 0 {
				return configError(fmt.Sprintf("package \"%s\" not found", path))
			}
			// Check if the package was already encountered.
			dep, ok := found[dir]
			if !ok {
				// Find the source files and parse the package.
				srcs, err := c.FindPackageFiles(dir, false)
				if err != nil {
					return err
				}
				dep, err = parser.ParsePackage(dir, srcs)
				if err != nil {
					return err
				}
				// Find the dependencies of the new package, recursively.
				found[dir] = dep
				err = c.findPackageDeps(dep, found)
				if err != nil {
					return err
				}
			}
			// Record the package dependency.
			pkg.Imports[dir] = dep
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

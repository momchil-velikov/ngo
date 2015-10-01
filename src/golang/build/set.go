package build

import (
	"errors"
	"fmt"
)

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

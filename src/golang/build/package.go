package build

// Package
type Package struct {
	Dir     string // Full pathname of the package directory
	Path    string // Import path of the package
	Name    string // Package name
	Files   []*File
	Imports map[string]*Package
	Mark    int
}

// Source file
type File struct {
	Path, Package string
	Imports       []string
	Comments      []string
}

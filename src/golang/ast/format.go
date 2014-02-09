package ast

// import "fmt"

type Formatter interface {
	format(uint) string
}

// Return a string, consisting of N spaces.
func nspaces(n uint) (s string) {
	for i := uint(0); i < n; i++ {
		s += " "
	}
	return
}

// Output a formatted source file.
func (f *File) Format() (s string) {
	s = "package " + f.PackageName + "\n\n"

	if len(f.Imports) > 0 {
		if len(f.Imports) == 1 {
			s += "import " + f.Imports[0].format(0) + "\n"
		} else {
			s += "import (\n"
			for _, i := range f.Imports {
				s += i.format(1) + "\n"
			}
			s += ")\n\n"
		}
	}

	for _, d := range f.Decls {
		s += d.format(0)
	}

	return
}

// Output formatted import clause with N levels of indentation.
func (i *Import) format(n uint) (s string) {
	s = nspaces(4 * n)
	if len(i.Name) > 0 {
		s += i.Name + " "
	}
	s += `"` + i.Path + `"`
	return s
}

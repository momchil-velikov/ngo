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

// Output a formatted type declaration.
func (t *TypeDecl) format(n uint) (s string) {
	s = "type " + t.Name + " "
	s += t.Type.format(n) + "\n\n"
	return
}

// Output a formatted type.
func (t *BaseType) format(n uint) (s string) {
	s = nspaces(4 * n)
	if len(t.Pkg) > 0 {
		s += t.Pkg + "."
	}
	s += t.Id
	return
}

func (t *ArrayType) format(n uint) (s string) {
	s = nspaces(4 * n)
	s += "[" + t.Dim.format(0) + "]" + t.EltType.format(0)
	return
}

func (t *SliceType) format(n uint) (s string) {
	s = nspaces(4 * n)
	s += "[]" + t.EltType.format(0)
	return
}

func (t *PtrType) format(n uint) string {
	return "*" + t.Base.format(0)
}

func (t *MapType) format(n uint) string {
	return "map[" + t.KeyType.format(0) + "]" + t.EltType.format(0)
}

// Output a formatted expression
func (e *Expr) format(n uint) string {
	return e.Const
}

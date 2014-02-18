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
func (t *TypeName) format(n uint) (s string) {
    if len(t.Pkg) > 0 {
        s += t.Pkg + "."
    }
    s += t.Id
    return
}

func (t *ArrayType) format(n uint) (s string) {
    return "[" + t.Dim.format(n+1) + "]" + t.EltType.format(n)
}

func (t *SliceType) format(n uint) (s string) {
    return "[]" + t.EltType.format(n)
}

func (t *PtrType) format(n uint) string {
    return "*" + t.Base.format(n)
}

func (t *MapType) format(n uint) string {
    return "map[" + t.KeyType.format(0) + "]" + t.EltType.format(n)
}

func (t *ChanType) format(n uint) (s string) {
    if !t.Send {
        s += "<-"
    }
    s += "chan"
    if !t.Recv {
        s += "<- "
    } else {
        s += " "
    }
    s += t.EltType.format(n)
    return
}

func (t *StructType) format(n uint) string {
    if len(t.Fields) == 0 {
        return "struct{}"
    }
    sp := nspaces(4 * n)
    sp1 := sp + "    "
    s := "struct {\n"
    for _, f := range t.Fields {
        s += sp1
        if len(f.Name) > 0 {
            s += f.Name + " "
        }
        s += f.Type.format(n + 1)
        if len(f.Tag) > 0 {
            s += " \"" + f.Tag +
                "\""
        }
        s += "\n"

    }
    s += sp + "}"
    return s
}

func (t *FuncType) format(n uint) string {
    return "func" + format_signature(t, n)
}

func format_signature(t *FuncType, n uint) (s string) {
    k := len(t.Params)
    if k == 0 {
        s = "()"
    } else {
        s = "(" + format_params(t.Params, n) + ")"
    }

    k = len(t.Returns)
    if k == 1 && len(t.Returns[0].Name) == 0 {
        s += " " + t.Returns[0].Type.format(n+1)
    } else if k > 0 {
        s += " (" + format_params(t.Returns, n) + ")"
    }
    return
}

func format_params(p []*ParamDecl, n uint) (s string) {
    if len(p[0].Name) > 0 {
        s += p[0].Name + " "
    }
    if p[0].Variadic {
        s += "..."
    }
    s += p[0].Type.format(n + 1)
    for i, k := 1, len(p); i < k; i++ {
        s += ", "
        if len(p[i].Name) > 0 {
            s += p[i].Name + " "
        }
        if p[i].Variadic {
            s += "..."
        }
        s += p[i].Type.format(n + 1)
    }
    return
}

// Output a formatted expression
func (e *Expr) format(n uint) string {
    return e.Const
}

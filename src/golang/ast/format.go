package ast

// import "fmt"

type Formatter interface {
    Format(uint) string
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
            s += "import " + f.Imports[0].Format(0) + "\n"
        } else {
            s += "import (\n"
            for _, i := range f.Imports {
                s += i.Format(1) + "\n"
            }
            s += ")\n\n"
        }
    }

    for _, d := range f.Decls {
        s += d.Format(0)
    }

    return
}

// Output formatted import clause with N levels of indentation.
func (i *Import) Format(n uint) (s string) {
    s = nspaces(4 * n)
    if len(i.Name) > 0 {
        s += i.Name + " "
    }
    s += `"` + i.Path + `"`
    return s
}

// Output a formatted type declaration.
func (t *TypeDecl) Format(n uint) (s string) {
    s = "type " + t.Name + " "
    s += t.Type.Format(n) + "\n\n"
    return
}

// Output a formatter constant group declaration
func (c *ConstGroup) Format(n uint) string {
    indent := nspaces(4 * n)
    s := indent + "const (\n"
    for _, d := range c.Decls {
        s += d.format_internal(n+1, false) + "\n"
    }
    s += indent + ")\n"
    return s
}

// Output a formatted constant declaration.
func (c *ConstDecl) Format(n uint) string {
    return c.format_internal(n, true) + "\n"
}

func (c *ConstDecl) format_internal(n uint, top bool) (s string) {
    if top {
        s = "const "
    } else {
        s = nspaces(4 * n)
    }
    s += c.Names[0]
    for i := 1; i < len(c.Names); i++ {
        s += ", " + c.Names[i]
    }
    if c.Type != nil {
        s += " " + c.Type.Format(n+1)
    }
    if k := len(c.Values); k > 0 {
        s += " = " + c.Values[0].Format(n+1)
        for i := 1; i < k; i++ {
            s += ", " + c.Values[i].Format(n+1)
        }
    }
    return s
}

// Output a formatted variable group declaration
func (c *VarGroup) Format(n uint) string {
    indent := nspaces(4 * n)
    s := indent + "var (\n"
    for _, d := range c.Decls {
        s += d.format_internal(n+1, false) + "\n"
    }
    s += indent + ")\n"
    return s
}

// Output a formatted variable declaration.
func (c *VarDecl) Format(n uint) string {
    return c.format_internal(n, true) + "\n"
}

func (c *VarDecl) format_internal(n uint, top bool) (s string) {
    if top {
        s = "var "
    } else {
        s = nspaces(4 * n)
    }
    s += c.Names[0]
    for i := 1; i < len(c.Names); i++ {
        s += ", " + c.Names[i]
    }
    if c.Type != nil {
        s += " " + c.Type.Format(n+1)
    }
    if k := len(c.Init); k > 0 {
        s += " = " + c.Init[0].Format(n+1)
        for i := 1; i < k; i++ {
            s += ", " + c.Init[i].Format(n+1)
        }
    }
    return s
}

// Output a formatted function declaration.
func (f *FuncDecl) Format(n uint) string {
    s := "func"
    if f.Recv != nil {
        s += " " + f.Recv.Format(n+1)
    }
    s += " " + f.Name + format_signature(f.Sig, n+1)
    if f.Body != nil {
        s += f.Body.Format(n + 1)
    }
    return s + "\n"
}

// Output a formatted block
func (b *Block) Format(n uint) string {
    return " {\n}"
}

// Output a formatter method receiver.
func (r *Receiver) Format(n uint) string {
    s := "("
    if len(r.Name) > 0 {
        s += r.Name + " "
    }
    return s + r.Type.Format(n+1) + ")"
}

// Output a formatted type.
func (t *QualId) Format(n uint) (s string) {
    if len(t.Pkg) > 0 {
        s += t.Pkg + "."
    }
    s += t.Id
    return
}

func (t *ArrayType) Format(n uint) (s string) {
    return "[" + t.Dim.Format(n+1) + "]" + t.EltType.Format(n)
}

func (t *SliceType) Format(n uint) (s string) {
    return "[]" + t.EltType.Format(n)
}

func (t *PtrType) Format(n uint) string {
    return "*" + t.Base.Format(n)
}

func (t *MapType) Format(n uint) string {
    return "map[" + t.KeyType.Format(0) + "]" + t.EltType.Format(n)
}

func (t *ChanType) Format(n uint) (s string) {
    if !t.Send {
        s += "<-"
    }
    s += "chan"
    if !t.Recv {
        s += "<- " + t.EltType.Format(n)
    } else if ch, ok := t.EltType.(*ChanType); ok && !ch.Send {
        s += " (" + ch.Format(n) + ")"
    } else {
        s += " " + t.EltType.Format(n)
    }
    return
}

func (t *StructType) Format(n uint) string {
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
        s += f.Type.Format(n + 1)
        if len(f.Tag) > 0 {
            s += " \"" + f.Tag +
                "\""
        }
        s += "\n"

    }
    s += sp + "}"
    return s
}

func (t *FuncType) Format(n uint) string {
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
        s += " " + t.Returns[0].Type.Format(n+1)
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
    s += p[0].Type.Format(n + 1)
    for i, k := 1, len(p); i < k; i++ {
        s += ", "
        if len(p[i].Name) > 0 {
            s += p[i].Name + " "
        }
        if p[i].Variadic {
            s += "..."
        }
        s += p[i].Type.Format(n + 1)
    }
    return
}

func (t *InterfaceType) Format(n uint) string {
    if len(t.Embed) == 0 && len(t.Methods) == 0 {
        return "interface{}"
    }
    s := "interface {\n"
    indent := nspaces(4 * (n + 1))
    for _, e := range t.Embed {
        s += indent + e.Format(n+1) + "\n"
    }
    for _, m := range t.Methods {
        s += indent + m.Name + format_signature(m.Sig, n+1) + "\n"
    }
    s += nspaces(4*n) + "}"
    return s
}

// Output a formatted expression
func (e *Operand) Format(n uint) string {
    return e.Const
}

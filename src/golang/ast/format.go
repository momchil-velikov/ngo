package ast

import (
	//    "fmt"
	s "golang/scanner"
)

type Formatter interface {
	Format(uint) string
}

const indentStr = "    "

// Return a string for N levels of indentation.
func indent(n uint) (s string) {
	for i := uint(0); i < n; i++ {
		s += indentStr
	}
	return
}

// Output a formatted source file.
func (f *File) Format() (s string) {
	s = "package " + f.PackageName + "\n"

	if len(f.Imports) > 0 {
		if len(f.Imports) == 1 {
			s += "\nimport " + f.Imports[0].Format(0)
		} else {
			s += "\nimport ("
			for _, i := range f.Imports {
				s += "\n" + indent(1) + i.Format(0)
			}
			s += "\n)"
		}
	}

	for _, d := range f.Decls {
		s += "\n" + d.Format(0)
	}

	return s + "\n"
}

// Output formatted import clause with N levels of indentation.
func (i *Import) Format(n uint) (s string) {
	if len(i.Name) > 0 {
		s += i.Name + " "
	}
	s += `"` + i.Path + `"`
	return s
}

// Output error node
func (e *Error) Format(n uint) string {
	return "<error>"
}

// Output a formatted type group declaration
func (c *TypeGroup) Format(n uint) string {
	s := "type ("
	for _, d := range c.Decls {
		s += "\n" + indent(n+1) + d.formatInternal(n+1, true)
	}
	s += "\n" + indent(n) + ")"
	return s
}

// Output a formatted type declaration.
func (t *TypeDecl) Format(n uint) string {
	return t.formatInternal(n, false)
}

func (t *TypeDecl) formatInternal(n uint, group bool) (s string) {
	if !group {
		s += "type "
	}
	s += t.Name + " " + t.Type.Format(n)
	return
}

// Output a formatted constant group declaration
func (c *ConstGroup) Format(n uint) string {
	s := "const ("
	ind := "\n" + indent(n+1)
	for _, d := range c.Decls {
		s += ind + d.formatInternal(n+1, true)
	}
	s += "\n" + indent(n) + ")"
	return s
}

// Output a formatted constant declaration.
func (c *ConstDecl) Format(n uint) string {
	return c.formatInternal(n, false)
}

func (c *ConstDecl) formatInternal(n uint, group bool) (s string) {
	if !group {
		s += "const "
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
	s := "var ("
	ind := "\n" + indent(n+1)
	for _, d := range c.Decls {
		s += ind + d.formatInternal(n+1, true)
	}
	s += "\n" + indent(n) + ")"
	return s
}

// Output a formatted variable declaration.
func (c *VarDecl) Format(n uint) string {
	return c.formatInternal(n, false)
}

func (c *VarDecl) formatInternal(n uint, group bool) (s string) {
	if !group {
		s += "var "
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
		s += " " + f.Recv.Format(0)
	}
	s += " " + f.Name + formatSignature(f.Sig, n)
	if f.Body != nil {
		return s + " " + f.Body.Format(n)
	} else {
		return s
	}
}

// Output a formatted method receiver.
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
	if t.Dim == nil {
		s = "[...]"
	} else {
		s = "[" + t.Dim.Format(n+1) + "]"
	}
	s += t.EltType.Format(n)
	return
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
	sp := indent(n)
	sp1 := indent(n + 1)
	s := "struct {\n"
	for _, f := range t.Fields {
		s += sp1
		s += formatIdList(f.Names)
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

func formatIdList(id []string) (s string) {
	n := len(id)
	for i, nm := range id {
		if len(nm) > 0 {
			s += nm
			if i+1 < n {
				s += ", "
			}
		}
	}
	if len(s) > 0 {
		s += " "
	}
	return s
}

func (t *FuncType) Format(n uint) string {
	return "func" + formatSignature(t, n)
}

func formatSignature(t *FuncType, n uint) (s string) {
	k := len(t.Params)
	if k == 0 {
		s = "()"
	} else {
		s = "(" + formatParams(t.Params, n) + ")"
	}

	k = len(t.Returns)
	if k == 1 && len(t.Returns[0].Name) == 0 {
		s += " " + t.Returns[0].Type.Format(n+1)
	} else if k > 0 {
		s += " (" + formatParams(t.Returns, n) + ")"
	}
	return
}

func formatParams(p []*ParamDecl, n uint) (s string) {
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
	s := "interface {"
	ind := "\n" + indent(n+1)
	for _, e := range t.Embed {
		s += ind + e.Format(n+1)
	}
	for _, m := range t.Methods {
		s += ind + m.Name + formatSignature(m.Sig, n+1)
	}
	s += "\n" + indent(n) + "}"
	return s
}

// Output a formatted expression
func (e *Literal) Format(n uint) string {
	switch e.Kind {
	case s.INTEGER, s.FLOAT:
		return e.Value
	case s.RUNE:
		return "'" + e.Value + "'"
	case s.IMAGINARY:
		return e.Value + "i"
	case s.STRING:
		return "\"" + e.Value + "\""
	default:
		panic("invalid literal")
	}
}

func (e *TypeAssertion) Format(n uint) (s string) {
	s = e.Arg.Format(n)
	switch e.Arg.(type) {
	case *UnaryExpr, *BinaryExpr:
		s = "(" + s + ")"
	}
	s += ".(" + e.Type.Format(n) + ")"
	return
}

func (e *Selector) Format(n uint) (s string) {
	s = e.Arg.Format(n)
	switch e.Arg.(type) {
	case *UnaryExpr, *BinaryExpr:
		s = "(" + s + ")"
	}
	s += "." + e.Id
	return
}

func (e *IndexExpr) Format(n uint) (s string) {
	s = e.Array.Format(n)
	switch e.Array.(type) {
	case *UnaryExpr, *BinaryExpr:
		s = "(" + s + ")"
	}
	s += "[" + e.Idx.Format(n) + "]"
	return
}

func (e *SliceExpr) Format(n uint) (s string) {
	s = e.Array.Format(n)
	switch e.Array.(type) {
	case *UnaryExpr, *BinaryExpr:
		s = "(" + s + ")"
	}
	s += "["
	if e.Low != nil {
		s += e.Low.Format(n) + " :"
	} else {
		s += ":"
	}
	if e.High != nil {
		s += " " + e.High.Format(n)
	}
	if e.Cap != nil {
		s += " : " + e.Cap.Format(n)
	}
	s += "]"
	return
}

func (e *MethodExpr) Format(n uint) string {
	return "(" + e.Type.Format(n) + ")." + e.Id
}

func (e *CompLiteral) Format(n uint) (s string) {
	if e.Type != nil {
		s = e.Type.Format(n)
	}
	s += "{"
	m := len(e.Elts)
	for i, elt := range e.Elts {
		s += elt.format()
		if i+1 < m {
			s += ", "
		}
	}
	s += "}"
	return
}

func (e *Element) format() (s string) {
	if e.Key != nil {
		s = e.Key.Format(0) + ": "
	}
	s += e.Value.Format(0)
	return
}

func (e *Conversion) Format(n uint) (s string) {
	s = e.Type.Format(n)
	switch f := e.Type.(type) {
	case *FuncType:
		if len(f.Returns) == 0 {
			s = "(" + s + ")"
		}
	case *PtrType, *ChanType:
		s = "(" + s + ")"
	}
	s += "(" + e.Arg.Format(n) + ")"
	return
}

func (e *Call) Format(n uint) string {
	s := e.Func.Format(n)
	if e.Type == nil && len(e.Args) == 0 {
		return s + "()"
	}
	var nargs = len(e.Args)
	s += "("
	if e.Type != nil {
		s += e.Type.Format(n)
		if nargs > 0 {
			s += ", "
		}
	}
	for i := 0; i+1 < nargs; i++ {
		s += e.Args[i].Format(n) + ", "
	}
	if nargs > 0 {
		s += e.Args[nargs-1].Format(n)
	}
	if e.Ellipsis {
		s += "..."
	}
	s += ")"
	return s
}

func (e *FuncLiteral) Format(n uint) string {
	return e.Sig.Format(n) + " " + e.Body.Format(n)
}

func (e *UnaryExpr) Format(n uint) string {
	return s.TokenNames[e.Op] + e.Arg.Format(n)
}

func (e *BinaryExpr) Format(n uint) string {
	var a0, a1 string
	prec := opPrec[e.Op]
	if ex, ok := e.Arg0.(*BinaryExpr); ok {
		if opPrec[ex.Op] < prec {
			a0 = "(" + ex.Format(n) + ")"
		} else {
			a0 = ex.Format(n)
		}
	} else {
		a0 = e.Arg0.Format(n)
	}
	if ex, ok := e.Arg1.(*BinaryExpr); ok {
		if opPrec[ex.Op] < prec {
			a1 = "(" + ex.Format(n) + ")"
		} else {
			a1 = ex.Format(n)
		}
	} else {
		a1 = e.Arg1.Format(n)
	}
	return a0 + " " + s.TokenNames[e.Op] + " " + a1
}

func (b *Block) Format(n uint) string {
	return "{}"
}

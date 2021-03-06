package ast

import (
	"bytes"
	"fmt"
	"golang/scanner"
	"io"
)

type posFlagsT uint

const (
	StmtPos posFlagsT = 1 << iota
	IdentPos
	ExprPos
	TypePos
)

const indentStr = "    "

type FormatContext struct {
	buf      bytes.Buffer
	anon     bool      // output _ for annonymous parameters
	posFlags posFlagsT // output positions for the AST nodes specified in posFlags
	group    bool      // set if the declaration is aprt of a group
}

// Initializes a format context.
func (ctx *FormatContext) Init() *FormatContext {
	ctx.buf = bytes.Buffer{}
	ctx.anon = false
	return ctx
}

func (ctx *FormatContext) EmitAnnonymousParams(b bool) (r bool) {
	ctx.anon, r = b, ctx.anon
	return
}

func (ctx *FormatContext) EmitSourcePositions(f posFlagsT) (r posFlagsT) {
	ctx.posFlags, r = f, ctx.posFlags
	return
}

func (ctx *FormatContext) stmtPositions() bool {
	return (ctx.posFlags & StmtPos) != 0
}

func (ctx *FormatContext) identPositions() bool {
	return (ctx.posFlags & IdentPos) != 0
}

func (ctx *FormatContext) exprPositions() bool {
	return (ctx.posFlags & ExprPos) != 0
}

func (ctx *FormatContext) typePositions() bool {
	return (ctx.posFlags & TypePos) != 0
}

// Writes the internal buffer to an `io.Writer`
func (ctx *FormatContext) Flush(w io.Writer) (int, error) {
	return w.Write(ctx.buf.Bytes())
}

// Gets the internal buffer contents as a string.
func (ctx *FormatContext) String() string {
	return ctx.buf.String()
}

// Appends a byte slice to the internal buffer.
func (ctx *FormatContext) Write(b []byte) (int, error) {
	return ctx.buf.Write(b)
}

// Appends s tring to the internal buffer.
func (ctx *FormatContext) WriteString(s string) (int, error) {
	return ctx.buf.WriteString(s)
}

// Append bytes to the internal buffer in a variety of methods.
func (ctx *FormatContext) WriteV(n uint, args ...interface{}) {
	for _, a := range args {
		switch v := a.(type) {
		case int:
			ctx.buf.WriteString(fmt.Sprintf("%d", v))
		case string:
			ctx.buf.WriteString(v)
		case []byte:
			ctx.buf.Write(v)
		case func():
			v()
		case func(uint):
			v(n)
		case func(*FormatContext, uint):
			v(ctx, n)
		default:
			panic("invalid argument type")
		}
	}
}

// Appends whitespace for `n` levels in indentation to the internal buffer.
func (ctx *FormatContext) Indent(n uint) {
	for i := uint(0); i < n; i++ {
		ctx.buf.WriteString(indentStr)
	}
}

// Formats a source file.
func (f *File) Format(ctx *FormatContext) string {
	ctx.WriteV(0, "package ", f.PkgName, "\n")

	for _, i := range f.Imports {
		ctx.WriteString("\n")
		i.Format(ctx, 0)
	}

	for _, d := range f.Decls {
		ctx.WriteString("\n")
		d.Format(ctx, 0)
	}

	ctx.WriteString("\n")

	return ctx.String()
}

// Formats an import clause with N levels of indentation.
func (i *ImportDecl) Format(ctx *FormatContext, _ uint) {
	ctx.WriteString("import ")
	if ctx.identPositions() {
		ctx.WriteV(0, "/* #", i.Off, " */")
	}
	if len(i.Name) > 0 {
		ctx.WriteV(0, i.Name, " ")
	}
	ctx.Write(i.Path)
}

// Formats an error node
func (e *Error) Format(ctx *FormatContext, _ uint) {
	ctx.WriteString("<error>")
}

// Formats an identifier.
func formatIdent(ctx *FormatContext, off int, name string) {
	if ctx.identPositions() {
		ctx.WriteV(0, "/* #", off, " */", name)
	} else {
		ctx.WriteString(name)
	}
}

// Formats a type declaration.
func (t *TypeDecl) Format(ctx *FormatContext, n uint) {
	if !ctx.group {
		ctx.WriteString("type ")
	}
	ctx.WriteV(n, t.Name, " ", t.Type.Format)
}

func (g *TypeDeclGroup) Format(ctx *FormatContext, n uint) {
	ctx.WriteV(n, "type (")
	ctx.group = true
	for _, t := range g.Types {
		ctx.WriteV(n+1, "\n", ctx.Indent, t.Format)
	}
	ctx.WriteV(n, "\n", ctx.Indent, ")")
	ctx.group = false
}

// Formats a constant declaration.
func (c *ConstDecl) Format(ctx *FormatContext, n uint) {
	if !ctx.group {
		ctx.WriteString("const ")
	}
	formatIdent(ctx, c.Names[0].Off, c.Names[0].Name)
	for i := 1; i < len(c.Names); i++ {
		ctx.WriteString(", ")
		formatIdent(ctx, c.Names[i].Off, c.Names[i].Name)
	}
	if c.Type != nil {
		ctx.WriteString(" ")
		c.Type.Format(ctx, n+1)
	}
	if k := len(c.Values); k > 0 {
		ctx.WriteString(" = ")
		c.Values[0].Format(ctx, n+1)
		for i := 1; i < k; i++ {
			ctx.WriteString(", ")
			c.Values[i].Format(ctx, n+1)
		}
	}
}

func (g *ConstDeclGroup) Format(ctx *FormatContext, n uint) {
	ctx.WriteV(n, "const (")
	ctx.group = true
	for _, t := range g.Consts {
		ctx.WriteV(n+1, "\n", ctx.Indent, t.Format)
	}
	ctx.WriteV(n, "\n", ctx.Indent, ")")
	ctx.group = false
}

// Formats a variable declaration.
func (c *VarDecl) Format(ctx *FormatContext, n uint) {
	if !ctx.group {
		ctx.WriteString("var ")
	}
	formatIdent(ctx, c.Names[0].Off, c.Names[0].Name)
	for i := 1; i < len(c.Names); i++ {
		ctx.WriteString(", ")
		formatIdent(ctx, c.Names[i].Off, c.Names[i].Name)
	}
	if c.Type != nil {
		ctx.WriteString(" ")
		c.Type.Format(ctx, n+1)
	}
	if k := len(c.Init); k > 0 {
		ctx.WriteString(" = ")
		c.Init[0].Format(ctx, n+1)
		for i := 1; i < k; i++ {
			ctx.WriteString(", ")
			c.Init[i].Format(ctx, n+1)
		}
	}
}

func (g *VarDeclGroup) Format(ctx *FormatContext, n uint) {
	ctx.WriteV(n, "var (")
	ctx.group = true
	for _, t := range g.Vars {
		ctx.WriteV(n+1, "\n", ctx.Indent, t.Format)
	}
	ctx.WriteV(n, "\n", ctx.Indent, ")")
	ctx.group = false
}

// Formats a function declaration.
func (f *FuncDecl) Format(ctx *FormatContext, n uint) {
	ctx.WriteString("func")
	if rcv := f.Func.Recv; rcv != nil {
		ctx.WriteString(" (")
		if len(rcv.Name) > 0 {
			formatIdent(ctx, rcv.Off, rcv.Name)
			ctx.WriteString(" ")
		}
		rcv.Type.Format(ctx, n+1)
		ctx.WriteString(")")
	}
	ctx.WriteString(" ")
	formatIdent(ctx, f.Off, f.Name)
	formatSignature(ctx, f.Func.Sig, n)
	if f.Func.Blk != nil {
		ctx.WriteString(" ")
		f.Func.Blk.Format(ctx, n)
	}
}

//
// Formats types.
//
func (t *QualifiedId) Format(ctx *FormatContext, n uint) {
	if ctx.identPositions() {
		ctx.WriteV(n, "/* #", t.Off, " */")
	}
	if len(t.Pkg) > 0 {
		ctx.WriteV(0, t.Pkg, ".")
	}
	ctx.WriteString(t.Id)
}

func (t *BuiltinType) Format(ctx *FormatContext, _ uint) {
	ctx.WriteString(t.String())
}

func (t *ArrayType) Format(ctx *FormatContext, n uint) {
	if ctx.typePositions() {
		ctx.WriteV(n, "/* #", t.Off, " */")
	}
	if t.Dim == nil {
		ctx.WriteString("[...]")
	} else {
		ctx.WriteV(n+1, "[", t.Dim.Format, "]")
	}
	t.Elt.Format(ctx, n)
}

func (t *SliceType) Format(ctx *FormatContext, n uint) {
	if ctx.typePositions() {
		ctx.WriteV(n, "/* #", t.Off, " */")
	}
	ctx.WriteString("[]")
	t.Elt.Format(ctx, n)
}

func (t *PtrType) Format(ctx *FormatContext, n uint) {
	if ctx.typePositions() {
		ctx.WriteV(n, "/* #", t.Off, " */")
	}
	ctx.WriteString("*")
	t.Base.Format(ctx, n)
}

func (t *MapType) Format(ctx *FormatContext, n uint) {
	if ctx.typePositions() {
		ctx.WriteV(n, "/* #", t.Off, " */")
	}
	ctx.WriteV(n, "map[", t.Key.Format, "]", t.Elt.Format)
}

func (t *ChanType) Format(ctx *FormatContext, n uint) {
	if ctx.typePositions() {
		ctx.WriteV(n, "/* #", t.Off, " */")
	}
	if t.Send {
		ctx.WriteString("chan")
	} else {
		ctx.WriteString("<-chan")
	}
	if !t.Recv {
		ctx.WriteString("<- ")
		t.Elt.Format(ctx, n)
	} else if ch, ok := t.Elt.(*ChanType); ok && !ch.Send {
		ctx.WriteV(n, " (", ch.Format, ")")
	} else {
		ctx.WriteString(" ")
		t.Elt.Format(ctx, n)
	}
}

func (t *StructType) Format(ctx *FormatContext, n uint) {
	if ctx.typePositions() {
		ctx.WriteV(n, "/* #", t.Off, " */")
	}
	if len(t.Fields) == 0 {
		ctx.WriteString("struct{}")
	} else {
		ctx.WriteString("struct {\n")
		ctx.Indent(n + 1)
		for i := range t.Fields {
			f := &t.Fields[i]
			if len(f.Name) > 0 {
				formatIdent(ctx, f.Off, f.Name)
			}
			if f.Type == nil {
				ctx.WriteString(", ")
			} else {
				if len(f.Name) > 0 {
					ctx.WriteString(" ")
				}
				f.Type.Format(ctx, n+1)
				if len(f.Tag) > 0 {
					ctx.WriteV(0, " `", f.Tag, "`")
				}
				ctx.WriteString("\n")
				if i+1 < len(t.Fields) {
					ctx.Indent(n + 1)
				}
			}
		}
		ctx.Indent(n)
		ctx.WriteString("}")
	}
}

func (t *TupleType) Format(ctx *FormatContext, n uint) {
	if ctx.typePositions() {
		ctx.WriteV(n, "/* #", t.Off, " */")
	}
	ctx.WriteString("(")
	m := len(t.Types)
	for i := 0; i+1 < m; i++ {
		t.Types[i].Format(ctx, n)
		ctx.WriteString(", ")
	}
	t.Types[m-1].Format(ctx, n)
	ctx.WriteString(")")

}

func (t *FuncType) Format(ctx *FormatContext, n uint) {
	if ctx.typePositions() {
		ctx.WriteV(n, "/* #", t.Off, " */")
	}
	ctx.WriteString("func")
	formatSignature(ctx, t, n)
}

func formatSignature(ctx *FormatContext, t *FuncType, n uint) {
	formatParams(ctx, n, t.Params, t.Var)

	k := len(t.Returns)
	if k > 0 {
		ctx.WriteString(" ")
	}
	if k == 1 && len(t.Returns[0].Name) == 0 {
		t.Returns[0].Type.Format(ctx, n)
	} else if k > 0 {
		formatParams(ctx, n, t.Returns, false)
	}
}

func formatParams(ctx *FormatContext, n uint, ps []Param, variadic bool) {
	ctx.WriteString("(")
	if len(ps) > 0 {
		formatParam(ctx, n, &ps[0], variadic && len(ps) == 1)
		for i := 1; i < len(ps); i++ {
			ctx.WriteString(", ")
			formatParam(ctx, n, &ps[i], variadic && i+1 == len(ps))
		}
	}
	ctx.WriteString(")")
}

func formatParam(ctx *FormatContext, n uint, p *Param, variadic bool) {
	if len(p.Name) > 0 || ctx.anon {
		if len(p.Name) > 0 {
			formatIdent(ctx, p.Off, p.Name)
		} else {
			ctx.WriteString("_")
		}
		if p.Type != nil {
			ctx.WriteString(" ")
		}
	}
	if p.Type != nil {
		if variadic {
			ctx.WriteString("...")
		}
		p.Type.Format(ctx, n+1)
	}
}

func (t *InterfaceType) Format(ctx *FormatContext, n uint) {
	if ctx.typePositions() {
		ctx.WriteV(n, "/* #", t.Off, " */")
	}
	if len(t.Methods) == 0 {
		ctx.WriteString("interface{}")
	} else {
		ctx.WriteString("interface {")
		for _, d := range t.Embedded {
			ctx.WriteV(n+1, "\n", ctx.Indent, d.Format)
		}
		for _, f := range t.Methods {
			ctx.WriteV(n+1, "\n", ctx.Indent)
			formatIdent(ctx, f.Off, f.Name)
			formatSignature(ctx, f.Func.Sig, n+1)
		}
		ctx.WriteV(n, "\n", ctx.Indent, "}")
	}
}

//
// Format expressions.
//

// FIXME: consider escaping string/rune literals
func (c *ConstValue) Format(ctx *FormatContext, _ uint) {
	if ctx.exprPositions() {
		ctx.WriteV(0, "/* #", c.Off, " */")
	}
	ctx.WriteString(c.String())
}

func (x *OperandName) Format(ctx *FormatContext, _ uint) {
	var name string
	switch s := x.Decl.(type) {
	case *Const:
		name = s.Name
	case *Var:
		name = s.Name
	case *FuncDecl:
		name = s.Name
	default:
		panic("not reached")
	}
	ctx.WriteString(name)
}

func (e *TypeAssertion) Format(ctx *FormatContext, n uint) {
	e.X.Format(ctx, n)
	if e.ATyp == nil {
		ctx.WriteString(".(type)")
	} else {
		ctx.WriteV(n, ".(", e.ATyp.Format, ")")
	}
}

func (e *Selector) Format(ctx *FormatContext, n uint) {
	ctx.WriteV(n, e.X.Format, ".", e.Id)
}

func (e *IndexExpr) Format(ctx *FormatContext, n uint) {
	ctx.WriteV(n, e.X.Format, "[", e.I.Format, "]")
}

func (e *SliceExpr) Format(ctx *FormatContext, n uint) {
	e.X.Format(ctx, n)
	ctx.WriteString("[")
	if e.Lo != nil {
		e.Lo.Format(ctx, n)
		ctx.WriteString(" :")
	} else {
		ctx.WriteString(":")
	}
	if e.Hi != nil {
		ctx.WriteString(" ")
		e.Hi.Format(ctx, n)
	}
	if e.Cap != nil {
		ctx.WriteString(" : ")
		e.Cap.Format(ctx, n)
	}
	ctx.WriteString("]")
}

func (e *MethodExpr) Format(ctx *FormatContext, n uint) {
	ctx.WriteV(n, "(", e.RTyp.Format, ").", e.Id)
}

func (x *ParensExpr) Format(ctx *FormatContext, n uint) {
	if ctx.exprPositions() {
		ctx.WriteV(n, "/* #", x.Off, " */")
	}
	ctx.WriteV(n, "(", x.X.Format, ")")
}

func (e *CompLiteral) Format(ctx *FormatContext, n uint) {
	if e.Typ != nil {
		e.Typ.Format(ctx, n)
	}
	ctx.WriteString("{")
	m := len(e.Elts)
	if m > 0 {
		e.Elts[0].format(ctx)
		for i := 1; i < m; i++ {
			ctx.WriteString(", ")
			e.Elts[i].format(ctx)
		}
	}
	ctx.WriteString("}")
}

func (e *KeyedElement) format(ctx *FormatContext) {
	if e.Key != nil {
		e.Key.Format(ctx, 0)
		ctx.WriteString(": ")
	}
	e.Elt.Format(ctx, 0)
}

func (e *Conversion) Format(ctx *FormatContext, n uint) {
	switch f := e.Typ.(type) {
	case *FuncType:
		if len(f.Returns) == 0 {
			ctx.WriteV(n, "(", e.Typ.Format, ")")
		} else {
			e.Typ.Format(ctx, n)
		}
	case *PtrType, *ChanType:
		ctx.WriteV(n, "(", e.Typ.Format, ")")
	default:
		e.Typ.Format(ctx, n)
	}
	ctx.WriteV(n, "(", e.X.Format, ")")
}

func (e *Call) Format(ctx *FormatContext, n uint) {
	e.Func.Format(ctx, n)
	if e.ATyp == nil && len(e.Xs) == 0 {
		ctx.WriteString("()")
	} else {
		var nargs = len(e.Xs)
		ctx.WriteString("(")
		if e.ATyp != nil {
			e.ATyp.Format(ctx, n)
			if nargs > 0 {
				ctx.WriteString(", ")
			}
		}
		if nargs > 0 {
			e.Xs[0].Format(ctx, n)
			for i := 1; i < nargs; i++ {
				ctx.WriteString(", ")
				e.Xs[i].Format(ctx, n)
			}
		}
		if e.Dots {
			ctx.WriteString("...")
		}
		ctx.WriteString(")")
	}
}

func (e *Func) Format(ctx *FormatContext, n uint) {
	ctx.WriteV(n, e.Sig.Format, " ", e.Blk.Format)
}

func (e *UnaryExpr) Format(ctx *FormatContext, n uint) {
	if ctx.exprPositions() {
		ctx.WriteV(n, "/* #", e.Off, " */")
	}
	ctx.WriteString(scanner.TokenNames[uint(e.Op)])
	e.X.Format(ctx, n)
}

func (e *BinaryExpr) Format(ctx *FormatContext, n uint) {
	ctx.WriteV(n, e.X.Format, " ", scanner.TokenNames[uint(e.Op)], " ", e.Y.Format)
}

func (b *Block) Format(ctx *FormatContext, n uint) {
	grp := ctx.group
	ctx.group = false
	ctx.WriteString("{")
	empty := true
	for i := range b.Body {
		if s, ok := b.Body[i].(*EmptyStmt); ok {
			if ctx.stmtPositions() {
				empty = false
				ctx.WriteV(n+1, "\n", ctx.Indent, s.Format)
			}
		} else {
			empty = false
			if s, ok := b.Body[i].(*Label); ok {
				s.Format(ctx, n)
			} else {
				ctx.WriteV(n+1, "\n", ctx.Indent, b.Body[i].Format)
			}
		}
	}
	if !empty {
		ctx.WriteV(n, "\n", ctx.Indent)
	}
	ctx.WriteString("}")
	ctx.group = grp
}

func (e *EmptyStmt) Format(ctx *FormatContext, n uint) {
	if ctx.stmtPositions() {
		ctx.WriteV(n, "/* #", e.Off, " */")
	}
}

func (s *Label) Format(ctx *FormatContext, n uint) {
	if ctx.stmtPositions() {
		ctx.WriteV(n, "\n", ctx.Indent, "/* #", s.Off, " */", s.Label, ":")
	} else {
		ctx.WriteV(n, "\n", ctx.Indent, s.Label, ":")
	}

	if e, ok := s.Stmt.(*EmptyStmt); ok {
		if ctx.stmtPositions() {
			ctx.WriteV(n+1, "\n", ctx.Indent, e.Format)
		}
	} else if l, ok := s.Stmt.(*Label); ok {
		l.Format(ctx, n)
	} else {
		ctx.WriteV(n+1, "\n", ctx.Indent, s.Stmt.Format)
	}
}

func (g *GoStmt) Format(ctx *FormatContext, n uint) {
	if ctx.stmtPositions() {
		ctx.WriteV(n, "/* #", g.Off, " */")
	}
	ctx.WriteString("go ")
	g.X.Format(ctx, n)
}

func (r *ReturnStmt) Format(ctx *FormatContext, n uint) {
	if ctx.stmtPositions() {
		ctx.WriteV(n, "/* #", r.Off, " */")
	}
	ctx.WriteString("return")
	if len(r.Xs) > 0 {
		ctx.WriteString(" ")
		r.Xs[0].Format(ctx, n)
		for i := 1; i < len(r.Xs); i++ {
			ctx.WriteString(", ")
			r.Xs[i].Format(ctx, n)
		}
	}
}

func (b *BreakStmt) Format(ctx *FormatContext, _ uint) {
	if ctx.stmtPositions() {
		ctx.WriteV(0, "/* #", b.Off, " */")
	}
	ctx.WriteString("break")
	if len(b.Label) > 0 {
		ctx.WriteV(0, " ", b.Label)
	}
}

func (c *ContinueStmt) Format(ctx *FormatContext, _ uint) {
	if ctx.stmtPositions() {
		ctx.WriteV(0, "/* #", c.Off, " */")
	}
	ctx.WriteString("continue")
	if len(c.Label) > 0 {
		ctx.WriteV(0, " ", c.Label)
	}
}

func (b *GotoStmt) Format(ctx *FormatContext, _ uint) {
	if ctx.stmtPositions() {
		ctx.WriteV(0, "/* #", b.Off, " */")
	}
	ctx.WriteV(0, "goto ", b.Label)
}

func (b *FallthroughStmt) Format(ctx *FormatContext, _ uint) {
	if ctx.stmtPositions() {
		ctx.WriteV(0, "/* #", b.Off, " */")
	}
	ctx.WriteString("fallthrough")
}

func (s *SendStmt) Format(ctx *FormatContext, n uint) {
	ctx.WriteV(n, s.Ch.Format, " <- ", s.X.Format)
}

func (s *RecvStmt) Format(ctx *FormatContext, n uint) {
	if s.X != nil {
		s.X.Format(ctx, n)
		if s.Y != nil {
			ctx.WriteV(0, ", ", s.Y.Format)
		}
		if s.Op == NOP {
			ctx.WriteString(" = ")
		} else {
			ctx.WriteString(" := ")
		}
	}
	s.Rcv.Format(ctx, n)
}

func (s *IncStmt) Format(ctx *FormatContext, n uint) {
	s.X.Format(ctx, n)
	ctx.WriteString("++")
}

func (s *DecStmt) Format(ctx *FormatContext, n uint) {
	s.X.Format(ctx, n)
	ctx.WriteString("--")
}

func (a *AssignStmt) Format(ctx *FormatContext, n uint) {
	m := len(a.LHS)
	if m > 0 {
		a.LHS[0].Format(ctx, n)
		for i := 1; i < m; i++ {
			ctx.WriteString(", ")
			a.LHS[i].Format(ctx, n)
		}
	}
	if a.Op == DCL {
		ctx.WriteString(" := ")
	} else if a.Op == NOP {
		ctx.WriteString(" = ")
	} else {
		ctx.WriteV(0, " ", scanner.TokenNames[a.Op], "= ")
	}
	m = len(a.RHS)
	if m > 0 {
		a.RHS[0].Format(ctx, n)
		for i := 1; i < m; i++ {
			ctx.WriteString(", ")
			a.RHS[i].Format(ctx, n)
		}
	}
}

func (e *ExprStmt) Format(ctx *FormatContext, n uint) {
	e.X.Format(ctx, n)
}

func (i *IfStmt) Format(ctx *FormatContext, n uint) {
	if ctx.stmtPositions() {
		ctx.WriteV(n, "/* #", i.Off, " */")
	}
	ctx.WriteString("if ")
	if i.Init != nil {
		i.Init.Format(ctx, n)
		ctx.WriteString("; ")
	}
	i.Cond.Format(ctx, n)
	ctx.WriteString(" ")
	i.Then.Format(ctx, n)
	if i.Else != nil {
		ctx.WriteString(" else ")
		i.Else.Format(ctx, n)
	}
}

func (f *ForStmt) Format(ctx *FormatContext, n uint) {
	if ctx.stmtPositions() {
		ctx.WriteV(n, "/* #", f.Off, " */")
	}
	if f.Init == nil && f.Cond == nil && f.Post == nil {
		ctx.WriteString("for ")
		f.Blk.Format(ctx, n)
		return
	}
	if f.Init == nil && f.Post == nil {
		ctx.WriteV(n, "for ", f.Cond.Format, " ", f.Blk.Format)
		return
	}
	ctx.WriteString("for ")
	if f.Init != nil {
		f.Init.Format(ctx, n+1)
	}
	if f.Cond != nil {
		ctx.WriteV(n, "; ", f.Cond.Format, ";")
	} else {
		ctx.WriteString("; ;")
	}
	if f.Post != nil {
		ctx.WriteString(" ")
		f.Post.Format(ctx, n)
	}
	ctx.WriteString(" ")
	f.Blk.Format(ctx, n)
}

func (f *ForRangeStmt) Format(ctx *FormatContext, n uint) {
	if ctx.stmtPositions() {
		ctx.WriteV(n, "/* #", f.Off, " */")
	}
	ctx.WriteString("for ")
	if m := len(f.LHS); m > 0 {
		f.LHS[0].Format(ctx, n)
		for i := 1; i < m; i++ {
			ctx.WriteString(", ")
			f.LHS[i].Format(ctx, n)
		}
		if f.Op == NOP {
			ctx.WriteString(" = ")
		} else {
			ctx.WriteString(" := ")
		}
	}
	ctx.WriteV(n, "range ", f.Range.Format, " ", f.Blk.Format)
}

func (d *DeferStmt) Format(ctx *FormatContext, n uint) {
	if ctx.stmtPositions() {
		ctx.WriteV(n, "/* #", d.Off, " */")
	}
	ctx.WriteString("defer ")
	d.X.Format(ctx, n)
}

func (s *ExprSwitchStmt) Format(ctx *FormatContext, n uint) {
	if ctx.stmtPositions() {
		ctx.WriteV(n, "/* #", s.Off, " */")
	}
	ctx.WriteString("switch")
	if s.Init != nil {
		ctx.WriteV(n, " ", s.Init.Format, ";")
	}
	if s.X != nil {
		ctx.WriteString(" ")
		s.X.Format(ctx, n)
	}
	if len(s.Cases) == 0 {
		ctx.WriteString(" {}")
		return
	}
	ctx.WriteString(" {")
	for _, c := range s.Cases {
		if c.Xs == nil {
			ctx.WriteV(n, "\n", ctx.Indent, "default:")
		} else {
			ctx.WriteV(n, "\n", ctx.Indent, "case ")
			c.Xs[0].Format(ctx, n)
			for i := 1; i < len(c.Xs); i++ {
				ctx.WriteString(", ")
				c.Xs[i].Format(ctx, 0)
			}
			ctx.WriteString(":")
		}
		for _, s := range c.Blk.Body {
			if _, ok := s.(*EmptyStmt); ok {
				ctx.WriteV(n+1, "\n", ctx.Indent, s.Format)
			} else {
				ctx.WriteV(n+1, "\n", ctx.Indent, s.Format)
			}
		}
	}
	ctx.WriteV(n, "\n", ctx.Indent)
	ctx.WriteString("}")
}

func (s *TypeSwitchStmt) Format(ctx *FormatContext, n uint) {
	if ctx.stmtPositions() {
		ctx.WriteV(n, "/* #", s.Off, " */")
	}
	ctx.WriteString("switch")
	if s.Init != nil {
		ctx.WriteV(n, " ", s.Init.Format, ";")
	}
	if len(s.Id) > 0 {
		ctx.WriteV(0, " ", s.Id, " :=")
	}
	ctx.WriteV(0, " ", s.X.Format, ".(type)")
	if len(s.Cases) == 0 {
		ctx.WriteString(" {}")
		return
	}
	ctx.WriteString(" {")
	for _, c := range s.Cases {
		if c.Types == nil {
			ctx.WriteV(n, "\n", ctx.Indent, "default:")
		} else {
			ctx.WriteV(n, "\n", ctx.Indent, "case ")
			c.Types[0].Format(ctx, n)
			for i := 1; i < len(c.Types); i++ {
				ctx.WriteV(0, ", ", c.Types[i].Format)
			}
			ctx.WriteString(":")
		}
		for _, s := range c.Blk.Body {
			if _, ok := s.(*EmptyStmt); !ok {
				ctx.WriteV(n+1, "\n", ctx.Indent, s.Format)
			}
		}
	}
	ctx.WriteV(n, "\n", ctx.Indent)
	ctx.WriteString("}")
}

func (s *SelectStmt) Format(ctx *FormatContext, n uint) {
	if ctx.stmtPositions() {
		ctx.WriteV(n, "/* #", s.Off, " */")
	}
	ctx.WriteString("select ")
	if len(s.Comms) == 0 {
		ctx.WriteString("{}")
		return
	}
	ctx.WriteString("{")
	for _, c := range s.Comms {
		if c.Comm == nil {
			ctx.WriteV(n, "\n", ctx.Indent, "default:")
		} else {
			ctx.WriteV(n, "\n", ctx.Indent, "case ", c.Comm.Format, ":")
		}
		for _, s := range c.Blk.Body {
			if _, ok := s.(*EmptyStmt); !ok {
				ctx.WriteV(n+1, "\n", ctx.Indent, s.Format)
			}
		}
	}
	ctx.WriteV(n, "\n", ctx.Indent)
	ctx.WriteString("}")
}

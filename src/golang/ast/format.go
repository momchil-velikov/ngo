package ast

import (
	"bytes"
	"fmt"
	s "golang/scanner"
	"io"
)

type posFlagsT uint

const (
	StmtPos posFlagsT = 1 << iota
	BlockPos
	IdentPos
	ExprPos
	TypePos
	DeclPos
)

const indentStr = "    "

type FormatContext struct {
	buf      bytes.Buffer
	anon     bool      // output _ for annonymous parameters
	posFlags posFlagsT // output positions for the AST nodes specified in posFlags
	group    string    // declaration kind for declaration groups
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

func (ctx *FormatContext) blockPositions() bool {
	return (ctx.posFlags & BlockPos) != 0
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

func (ctx *FormatContext) declPositions() bool {
	return (ctx.posFlags & DeclPos) != 0
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

// Append bytes to the internal buffer ina variety of methods.
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
	ctx.WriteV(0, "package ", f.Package, "\n")

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

// Formats a declaration group.
func (g *DeclGroup) Format(ctx *FormatContext, n uint) {
	switch g.Kind {
	case s.IMPORT:
		ctx.group = "import"
	case s.TYPE:
		ctx.group = "type"
	case s.CONST:
		ctx.group = "const"
	case s.VAR:
		ctx.group = "var"
	default:
		panic("invalid declaration group kind")
	}
	if ctx.declPositions() {
		ctx.WriteV(n, "/* #", g.Off, " */", ctx.group, " (")
	} else {
		ctx.WriteV(n, ctx.group, " (")
	}
	for _, d := range g.Decls {
		ctx.WriteV(n+1, "\n", ctx.Indent, d.Format)
	}
	ctx.WriteV(n, "\n", ctx.Indent, ")")
	ctx.group = ""
}

// Formats an import clause with N levels of indentation.
func (i *Import) Format(ctx *FormatContext, _ uint) {
	if len(ctx.group) == 0 {
		if ctx.declPositions() {
			ctx.WriteV(0, "/* #", i.Off, " */import ")
		} else {
			ctx.WriteString("import ")
		}
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

// Formats a type declaration.
func (t *TypeDecl) Format(ctx *FormatContext, n uint) {
	if len(ctx.group) == 0 {
		if ctx.declPositions() {
			ctx.WriteV(0, "/* #", t.Off, " */type ")
		} else {
			ctx.WriteString("type ")
		}
	}
	ctx.WriteV(n, t.Name, " ", t.Type.Format)
}

// Formats a constant declaration.
func (c *ConstDecl) Format(ctx *FormatContext, n uint) {
	if len(ctx.group) == 0 {
		ctx.WriteString("const ")
	}
	c.Names[0].Format(ctx, n)
	for i := 1; i < len(c.Names); i++ {
		ctx.WriteV(0, ", ", c.Names[i].Format)
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

// Formats a variable declaration.
func (c *VarDecl) Format(ctx *FormatContext, n uint) {
	if len(ctx.group) == 0 {
		ctx.WriteString("var ")
	}
	c.Names[0].Format(ctx, n)
	for i := 1; i < len(c.Names); i++ {
		ctx.WriteV(0, ", ", c.Names[i].Format)
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

// Formats a function declaration.
func (f *FuncDecl) Format(ctx *FormatContext, n uint) {
	if ctx.declPositions() {
		ctx.WriteV(0, "/* #", f.Off, " */")
	}
	ctx.WriteString("func")
	if f.Recv != nil {
		ctx.WriteString(" ")
		f.Recv.Format(ctx, 0)
	}
	ctx.WriteString(" ")
	f.Name.Format(ctx, n)
	formatSignature(ctx, f.Sig, n)
	if f.Blk != nil {
		ctx.WriteString(" ")
		f.Blk.Format(ctx, n)
	}
}

// Formats a method receiver.
func (r *Receiver) Format(ctx *FormatContext, n uint) {
	ctx.WriteString("(")
	if r.Name != nil {
		r.Name.Format(ctx, n)
		ctx.WriteString(" ")
	}
	r.Type.Format(ctx, n+1)
	ctx.WriteString(")")
}

//
// Formats types.
//
func (t *Ident) Format(ctx *FormatContext, n uint) {
	if ctx.identPositions() {
		ctx.WriteV(n, "/* #", t.Off, " */")
	}
	if len(t.Pkg) > 0 {
		ctx.WriteV(0, t.Pkg, ".")
	}
	ctx.WriteString(t.Id)
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
		for _, f := range t.Fields {
			ctx.Indent(n + 1)
			if m := len(f.Names); m > 0 {
				f.Names[0].Format(ctx, n)
				for i := 1; i < m; i++ {
					ctx.WriteString(", ")
					f.Names[i].Format(ctx, n)
				}
				ctx.WriteString(" ")
			}
			f.Type.Format(ctx, n+1)
			if f.Tag != nil {
				ctx.WriteV(0, " ", f.Tag)
			}
			ctx.WriteString("\n")
		}
		ctx.Indent(n)
		ctx.WriteString("}")
	}
}

func (t *FuncType) Format(ctx *FormatContext, n uint) {
	if ctx.typePositions() {
		ctx.WriteV(n, "/* #", t.Off, " */")
	}
	ctx.WriteString("func")
	formatSignature(ctx, t, n)
}

func formatSignature(ctx *FormatContext, t *FuncType, n uint) {
	formatParams(ctx, t.Params, n)

	k := len(t.Returns)
	if k > 0 {
		ctx.WriteString(" ")
	}
	if k == 1 && len(t.Returns[0].Names) == 0 {
		t.Returns[0].Type.Format(ctx, n)
	} else if k > 0 {
		formatParams(ctx, t.Returns, n)
	}
}

func formatParams(ctx *FormatContext, ps []*ParamDecl, n uint) {
	ctx.WriteString("(")
	if len(ps) > 0 {
		ps[0].Format(ctx, 0)
		for i := 1; i < len(ps); i++ {
			ctx.WriteString(", ")
			ps[i].Format(ctx, 0)
		}
	}
	ctx.WriteString(")")
}

func (p *ParamDecl) Format(ctx *FormatContext, n uint) {
	if m := len(p.Names); m > 0 {
		p.Names[0].Format(ctx, n)
		for i := 1; i < m; i++ {
			ctx.WriteString(", ")
			p.Names[i].Format(ctx, n)
		}
		ctx.WriteString(" ")
	} else if ctx.anon {
		ctx.WriteString("_ ")
	}
	if p.Var {
		ctx.WriteString("...")
	}
	p.Type.Format(ctx, n+1)
}

func (t *InterfaceType) Format(ctx *FormatContext, n uint) {
	if ctx.typePositions() {
		ctx.WriteV(n, "/* #", t.Off, " */")
	}
	if len(t.Methods) == 0 {
		ctx.WriteString("interface{}")
	} else {
		ctx.WriteString("interface {")
		for _, m := range t.Methods {
			if m.Name == nil {
				ctx.WriteV(n+1, "\n", ctx.Indent, m.Type.Format)
			} else {
				ctx.WriteV(n+1, "\n", ctx.Indent, m.Name.Format)
				formatSignature(ctx, m.Type.(*FuncType), n+1)
			}
		}
		ctx.WriteV(n, "\n", ctx.Indent, "}")
	}
}

//
// Format expressions.
//
func (e *Literal) Format(ctx *FormatContext, _ uint) {
	if ctx.exprPositions() {
		ctx.WriteV(0, "/* #", e.Off, " */")
	}
	switch e.Kind {
	case s.INTEGER, s.FLOAT:
		ctx.Write(e.Value)
	case s.RUNE:
		ctx.WriteV(0, "'", e.Value, "'")
	case s.IMAGINARY:
		ctx.WriteV(0, e.Value, "i")
	case s.STRING:
		ctx.Write(e.Value)
	default:
		panic("invalid literal")
	}
}

func (e *TypeAssertion) Format(ctx *FormatContext, n uint) {
	e.X.Format(ctx, n)
	if e.Type == nil {
		ctx.WriteString(".(type)")
	} else {
		ctx.WriteV(n, ".(", e.Type.Format, ")")
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
	ctx.WriteV(n, "(", e.Type.Format, ").", e.Id)
}

func (x *ParensExpr) Format(ctx *FormatContext, n uint) {
	if ctx.exprPositions() {
		ctx.WriteV(n, "/* #", x.Off, " */")
	}
	ctx.WriteV(n, "(", x.X.Format, ")")
}

func (e *CompLiteral) Format(ctx *FormatContext, n uint) {
	if e.Type != nil {
		e.Type.Format(ctx, n)
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

func (e *Element) format(ctx *FormatContext) {
	if e.Key != nil {
		e.Key.Format(ctx, 0)
		ctx.WriteString(": ")
	}
	e.Value.Format(ctx, 0)
}

func (e *Conversion) Format(ctx *FormatContext, n uint) {
	switch f := e.Type.(type) {
	case *FuncType:
		if len(f.Returns) == 0 {
			ctx.WriteV(n, "(", e.Type.Format, ")")
		} else {
			e.Type.Format(ctx, n)
		}
	case *PtrType, *ChanType:
		ctx.WriteV(n, "(", e.Type.Format, ")")
	default:
		e.Type.Format(ctx, n)
	}
	ctx.WriteV(n, "(", e.X.Format, ")")
}

func (e *Call) Format(ctx *FormatContext, n uint) {
	e.Func.Format(ctx, n)
	if e.Type == nil && len(e.Xs) == 0 {
		ctx.WriteString("()")
	} else {
		var nargs = len(e.Xs)
		ctx.WriteString("(")
		if e.Type != nil {
			e.Type.Format(ctx, n)
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
		if e.Ell {
			ctx.WriteString("...")
		}
		ctx.WriteString(")")
	}
}

func (e *FuncLiteral) Format(ctx *FormatContext, n uint) {
	ctx.WriteV(n, e.Sig.Format, " ", e.Blk.Format)
}

func (e *UnaryExpr) Format(ctx *FormatContext, n uint) {
	if ctx.exprPositions() {
		ctx.WriteV(n, "/* #", e.Off, " */")
	}
	ctx.WriteString(s.TokenNames[e.Op])
	e.X.Format(ctx, n)
}

func (e *BinaryExpr) Format(ctx *FormatContext, n uint) {
	ctx.WriteV(n, e.X.Format, " ", s.TokenNames[e.Op], " ", e.Y.Format)
}

func (b *Block) Format(ctx *FormatContext, n uint) {
	if ctx.blockPositions() {
		ctx.WriteV(n, "/* #", b.Begin, " */")
	}
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
			if s, ok := b.Body[i].(*LabeledStmt); ok {
				s.Format(ctx, n)
			} else {
				ctx.WriteV(n+1, "\n", ctx.Indent, b.Body[i].Format)
			}
		}
	}
	if ctx.blockPositions() {
		if !empty {
			ctx.WriteV(n, "\n", ctx.Indent)
		}
		ctx.WriteV(n, "/* #", b.End, " */}")
	} else {
		if !empty {
			ctx.WriteV(n, "\n", ctx.Indent)
		}
		ctx.WriteString("}")
	}
}

func (e *EmptyStmt) Format(ctx *FormatContext, n uint) {
	if ctx.stmtPositions() {
		ctx.WriteV(n, "/* #", e.Off, " */")
	}
}

func (s *LabeledStmt) Format(ctx *FormatContext, n uint) {
	if ctx.stmtPositions() {
		ctx.WriteV(n, "\n", ctx.Indent, "/* #", s.Off, " */", s.Label, ":")
	} else {
		ctx.WriteV(n, "\n", ctx.Indent, s.Label, ":")
	}
	ctx.WriteV(n+1, "\n", ctx.Indent, s.Stmt.Format)
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
	ctx.WriteV(0, " ", s.TokenNames[a.Op], " ")
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
		ctx.WriteV(0, " ", s.TokenNames[f.Op], " ")
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
		if ctx.blockPositions() {
			ctx.WriteV(n, " /* #", s.Begin, " */{/* #", s.End, " */}")
		} else {
			ctx.WriteString(" {}")
		}
		return
	}
	if ctx.blockPositions() {
		ctx.WriteV(n, " /* #", s.Begin, " */{")
	} else {
		ctx.WriteString(" {")
	}
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
		for _, s := range c.Body {
			if _, ok := s.(*EmptyStmt); ok {
				ctx.WriteV(n+1, "\n", ctx.Indent, s.Format)
			} else {
				ctx.WriteV(n+1, "\n", ctx.Indent, s.Format)
			}
		}
	}
	ctx.WriteV(n, "\n", ctx.Indent)
	if ctx.blockPositions() {
		ctx.WriteV(n, "/* #", s.End, " */}")
	} else {
		ctx.WriteString("}")
	}
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
		if ctx.blockPositions() {
			ctx.WriteV(n, " /* #", s.Begin, " */{/* #", s.End, " */}")
		} else {
			ctx.WriteString(" {}")
		}
		return
	}
	if ctx.blockPositions() {
		ctx.WriteV(n, " /* #", s.Begin, " */{")
	} else {
		ctx.WriteString(" {")
	}
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
		for _, s := range c.Body {
			if _, ok := s.(*EmptyStmt); !ok {
				ctx.WriteV(n+1, "\n", ctx.Indent, s.Format)
			}
		}
	}
	ctx.WriteV(n, "\n", ctx.Indent)
	if ctx.blockPositions() {
		ctx.WriteV(n, "/* #", s.End, " */}")
	} else {
		ctx.WriteString("}")
	}
}

func (s *SelectStmt) Format(ctx *FormatContext, n uint) {
	if ctx.stmtPositions() {
		ctx.WriteV(n, "/* #", s.Off, " */")
	}
	ctx.WriteString("select ")
	if len(s.Comms) == 0 {
		if ctx.blockPositions() {
			ctx.WriteV(n, "/* #", s.Begin, " */{/* #", s.End, " */}")
		} else {
			ctx.WriteString("{}")
		}
		return
	}
	if ctx.blockPositions() {
		ctx.WriteV(n, "/* #", s.Begin, " */{")
	} else {
		ctx.WriteString("{")
	}
	for _, c := range s.Comms {
		if c.Comm == nil {
			ctx.WriteV(n, "\n", ctx.Indent, "default:")
		} else {
			ctx.WriteV(n, "\n", ctx.Indent, "case ", c.Comm.Format, ":")
		}
		for _, s := range c.Body {
			if _, ok := s.(*EmptyStmt); !ok {
				ctx.WriteV(n+1, "\n", ctx.Indent, s.Format)
			}
		}
	}
	ctx.WriteV(n, "\n", ctx.Indent)
	if ctx.blockPositions() {
		ctx.WriteV(n, "/* #", s.End, " */}")
	} else {
		ctx.WriteString("}")
	}
}

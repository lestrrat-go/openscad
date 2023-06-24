package ast

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
)

func emitLetVars(ctx *EmitContext, w io.Writer, vars []*Variable) error {
	// If there are more than 3 variables, we want to put each variable on its own line.
	separateLine := len(vars) > 3
	ctx = ctx.WithAllowAssignment(true)
	for i, v := range vars {
		if i > 0 {
			fmt.Fprintf(w, `,`)
		}
		if separateLine {
			fmt.Fprint(w, "\n")
		} else {
			fmt.Fprintf(w, " ")
		}
		if err := emitExpr(ctx, w, v); err != nil {
			return err
		}
	}
	return nil
}

func emitLetPreamble(ctx *EmitContext, w io.Writer, vars []*Variable) error {
	fmt.Fprintf(w, `let(`)

	var letVars bytes.Buffer
	if err := emitLetVars(ctx, &letVars, vars); err != nil {
		return err
	}

	if strings.ContainsRune(letVars.String(), '\n') {
		if err := addIndent(w, &letVars, singleIndent); err != nil {
			return fmt.Errorf(`failed to write let variables: %v`, err)
		}
		fmt.Fprintf(w, "\n)")
	} else {
		if _, err := letVars.WriteTo(w); err != nil {
			return fmt.Errorf(`failed to write let variables: %v`, err)
		}
		fmt.Fprint(w, ")")
	}
	return nil
}

type LetExpr struct {
	variables []*Variable
	expr      interface{}
}

func NewLetExpr(variables ...*Variable) *LetExpr {
	return &LetExpr{
		variables: variables,
	}
}

func (l *LetExpr) Expr(expr interface{}) *LetExpr {
	l.expr = expr
	return l
}

func (l *LetExpr) EmitExpr(ctx *EmitContext, w io.Writer) error {
	var preamble bytes.Buffer
	if err := emitLetPreamble(ctx, &preamble, l.variables); err != nil {
		return err
	}

	var body bytes.Buffer
	if err := emitExpr(ctx, &body, l.expr); err != nil {
		return err
	}

	fmtAsBlock := strings.ContainsRune(preamble.String(), '\n') || strings.ContainsRune(body.String(), '\n')

	if _, err := preamble.WriteTo(w); err != nil {
		return fmt.Errorf(`failed to write let preamble: %v`, err)
	}

	if fmtAsBlock {
		fmt.Fprintf(w, "\n%s", singleIndent)
	}

	if _, err := body.WriteTo(w); err != nil {
		return fmt.Errorf(`failed to write let body: %v`, err)
	}

	return nil
}

type LetBlock struct {
	variables []*Variable
	children  []Stmt
}

func NewLetBlock(variables ...*Variable) *LetBlock {
	return &LetBlock{
		variables: variables,
	}
}

func (l *LetBlock) Body(children ...Stmt) *LetBlock {
	l.children = make([]Stmt, len(children))
	copy(l.children, children)
	return l
}

func (l *LetBlock) Add(stmt ...Stmt) *LetBlock {
	l.children = append(l.children, stmt...)
	return l
}

func (l *LetBlock) EmitStmt(ctx *EmitContext, w io.Writer) error {
	if err := emitLetPreamble(ctx, w, l.variables); err != nil {
		return err
	}
	if err := emitChildren(ctx, w, l.children, false); err != nil {
		return fmt.Errorf(`failed to emit let block: %v`, err)
	}
	return nil
}

type ForRange struct {
	start, end, increment interface{}
}

func NewForRange(start, end interface{}) *ForRange {
	if start == nil || end == nil {
		panic(`start and end must not be nil`)
	}
	return &ForRange{
		start: start,
		end:   end,
	}
}

func (fr *ForRange) Increment(incr interface{}) *ForRange {
	fr.increment = incr
	return fr
}

func (fr *ForRange) EmitExpr(ctx *EmitContext, w io.Writer) error {
	fmt.Fprint(w, `[`)
	if err := emitValue(ctx, w, fr.start); err != nil {
		return fmt.Errorf(`failed to emit for range start: %v`, err)
	}
	fmt.Fprint(w, `:`)
	if incr := fr.increment; incr != nil {
		if err := emitValue(ctx, w, incr); err != nil {
			return fmt.Errorf(`failed to emit for range increment: %v`, err)
		}
		fmt.Fprint(w, `:`)
	}
	if err := emitValue(ctx, w, fr.end); err != nil {
		return fmt.Errorf(`failed to emit for range end: %v`, err)
	}
	fmt.Fprint(w, `]`)
	return nil
}

type LoopVar struct {
	variable *Variable
	expr     interface{}
}

func NewLoopVar(variable *Variable, expr interface{}) *LoopVar {
	return &LoopVar{
		variable: variable,
		expr:     expr,
	}
}

func (lv *LoopVar) String() string {
	var sb strings.Builder
	if err := lv.EmitExpr(newEmitContext(), &sb); err != nil {
		return fmt.Sprintf(`#LoopVar(error=%s)`, err)
	}
	return sb.String()
}

func (lv *LoopVar) EmitExpr(ctx *EmitContext, w io.Writer) error {
	if err := emitValue(ctx, w, lv.variable); err != nil {
		return err
	}
	fmt.Fprint(w, `=`)
	if err := emitValue(ctx.WithAllowAssignment(false), w, lv.expr); err != nil {
		return err
	}
	return nil
}

type ForExpr struct {
	loopVars []*LoopVar
	expr     interface{}
}

func NewForExpr(loopVars []*LoopVar) *ForExpr {
	return &ForExpr{
		loopVars: loopVars,
	}
}

func (f *ForExpr) Body(expr interface{}) *ForExpr {
	f.expr = expr
	return f
}

func emitForDecl(ctx *EmitContext, w io.Writer, loopVars []*LoopVar) error {
	fmt.Fprint(w, "for (")
	ctx = ctx.WithAllowAssignment(false)
	for i, v := range loopVars {
		if i > 0 {
			fmt.Fprint(w, `, `)
		}
		if err := emitExpr(ctx, w, v); err != nil {
			return err
		}
	}
	fmt.Fprint(w, `)`)
	return nil
}

func addIndent(dst io.Writer, src io.Reader, indent string) error {
	scanner := bufio.NewScanner(src)
	i := 0
	for scanner.Scan() {
		if i > 0 {
			fmt.Fprintf(dst, "\n")
		}
		fmt.Fprintf(dst, "%s%s", indent, scanner.Text())
		i++
	}
	return scanner.Err()
}

func (f *ForExpr) EmitExpr(ctx *EmitContext, w io.Writer) error {
	// If the body expression contains any newlines, treat it as a block
	var body bytes.Buffer
	if err := emitExpr(ctx, &body, f.expr); err != nil {
		return err
	}
	if strings.ContainsRune(body.String(), '\n') {
		fmt.Fprintf(w, "\n")
		if err := emitForDecl(ctx, w, f.loopVars); err != nil {
			return fmt.Errorf(`failed to emit for expression: %w`, err)
		}

		fmt.Fprintln(w)
		if err := addIndent(w, &body, singleIndent); err != nil {
			return fmt.Errorf(`failed to emit for expression: %w`, err)
		}
	} else {
		if err := emitForDecl(ctx, w, f.loopVars); err != nil {
			return err
		}
		if err := emitExpr(ctx, w, f.expr); err != nil {
			return err
		}
	}

	return nil
}

// ForBlock represents a for loop block.
//
// In OpenSCAD for can take two distinctive styles: one as an expression, and
// one as a statement. ForBlock the expression, use ForExpr
type ForBlock struct {
	loopVars []*LoopVar
	children []Stmt
}

func NewFor(loopVars []*LoopVar) *ForBlock {
	return &ForBlock{
		loopVars: loopVars,
	}
}

func (f *ForBlock) Body(stmts ...Stmt) *ForBlock {
	f.children = make([]Stmt, len(stmts))
	copy(f.children, stmts)
	return f
}

func (f *ForBlock) Add(stmts ...Stmt) *ForBlock {
	f.children = append(f.children, stmts...)
	return f
}

func (f *ForBlock) EmitStmt(ctx *EmitContext, w io.Writer) error {
	indent := ctx.Indent()
	fmt.Fprintf(w, "\n%s", indent)
	if err := emitForDecl(ctx, w, f.loopVars); err != nil {
		return err
	}
	if err := emitChildren(ctx, w, f.children, true); err != nil {
		return fmt.Errorf(`failed to emit for block children: %w`, err)
	}
	return nil
}

type TernaryOp struct {
	condition interface{}
	trueExpr  interface{}
	falseExpr interface{}
}

func NewTernaryOp(condition, trueExpr, falseExpr interface{}) *TernaryOp {
	return &TernaryOp{
		condition: condition,
		trueExpr:  trueExpr,
		falseExpr: falseExpr,
	}
}

func (op *TernaryOp) Condition() interface{} {
	return op.condition
}

func (op *TernaryOp) TrueExpr() interface{} {
	return op.trueExpr
}

func (op *TernaryOp) FalseExpr() interface{} {
	return op.falseExpr
}

func (op *TernaryOp) EmitExpr(ctx *EmitContext, w io.Writer) error {
	ctx = ctx.WithAllowAssignment(false)

	var cond, trueExpr, falseExpr bytes.Buffer

	if err := emitExpr(ctx, &cond, op.condition); err != nil {
		return fmt.Errorf(`failed to emit ternary condition: %w`, err)
	}
	if err := emitExpr(ctx, &trueExpr, op.trueExpr); err != nil {
		return fmt.Errorf(`failed to emit ternary true expression: %w`, err)
	}
	if err := emitExpr(ctx, &falseExpr, op.falseExpr); err != nil {
		return fmt.Errorf(`failed to emit ternary false expression: %w`, err)
	}

	fmtAsBlock := strings.ContainsRune(cond.String(), '\n') ||
		strings.ContainsRune(trueExpr.String(), '\n') ||
		strings.ContainsRune(falseExpr.String(), '\n')
	if fmtAsBlock {
		if err := addIndent(w, &cond, ctx.Indent()); err != nil {
			return fmt.Errorf(`failed to write ternary condition: %w`, err)
		}
		fmt.Fprint(w, " ?\n")
		if err := addIndent(w, &trueExpr, ctx.Indent()+singleIndent); err != nil {
			return fmt.Errorf(`failed to write ternary true expression: %w`, err)
		}
		fmt.Fprint(w, " :\n")
		if err := addIndent(w, &falseExpr, ctx.Indent()+singleIndent); err != nil {
			return fmt.Errorf(`failed to write ternary false expression: %w`, err)
		}
	} else {
		if _, err := cond.WriteTo(w); err != nil {
			return fmt.Errorf(`failed to err ternary condition: %w`, err)
		}
		fmt.Fprint(w, ` ? `)
		if _, err := trueExpr.WriteTo(w); err != nil {
			return fmt.Errorf(`failed to err ternary true expression: %w`, err)
		}
		fmt.Fprint(w, ` : `)
		if _, err := falseExpr.WriteTo(w); err != nil {
			return fmt.Errorf(`failed to err ternary false expression: %w`, err)
		}
	}

	return nil
}

func (op *TernaryOp) EmitStmt(ctx *EmitContext, w io.Writer) error {
	fmt.Fprint(w, ctx.Indent())
	if err := op.EmitExpr(ctx, w); err != nil {
		return err
	}
	fmt.Fprint(w, `;`)
	return nil
}

func emitIfPreamble(ctx *EmitContext, w io.Writer, cond interface{}) error {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, "if (")
	if err := emitExpr(ctx, &buf, cond); err != nil {
		return fmt.Errorf(`failed to emit if condition: %w`, err)
	}
	if strings.ContainsRune(buf.String(), '\n') {
		if err := addIndent(w, &buf, ctx.Indent()+singleIndent); err != nil {
			return fmt.Errorf(`failed to write if condition: %w`, err)
		}
		fmt.Fprintf(w, "\n)")
	} else {
		if _, err := buf.WriteTo(w); err != nil {
			return fmt.Errorf(`failed to write if condition: %w`, err)
		}
		fmt.Fprintf(w, ")")
	}
	return nil
}

type IfExpr struct {
	cond interface{}
	body interface{}
}

func NewIfExpr(cond interface{}) *IfExpr {
	return &IfExpr{
		cond: cond,
	}
}

func (ib *IfExpr) Body(body interface{}) *IfExpr {
	ib.body = body
	return ib
}

func (ib *IfExpr) EmitExpr(ctx *EmitContext, w io.Writer) error {
	if err := emitIfPreamble(ctx, w, ib.cond); err != nil {
		return err
	}
	if err := emitExpr(ctx, w, ib.body); err != nil {
		return fmt.Errorf(`failed to emit if body: %w`, err)
	}
	return nil
}

type IfStmt struct {
	cond         interface{}
	body         []Stmt
	elseifBlocks []*ElseIfStmt
	elseBlock    []Stmt
}

func NewIfStmt(cond interface{}) *IfStmt {
	return &IfStmt{
		cond: cond,
	}
}

func (ib *IfStmt) Body(stmts ...Stmt) *IfStmt {
	ib.body = make([]Stmt, len(stmts))
	copy(ib.body, stmts)
	return ib
}

func (ib *IfStmt) AddElseIf(cond interface{}, stmts ...Stmt) *IfStmt {
	ib.elseifBlocks = append(ib.elseifBlocks, &ElseIfStmt{
		cond: cond,
		body: stmts,
	})
	return ib
}

func (ib *IfStmt) Else(stmts ...Stmt) *IfStmt {
	ib.elseBlock = make([]Stmt, len(stmts))
	copy(ib.elseBlock, stmts)
	return ib
}

func (ib *IfStmt) EmitStmt(ctx *EmitContext, w io.Writer) error {
	fmt.Fprintf(w, "\n%s", ctx.Indent())

	if err := emitIfPreamble(ctx, w, ib.cond); err != nil {
		return fmt.Errorf(`failed to emit if preamble: %w`, err)
	}

	if err := emitChildren(ctx, w, ib.body, false); err != nil {
		return fmt.Errorf(`failed to emit if true block: %w`, err)
	}

	for _, elseifBlock := range ib.elseifBlocks {
		if err := elseifBlock.EmitStmt(ctx, w); err != nil {
			return fmt.Errorf(`failed to emit elseif block: %w`, err)
		}
	}

	if ib.elseBlock != nil {
		fmt.Fprintf(w, "\n%selse", ctx.Indent())
		if err := emitChildren(ctx, w, ib.elseBlock, false); err != nil {
			return fmt.Errorf(`failed to emit if false block: %w`, err)
		}
	}

	return nil
}

type ElseIfStmt struct {
	cond interface{}
	body []Stmt
}

func (eb *ElseIfStmt) EmitStmt(ctx *EmitContext, w io.Writer) error {
	fmt.Fprintf(w, "\n%selse if (", ctx.Indent())
	if err := emitExpr(ctx, w, eb.cond); err != nil {
		return fmt.Errorf(`failed to emit else if condition: %w`, err)
	}
	fmt.Fprintf(w, ")")
	if err := emitChildren(ctx, w, eb.body, false); err != nil {
		return fmt.Errorf(`failed to emit else if block: %w`, err)
	}
	return nil
}

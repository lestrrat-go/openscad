package parser

import (
	"fmt"
	"log"
	"strconv"

	"github.com/lestrrat-go/openscad"
	"github.com/lestrrat-go/openscad/dsl"
)

type parser struct {
	ch      chan *Token
	peeked  []*Token
	readPos int
}

func Parse(src []byte) (openscad.Stmts, error) {
	ch := make(chan *Token, 1)

	go Lex(ch, src)

	p := &parser{
		ch:      ch,
		readPos: -1,
	}
	stmts, err := p.handleStatements()
	if err != nil {
		return nil, fmt.Errorf(`failed to parse: %w`, err)
	}
	return stmts, nil
}

func (p *parser) handleStatements() (openscad.Stmts, error) {
	var stmts openscad.Stmts
	for {
		log.Printf("new loop in handle any")
		tok := p.Peek()
		log.Printf("handleStatements: %#v", tok)
		switch tok.Type {
		case Keyword:
			switch tok.Value {
			case "let":
				p.Unread()
				letBlock, err := p.handleLetBlock()
				if err != nil {
					return nil, err
				}
				stmts = append(stmts, letBlock)
			case "module":
				p.Unread()
				module, err := p.handleModule()
				if err != nil {
					return nil, err
				}
				stmts = append(stmts, module)
			case "function":
				p.Unread()
				fn, err := p.handleFunction()
				if err != nil {
					return nil, err
				}
				tok = p.Next()
				if tok.Type != Semicolon {
					return nil, fmt.Errorf(`expected semicolon after function declaration for %q, got %q`, fn.Name(), tok.Value)
				}
				stmts = append(stmts, fn)
			case "for":
				p.Unread()
				block, err := p.handleForBlock()
				if err != nil {
					return nil, err
				}
				stmts = append(stmts, block)
			default:
				return nil, fmt.Errorf(`unknown keyword %q`, tok.Value)
			}
		case Ident:
			p.Unread()
			stmt, semicolon, err := p.handleAssignmentOrFunctionCall()
			if err != nil {
				return nil, err
			}

			if semicolon {
				tok = p.Next()
				if tok.Type != Semicolon {
					log.Printf("%#v", stmt)
					return nil, fmt.Errorf(`expected semicolon after assignment or function call, got %q`, tok.Value)
				}
			}
			stmts = append(stmts, stmt)
		case CloseBrace:
			p.Unread()
			return stmts, nil
		case EOF:
			return stmts, nil
		default:
			return nil, fmt.Errorf(`unhandled token %q`, tok.Value)
		}
	}
}

// Peek obtains the next token, but retains it in a buffer so that we can backtrack it.
// If we peek and then unread, we would be peeking the previously read, cached token
func (p *parser) Peek() *Token {
	// Only read more if we're at the end of the buffer
	if len(p.peeked)-1 == p.readPos {
		tok := <-p.ch
		if tok == nil {
			return nil
		}
		p.peeked = append(p.peeked, tok)
	}
	p.readPos++
	log.Printf("peek peeked=%d, readPos=%d", len(p.peeked), p.readPos)
	return p.peeked[p.readPos]
}

// Advance is akin to committing the previously peeked reads, effectively
// throwing away every buffered Token up to the current reading position
func (p *parser) Advance() {
	log.Printf("advance (BEFORE): peeked=%d, p.readPos=%d", len(p.peeked), p.readPos)
	if p.readPos > -1 {
		p.peeked = p.peeked[p.readPos:]
		if len(p.peeked) > 0 {
			p.readPos = 0
		} else {
			p.readPos = -1
		}
	}
	log.Printf("advance: peeked=%d, p.readPos=%d", len(p.peeked), p.readPos)
}

func (p *parser) Unread() {
	if p.readPos >= 0 {
		p.readPos--
	}
	log.Printf("unread -> %d", p.readPos)
}

func (p *parser) Next() *Token {
	tok := p.Peek()
	if tok != nil {
		p.Advance()
		return tok
	}
	panic("failed to peek")
	return nil
}

func (p *parser) handleModule() (*openscad.Module, error) {
	log.Printf("START module")
	defer log.Printf("END module")
	// module moduleName ( [... args ...]? ) { ... body ... }
	tok := p.Next()
	if tok.Type != Keyword || tok.Value != `module` {
		return nil, fmt.Errorf(`expected module`)
	}

	tok = p.Next()
	moduleName := tok.Value
	module := dsl.Module(moduleName)

	params, err := p.handleParameterList()
	if err != nil {
		return nil, fmt.Errorf(`failed to parse parameter list for module %q: %w`, moduleName, err)
	}

	stmts, err := p.handleBlock()
	if err != nil {
		return nil, fmt.Errorf(`failed to parse block for module %q: %w`, moduleName, err)
	}

	module.Parameters(params...)
	module.Body(stmts...)
	return module, nil
}

func (p *parser) handleParameterList() ([]*openscad.Variable, error) {
	log.Printf("START parameter list")
	defer log.Printf("END parameter list")
	tok := p.Next()
	if tok.Type != OpenParen {
		return nil, fmt.Errorf("expected open parenthesis")
	}

	var ret []*openscad.Variable

OUTER:
	for count := 1; ; count++ {
		tok = p.Peek()
		if tok.Type == CloseParen {
			p.Advance()
			break OUTER
		}
		p.Unread()

		v, err := p.handleParamDecl()
		if err != nil {
			return nil, fmt.Errorf(`failed to parse parameter %d: %w`, count, err)
		}
		ret = append(ret, v)

		// if we see a comma, then we expect more
		tok = p.Peek()
		if tok.Type == Comma {
			continue
		}
		p.Unread()
	}
	return ret, nil
}

func (p *parser) handleParamDecl() (*openscad.Variable, error) {
	log.Printf("START param decl")
	defer log.Printf("END param decl")
	tok := p.Next()
	if tok.Type != Ident {
		return nil, fmt.Errorf(`expected ident for param decl, got %q`, tok.Value)
	}

	name := tok.Value

	v := dsl.Variable(name)

	tok = p.Peek()
	if tok.Type != Equal {
		p.Unread()
	} else {
		expr, err := p.handleExpr()
		if err != nil {
			return nil, fmt.Errorf(`failed to parse expr for param decl: %w`, err)
		}
		v.Value(expr)
	}
	return v, nil
}

func (p *parser) handleBlock() (openscad.Stmts, error) {
	log.Printf("START block")
	defer log.Printf("END block")
	tok := p.Next()
	if tok.Type != OpenBrace {
		return nil, fmt.Errorf(`expected open brace, got %q`, tok.Value)
	}

	stmts, err := p.handleStatements()
	if err != nil {
		return nil, fmt.Errorf(`failed to parse block: %w`, err)
	}

	tok = p.Next()
	if tok.Type != CloseBrace {
		return nil, fmt.Errorf(`expected close brace, got %q`, tok.Value)
	}
	log.Printf("consumed close brace")
	return stmts, nil
}

func (p *parser) handleAssignment() (*openscad.Variable, error) {
	log.Printf("START assignment")
	defer log.Printf("END assignment")
	tok := p.Next()
	if tok.Type != Ident {
		return nil, fmt.Errorf(`expected identity of variable to assign to, got %q`, tok.Value)
	}
	varName := tok.Value
	v := dsl.Variable(varName)
	log.Printf("handleAssignment %#v", tok)

	tok = p.Next()
	if tok.Type != Equal {
		return nil, fmt.Errorf(`expected '=', got %q`, tok.Value)
	}

	expr, err := p.handleExpr()
	if err != nil {
		return nil, fmt.Errorf(`failed to parse expression: %w`, err)
	}
	v.Value(expr)

	log.Printf("variable %q value %#v", varName, expr)
	return v, nil
}

func (p *parser) handleCall() (*openscad.Call, bool, error) {
	log.Printf("START call")
	defer log.Printf("END call")
	tok := p.Next()
	if tok.Type != Ident {
		return nil, false, fmt.Errorf(`expected function name, got %q`, tok.Value)
	}

	callName := tok.Value
	call := dsl.Call(callName)

	tok = p.Next()
	if tok.Type != OpenParen {
		return nil, false, fmt.Errorf(`expected open paren, got %q for function call on %q`, tok.Value, callName)
	}

	var parameters []interface{}
OUTER:
	for {
		tok = p.Peek()
		if tok.Type == CloseParen {
			p.Advance()
			break OUTER
		}
		p.Unread()

		expr, err := p.handleExpr()
		if err != nil {
			return nil, false, fmt.Errorf(`failed to parse expression in parameter list for function call %q: %w`, callName, err)
		}
		parameters = append(parameters, expr)

		tok = p.Peek()
		switch tok.Type {
		case Comma:
			continue
		default:
			p.Unread()
		}
	}

	log.Printf("call parameters %#v", parameters)
	call.Parameters(parameters...)

	// If there is either a block or another function call, that's a child statement
	var semicolon bool
	tok = p.Peek()
	switch tok.Type {
	case OpenBrace:
		p.Unread()
		stmts, err := p.handleBlock()
		if err != nil {
			return nil, false, fmt.Errorf(`failed to parse block: %w`, err)
		}
		call.Add(stmts...)
		semicolon = false
	case Ident:
		// This should be a function call
		p.Unread()
		child, childsemicolon, err := p.handleCall()
		if err != nil {
			return nil, false, fmt.Errorf(`failed to parse child function call: %w`, err)
		}
		call.Add(child)
		semicolon = childsemicolon
	default:
		log.Printf("No children")
		semicolon = true
		p.Unread()
	}

	log.Printf("call %q semicolon %t", callName, semicolon)
	return call, semicolon, nil
}

func (p *parser) handleParenExpr() (ret interface{}, reterr error) {
	log.Printf("START parenexpr")
	defer func(reterr *error) {
		log.Printf("END parenexpr: %#v", *reterr)
	}(&reterr)
	tok := p.Next()
	if tok.Type != OpenParen {
		return nil, fmt.Errorf(`expected open paren, got %q`, tok.Value)
	}

	expr, err := p.handleExpr()
	if err != nil {
		return nil, fmt.Errorf(`failed to parse expression: %w`, err)
	}

	tok = p.Next()
	if tok.Type != CloseParen {
		return nil, fmt.Errorf(`expected close paren, got %#v`, tok)
	}
	return dsl.Group(expr), nil
}

func (p *parser) handleExpr() (ret interface{}, reterr error) {
	log.Printf("START expr")
	defer func(ret *interface{}) {
		log.Printf("END expr %#v", *ret)
	}(&ret)

	var expr interface{}

	tok := p.Next()
	switch tok.Type {
	case Keyword:
		p.Unread()
		switch tok.Value {
		case "let":
			ve, err := p.handleLetExpr()
			if err != nil {
				return nil, fmt.Errorf(`failed to parse let expression: %w`, err)
			}
			expr = ve
		case "for":
			fe, err := p.handleForExpr()
			if err != nil {
				return nil, fmt.Errorf(`failed to parse for expression: %w`, err)
			}
			expr = fe
		default:
			return nil, fmt.Errorf(`unexpected keyword %q`, tok.Value)
		}
	case OpenParen:
		p.Unread()
		pe, err := p.handleParenExpr()
		if err != nil {
			return nil, fmt.Errorf(`failed to parse parenthesized expression: %w`, err)
		}
		expr = pe
	case Minus:
		// This is a unary minus
		operand, err := p.handleExpr()
		if err != nil {
			return nil, fmt.Errorf(`failed to parse expression: %w`, err)
		}
		expr = dsl.Negative(operand)
	case Literal:
		expr = tok.Value
	case Numeric:
		f, err := strconv.ParseFloat(tok.Value, 64)
		if err != nil {
			return nil, fmt.Errorf(`failed to parse numeric literal %q: %w`, tok.Value, err)
		}
		expr = f
	case Ident:
		p.Unread()
		// could be a function call, or just a variable
		stmt, _, err := p.handleAssignmentOrFunctionCall()
		if err == nil {
			expr = stmt
		} else {
			expr = dsl.Variable(tok.Value)
		}
	case OpenBracket:
		// This is a list or a loop range
		p.Unread()
		list, err := p.handleList()
		if err != nil {
			return nil, fmt.Errorf(`failed to parse list: %w`, err)
		}
		expr = list
	default:
		return nil, fmt.Errorf("unhandled expr %#v", tok)
	}

	if op, err := p.tryOperator(expr); op != nil && err == nil {
		expr = op
	}

	if expr == nil {
		panic(`nil expr!`)
	}
	return expr, nil
}

func (p *parser) handleTernary(cond interface{}) (interface{}, error) {
	log.Printf("START ternary")
	defer log.Printf("END ternary")
	tok := p.Next()
	if tok.Type != Question {
		return nil, fmt.Errorf(`expected question mark, got %q`, tok.Value)
	}

	trueExpr, err := p.handleExpr()
	if err != nil {
		return nil, fmt.Errorf(`failed to parse true expression: %w`, err)
	}

	tok = p.Next()
	if tok.Type != Colon {
		return nil, fmt.Errorf(`expected colon, got %q`, tok.Value)
	}

	falseExpr, err := p.handleExpr()
	if err != nil {
		return nil, fmt.Errorf(`failed to parse false expression: %w`, err)
	}

	return dsl.Ternary(cond, trueExpr, falseExpr), nil
}

func (p *parser) handleAssignmentOrFunctionCall() (openscad.Stmt, bool, error) {
	tok := p.Peek()
	if tok.Type != Ident {
		return nil, false, fmt.Errorf(`expected ident, got %q`, tok.Value)
	}

	tok = p.Peek()
	switch tok.Type {
	case Equal:
		p.Unread()
		p.Unread()
		variable, err := p.handleAssignment()
		if err != nil {
			return nil, true, fmt.Errorf(`failed to parse assignment: %w`, err)
		}
		return variable, true, nil
	case OpenParen:
		p.Unread()
		p.Unread()
		call, semicolon, err := p.handleCall()
		if err != nil {
			return nil, false, fmt.Errorf(`failed to parse function call: %w`, err)
		}
		return call, semicolon, nil
	default:
		p.Unread()
		return nil, false, fmt.Errorf(`expected assignment or function call, got %q after ident`, tok.Value)
	}
}

func (p *parser) handleList() ([]interface{}, error) {
	log.Printf("START list")
	defer log.Printf("END list")
	tok := p.Next()
	if tok.Type != OpenBracket {
		return nil, fmt.Errorf(`expected open bracket, got %q`, tok.Value)
	}

	var list []interface{}
	for {
		tok = p.Peek()
		if tok.Type == CloseBracket {
			p.Advance()
			break
		}
		p.Unread()

		expr, err := p.handleExpr()
		if err != nil {
			return nil, fmt.Errorf(`failed to parse expression: %w`, err)
		}
		list = append(list, expr)

		tok = p.Peek()
		switch tok.Type {
		case Comma:
			continue
		default:
			p.Unread()
		}
	}

	return list, nil
}

func (p *parser) handleFunction() (*openscad.Function, error) {
	tok := p.Next()
	if tok.Type != Keyword && tok.Value != "function" {
		return nil, fmt.Errorf(`expected function, got %q`, tok.Value)
	}

	tok = p.Next()
	if tok.Type != Ident {
		return nil, fmt.Errorf(`expected function name, got %q`, tok.Value)
	}

	name := tok.Value
	fn := dsl.Function(name)

	tok = p.Next()
	if tok.Type != OpenParen {
		return nil, fmt.Errorf(`expected open paren, got %q`, tok.Value)
	}

	var parameters []interface{}
	for {
		tok = p.Peek()
		if tok.Type == CloseParen {
			p.Advance()
			break
		}
		p.Unread()

		expr, err := p.handleExpr()
		if err != nil {
			return nil, fmt.Errorf(`failed to parse function parameter: %w`, err)
		}
		parameters = append(parameters, expr)

		tok = p.Peek()
		switch tok.Type {
		case Comma:
			continue
		default:
			p.Unread()
		}
	}

	// TODO for functions, parameters need to be *openscad.Variable
	vparams := make([]*openscad.Variable, len(parameters))
	for i, p := range parameters {
		switch v := p.(type) {
		case *openscad.Variable:
			vparams[i] = v
		default:
			return nil, fmt.Errorf(`expected variable in function parameter, got %T`, p)
		}
	}

	fn.Parameters(vparams...)

	tok = p.Next()
	if tok.Type != Equal {
		return nil, fmt.Errorf(`expected equal, got %q`, tok.Value)
	}

	expr, err := p.handleExpr()
	if err != nil {
		return nil, fmt.Errorf(`failed to parse function expression: %w`, err)
	}

	fn.Body(expr)
	return fn, nil
}

func (p *parser) mungeOperatorPrecedence(expr interface{}) interface{} {
	bop, ok := expr.(*openscad.BinaryOp)
	if !ok {
		return expr
	}

	switch rop := bop.Right().(type) {
	case *openscad.BinaryOp:
		return bop.Rearrange(rop)
	case *openscad.TernaryOp:
		// Take the condition of the ternary op, and make it the right hand side
		return openscad.NewTernaryOp(
			openscad.NewBinaryOp(bop.Op(), bop.Left(), rop.Condition()),
			rop.TrueExpr(),
			rop.FalseExpr(),
		)
	default:
		return expr
	}

}

func (p *parser) tryOperator(left interface{}) (interface{}, error) {
	tok := p.Peek()
	var ret interface{}
	switch tok.Type {
	case Question:
		p.Unread()
		ternary, err := p.handleTernary(left)
		if err != nil {
			return nil, fmt.Errorf(`failed to parse ternary expression: %w`, err)
		}
		ret = ternary
	case Equality:
		expr, err := p.handleExpr()
		if err != nil {
			return nil, fmt.Errorf(`failed to parse right hand expression of '==': %w`, err)
		}
		ret = dsl.EQ(left, expr)
	case LessThan:
		expr, err := p.handleExpr()
		if err != nil {
			return nil, fmt.Errorf(`failed to parse right hand expression of '<': %w`, err)
		}
		ret = dsl.LT(left, expr)
	case LessThanEqual:
		expr, err := p.handleExpr()
		if err != nil {
			return nil, fmt.Errorf(`failed to parse right hand expression of '<=': %w`, err)
		}
		ret = dsl.LE(left, expr)
	case GreaterThan:
		expr, err := p.handleExpr()
		if err != nil {
			return nil, fmt.Errorf(`failed to parse right hand expression of '>': %w`, err)
		}
		ret = dsl.GT(left, expr)
	case GreaterThanEqual:
		expr, err := p.handleExpr()
		if err != nil {
			return nil, fmt.Errorf(`failed to parse right hand expression of '>=': %w`, err)
		}
		ret = dsl.GE(left, expr)
	case Asterisk:
		expr, err := p.handleExpr()
		if err != nil {
			return nil, fmt.Errorf(`failed to parse right hand expression of '*': %w`, err)
		}

		ret = dsl.Mul(left, expr)
	case Slash:
		expr, err := p.handleExpr()
		if err != nil {
			return nil, fmt.Errorf(`failed to parse right hand expression of '/': %w`, err)
		}
		ret = dsl.Div(left, expr)
	case Plus:
		expr, err := p.handleExpr()
		if err != nil {
			return nil, fmt.Errorf(`failed to parse right hand expression of '+': %w`, err)
		}
		ret = dsl.Add(left, expr)
	case Minus:
		expr, err := p.handleExpr()
		if err != nil {
			return nil, fmt.Errorf(`failed to parse right hand expression of '-': %w`, err)
		}
		ret = dsl.Sub(left, expr)
	case Percent:
		expr, err := p.handleExpr()
		if err != nil {
			return nil, fmt.Errorf(`failed to parse right hand expression of '%%': %w`, err)
		}
		ret = dsl.Mod(left, expr)
	case OpenBracket:
		expr, err := p.handleExpr()
		if err != nil {
			return nil, fmt.Errorf(`failed to parse index expression of '[]': %w`, err)
		}

		tok = p.Peek()
		if tok.Type != CloseBracket {
			return nil, fmt.Errorf(`expected close bracket, got %q`, tok.Value)
		}
		p.Advance()

		// Only in this instance, we need to further, because we may
		// have a list[expr] followed by another operator
		expr2, err := p.tryOperator(dsl.Index(left, expr))
		if err != nil {
			return nil, err
		}
		ret = expr2
	default:
		log.Printf("not an operator %#v", tok)
		p.Unread()
		return left, nil
	}

	return p.mungeOperatorPrecedence(ret), nil
}

func (p *parser) handleForExpr() (*openscad.ForExpr, error) {
	loopVars, err := p.handleForPreamble()
	if err != nil {
		return nil, fmt.Errorf(`failed to parse for loop preamble: %w`, err)
	}

	forExpr := dsl.ForExpr(loopVars...)
	expr, err := p.handleExpr()
	if err != nil {
		return nil, fmt.Errorf(`failed to parse for expression: %w`, err)
	}
	forExpr.Body(expr)
	return forExpr, nil
}

func (p *parser) handleForBlock() (*openscad.ForBlock, error) {
	loopVars, err := p.handleForPreamble()
	if err != nil {
		return nil, fmt.Errorf(`failed to parse for loop preamble: %w`, err)
	}

	forStmt := dsl.For(loopVars...)
	stmts, err := p.handleBlock()
	if err != nil {
		return nil, fmt.Errorf(`failed to parse for block: %w`, err)
	}
	forStmt.Body(stmts...)

	return forStmt, nil
}

func (p *parser) handleForPreamble() ([]*openscad.LoopVar, error) {
	tok := p.Next()
	if tok.Type != Keyword || tok.Value != "for" {
		return nil, fmt.Errorf(`expected for, got %q`, tok.Value)
	}

	tok = p.Next()
	if tok.Type != OpenParen {
		return nil, fmt.Errorf(`expected open paren, got %q`, tok.Value)
	}

	// Multiple for variables can be specified, such as
	// for (i=[0:1], j=foobar(), z=[1, 2, 3])
	var loopVars []*openscad.LoopVar
	for {
		tok = p.Peek()
		if tok.Type == CloseParen {
			break
		}
		p.Unread()

		variable, err := p.handleForLoopVariable()
		if err != nil {
			return nil, fmt.Errorf(`failed to parse for loop variable: %w`, err)
		}

		loopVars = append(loopVars, variable)

		tok = p.Peek()
		if tok.Type == Comma {
			continue
		}
		p.Unread()
	}
	return loopVars, nil
}

func (p *parser) handleForRange() (*openscad.ForRange, error) {
	log.Printf("START handleForRange")
	defer log.Printf("END handleForRange")
	tok := p.Next()
	if tok.Type != OpenBracket {
		return nil, fmt.Errorf(`expected open bracket, got %q`, tok.Value)
	}

	initExpr, err := p.handleExpr()
	if err != nil {
		return nil, fmt.Errorf(`failed to parse first element for range expression: %w`, err)
	}

	tok = p.Next()
	if tok.Type != Colon {
		return nil, fmt.Errorf(`expected colon, got %q`, tok.Value)
	}

	endExpr, err := p.handleExpr()
	if err != nil {
		return nil, fmt.Errorf(`failed to parse second element for range expression: %w`, err)
	}

	var stepExpr interface{}
	tok = p.Peek()
	if tok.Type != Colon {
		p.Unread()
	} else {
		// we have a three element range expression
		endExpr = stepExpr
		stepExpr, err = p.handleExpr()
		if err != nil {
			return nil, fmt.Errorf(`failed to parse 'end' element for range expression: %w`, err)
		}
	}

	tok = p.Next()
	if tok.Type != CloseBracket {
		return nil, fmt.Errorf(`expected close bracket, got %q`, tok.Value)
	}

	fr := dsl.ForRange(initExpr, endExpr)
	if stepExpr != nil {
		fr.Increment(stepExpr)
	}
	return fr, nil
}

func (p *parser) handleForLoopVariable() (*openscad.LoopVar, error) {
	tok := p.Next()
	if tok.Type != Ident {
		return nil, fmt.Errorf(`expected loop variable identifier, got %q`, tok.Value)
	}

	variable := dsl.Variable(tok.Value)

	tok = p.Next()
	if tok.Type != Equal {
		return nil, fmt.Errorf(`expected equals, got %q`, tok.Value)
	}

	// First, try a for range expression, if that fails, try expr
	var frexpr interface{}
	fr, err := p.handleForRange()
	if err == nil {
		frexpr = fr
	} else {
		log.Printf("failed for range: %s", err)
		expr, err := p.handleExpr()
		if err != nil {
			return nil, fmt.Errorf(`failed to parse for loop variable expression: %w`, err)
		}
		frexpr = expr
	}
	return dsl.LoopVar(variable, frexpr), nil
}

func (p *parser) handleLetExpr() (*openscad.LetExpr, error) {
	vars, err := p.handleLetPreamble()
	if err != nil {
		return nil, fmt.Errorf(`failed to parse let preamble: %w`, err)
	}
	letExpr := dsl.LetExpr(vars...)
	expr, err := p.handleExpr()
	if err != nil {
		return nil, fmt.Errorf(`failed to parse let expression: %w`, err)
	}
	letExpr.Expr(expr)
	return letExpr, nil
}

func (p *parser) handleLetBlock() (*openscad.LetBlock, error) {
	vars, err := p.handleLetPreamble()
	if err != nil {
		return nil, fmt.Errorf(`failed to parse let preamble: %w`, err)
	}
	letBlock := dsl.LetBlock(vars...)
	stmts, err := p.handleBlock()
	if err != nil {
		return nil, fmt.Errorf(`failed to parse let block: %w`, err)
	}
	letBlock.Add(stmts...)
	return letBlock, nil
}

func (p *parser) handleLetPreamble() ([]*openscad.Variable, error) {
	tok := p.Next()
	if tok.Type != Keyword || tok.Value != "let" {
		return nil, fmt.Errorf(`expected let, got %q`, tok.Value)
	}

	tok = p.Next()
	if tok.Type != OpenParen {
		return nil, fmt.Errorf(`expected open paren, got %q`, tok.Value)
	}

	// Multiple variables can be declared
	var letVars []*openscad.Variable
	for {
		tok = p.Peek()
		if tok.Type == CloseParen {
			break
		}
		p.Unread()

		variable, err := p.handleAssignment()
		if err != nil {
			return nil, fmt.Errorf(`failed to parse let variable: %w`, err)
		}

		letVars = append(letVars, variable)

		tok = p.Peek()
		if tok.Type == Comma {
			continue
		}
		p.Unread()
	}
	return letVars, nil
}

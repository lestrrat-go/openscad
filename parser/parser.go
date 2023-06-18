package parser

import (
	"fmt"
	"log"
	"strconv"

	"github.com/lestrrat-go/openscad"
	"github.com/lestrrat-go/openscad/dsl"
)

type parser struct {
	ch     chan *Token
	parent interface{}

	peeked  []*Token
	readPos int
}

func (p *parser) addChild(child interface{}) {
	log.Printf("addChild: %#v", child)
	switch parent := p.parent.(type) {
	case openscad.Stmts:
		parent.Add(child.(openscad.Stmt))
		p.parent = parent
	case *openscad.Module:
		log.Printf("adding child to module")
		parent.Add(child.(openscad.Stmt))
	case *openscad.Call:
		parent.Add(child.(openscad.Stmt))
	default:
		panic(fmt.Errorf(`unknown parent type: %T`, parent))
	}
}

func Parse(src []byte) (openscad.Stmts, error) {
	ch := make(chan *Token, 1)

	go Lex(ch, src)

	stmts := dsl.Stmts([]openscad.Stmt{}...)
	log.Printf("stmts -> %p", stmts)
	p := &parser{
		ch:      ch,
		parent:  stmts,
		readPos: -1,
	}
	if err := p.handleStatements(); err != nil {
		return nil, fmt.Errorf(`failed to parse: %w`, err)
	}
	log.Printf("p.parent -> %p", p.parent)
	return p.parent.(openscad.Stmts), nil
}

func (p *parser) handleStatements() error {
	for {
		log.Printf("new loop in handle any")
		tok := p.Peek()
		log.Printf("handleStatements: %#v", tok)
		switch tok.Type {
		case Keyword:
			switch tok.Value {
			case "module":
				p.Unread()
				module, err := p.handleModule()
				if err != nil {
					return err
				}
				p.addChild(module)
				log.Printf("returned from handleModule")
			case "function":
				p.Unread()
				fn, err := p.handleFunction()
				if err != nil {
					return err
				}
				tok = p.Next()
				if tok.Type != Semicolon {
					return fmt.Errorf(`expected semicolon after function declaration, got %q`, tok.Value)
				}
				p.addChild(fn)
			default:
				return fmt.Errorf(`unknown keyword %q`, tok.Value)
			}
		case Ident:
			p.Unread()
			stmt, semicolon, err := p.handleAssignmentOrFunctionCall()
			if err != nil {
				return err
			}

			if semicolon {
				tok = p.Next()
				if tok.Type != Semicolon {
					return fmt.Errorf(`expected semicolon after assignment or function call, got %q`, tok.Value)
				}
			}
			p.addChild(stmt)
		case CloseBrace:
			p.Unread()
			return nil
		case EOF:
			return nil
		default:
			p.Unread()
			stmt, err := p.handleExpr()
			if err != nil {
				return fmt.Errorf(`failed to parse expr in statements: %w:`, err)
			}
			tok = p.Next()
			if tok.Type != Semicolon {
				return fmt.Errorf(`expected semicolon after statement, got %q`, tok.Value)
			}
			p.addChild(stmt)
		}
	}
	return nil
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

	parent := p.parent
	p.parent = module
	defer func() { p.parent = parent }()

	params, err := p.handleParameterList()
	if err != nil {
		return nil, fmt.Errorf(`failed to parse parameter list for module %q: %w`, moduleName, err)
	}

	if err := p.handleBlock(); err != nil {
		return nil, fmt.Errorf(`failed to parse block for module %q: %w`, moduleName, err)
	}

	module.Parameters(params...)
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

	// "best" case, there are no arguments
	tok = p.Peek()
	if tok.Type == CloseParen {
		p.Advance()
		return nil, nil
	}
	p.Unread()

	// There are arguments!
	for {
		// ident or literal
		tok = p.Next()
		switch tok.Type {
		case Ident:
			ident := tok.Value
			v := dsl.Variable(ident)
			ret = append(ret, v)
			// There could be a default value

			log.Printf("param: %q", ident)
			tok = p.Peek()
			switch tok.Type {
			case Equal:
				p.Advance()
				tok = p.Next()
				if tok.Type == Literal {
					v.Value(tok.Value)
				}
			default:
				p.Unread()
			}
		}

		// if we see a comma, then we expect more
		tok = p.Peek()
		if tok.Type == Comma {
			p.Advance()
			continue
		}
		p.Unread()

		// otherwise, it better be a close paren
		tok = p.Next()
		if tok.Type != CloseParen {
			return nil, fmt.Errorf(`expected close paren, got %q`, tok.Value)
		}
		break
	}
	return ret, nil
}

func (p *parser) handleBlock() error {
	log.Printf("START block")
	defer log.Printf("END block")
	tok := p.Next()
	if tok.Type != OpenBrace {
		return fmt.Errorf(`expected open brace, got %q`, tok.Value)
	}

	if err := p.handleStatements(); err != nil {
		return fmt.Errorf(`failed to parse block: %w`, err)
	}

	tok = p.Next()
	if tok.Type != CloseBrace {
		return fmt.Errorf(`expected close brace, got %q`, tok.Value)
	}
	log.Printf("consumed close brace")
	return nil
}

func (p *parser) handleAssignment() (openscad.Stmt, error) {
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
		parent := p.parent
		p.parent = call
		err := p.handleBlock()
		p.parent = parent
		if err != nil {
			return nil, false, fmt.Errorf(`failed to parse block: %w`, err)
		}
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

func (p *parser) handleParenExpr() (interface{}, error) {
	log.Printf("START parenexpr")
	defer log.Printf("END parenexpr")
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
		return nil, fmt.Errorf(`expected close paren, got %q`, tok.Value)
	}
	return dsl.Group(expr), nil
}

func (p *parser) handleExpr() (interface{}, error) {
	log.Printf("START expr")
	defer log.Printf("END expr")

	var expr interface{}

	tok := p.Next()
	switch tok.Type {
	case OpenParen:
		p.Unread()
		pe, err := p.handleParenExpr()
		if err != nil {
			return nil, fmt.Errorf(`failed to parse parenthesized expression: %w`, err)
		}
		expr = pe
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
		// This is a list
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

func (p *parser) mungeNumericOperatorPrecedence(expr interface{}) interface{} {
	bop, ok := expr.(*openscad.BinaryOp)
	if !ok {
		return expr
	}

	var parentop func(interface{}, interface{}) *openscad.BinaryOp
	switch bop.Op() {
	case "*":
		parentop = dsl.Mul
	case "/":
		parentop = dsl.Div
	default:
		return expr
	}

	rop, ok := bop.Right().(*openscad.BinaryOp)
	if !ok {
		return expr
	}

	switch rop.Op() {
	case "+":
		return dsl.Add(parentop(bop.Left(), rop.Left()), rop.Right())
	case "-":
		return dsl.Sub(parentop(bop.Left(), rop.Left()), rop.Right())
	}
	return expr
}

func (p *parser) tryOperator(left interface{}) (interface{}, error) {
	log.Printf("START tryOperator")
	defer log.Printf("END tryOperator")
	tok := p.Peek()
	log.Printf("%#v", tok)
	switch tok.Type {
	case Question:
		p.Unread()
		return p.handleTernary(left)
	case Equality:
		expr, err := p.handleExpr()
		if err != nil {
			return nil, fmt.Errorf(`failed to parse right hand expression of '==': %w`, err)
		}
		return dsl.EQ(left, expr), nil
	case Asterisk:
		expr, err := p.handleExpr()
		if err != nil {
			return nil, fmt.Errorf(`failed to parse right hand expression of '*': %w`, err)
		}

		return p.mungeNumericOperatorPrecedence(dsl.Mul(left, expr)), nil
	case Slash:
		expr, err := p.handleExpr()
		if err != nil {
			return nil, fmt.Errorf(`failed to parse right hand expression of '/': %w`, err)
		}
		return p.mungeNumericOperatorPrecedence(dsl.Div(left, expr)), nil
	case Plus:
		expr, err := p.handleExpr()
		if err != nil {
			return nil, fmt.Errorf(`failed to parse right hand expression of '+': %w`, err)
		}
		return dsl.Add(left, expr), nil
	case Minus:
		expr, err := p.handleExpr()
		if err != nil {
			return nil, fmt.Errorf(`failed to parse right hand expression of '-': %w`, err)
		}
		return dsl.Sub(left, expr), nil
	default:
		log.Printf("not an operator %#v", tok)
		p.Unread()
		return nil, nil
	}
}

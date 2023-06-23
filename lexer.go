package openscad

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

var moduleKeyword = []byte(`module`)
var functionKeyword = []byte(`function`)
var forKeyword = []byte(`for`)
var letKeyword = []byte(`let`)
var includeKeyword = []byte(`include`)
var useKeyword = []byte(`use`)

const (
	plus         = '+'
	minus        = '-'
	asterisk     = '*'
	slash        = '/'
	equal        = '='
	dquote       = '"'
	colon        = ':'
	semicolon    = ';'
	comma        = ','
	openParen    = '('
	closeParen   = ')'
	openBrace    = '{'
	closeBrace   = '}'
	openBracket  = '['
	closeBracket = ']'
	lessThan     = '<'
	greaterThan  = '>'
	question     = '?'
	percent      = '%'
)
const (
	EOF = iota
	Keyword
	Literal
	Numeric
	Ident
	Comma
	Equal
	Semicolon
	Colon
	Question
	DoubleQuote
	Asterisk
	Plus
	Minus
	Slash
	LessThan
	LessThanEqual
	GreaterThan
	GreaterThanEqual
	Equality     // ==
	OpenParen    // (
	CloseParen   // )
	OpenBracket  // [
	CloseBracket // ]
	OpenBrace    // {
	CloseBrace   // }
	Percent
)

type Token struct {
	Type  int
	Value string
}

type lexer struct {
	src     []byte
	ch      chan *Token
	pos     int
	peekPos []int
}

func (l *lexer) skipWhiteSpaces() {
	for len(l.src) > 0 {
		r := l.peek()
		if !unicode.IsSpace(r) {
			l.unread()
			break
		}
		l.advance()
	}
}

func (l *lexer) consume(v []byte) {
	l.src = bytes.TrimPrefix(l.src, v)
}

func (l *lexer) emit(typ int, value string) {
	l.ch <- &Token{Type: typ, Value: value}
}

type cache struct {
	r rune
	s int
}

func (l *lexer) peek() rune {
	r, s := utf8.DecodeRune(l.src[l.pos:])
	l.pos += s
	l.peekPos = append(l.peekPos, s)
	return r
}

func (l *lexer) unread() {
	if lp := len(l.peekPos); lp > 0 {
		l.pos -= l.peekPos[lp-1]
		l.peekPos = l.peekPos[:lp-1]
	}
}

func (l *lexer) advance() {
	l.peekPos = nil
	l.src = l.src[l.pos:]
	l.pos = 0
}

func (l *lexer) emitBuffer(typ int) {
	l.emit(typ, string(l.src[:l.pos]))
	l.advance()
}

func Lex(ch chan *Token, src []byte) {
	l := lexer{
		src: src,
		ch:  ch,
	}
	defer close(ch)

	var inInclude bool
	for len(l.src) > 0 {
		l.skipWhiteSpaces()

		found := true

		peeked := l.peek()
		switch peeked {
		case comma:
			l.emitBuffer(Comma)
		case equal:
			next := l.peek()
			if next == equal {
				l.emitBuffer(Equality)
			} else {
				l.unread()
				l.emitBuffer(Equal)
			}
		case semicolon:
			l.emitBuffer(Semicolon)
		case colon:
			l.emitBuffer(Colon)
		case openBracket:
			l.emitBuffer(OpenBracket)
		case closeBracket:
			l.emitBuffer(CloseBracket)
		case openParen:
			l.emitBuffer(OpenParen)
		case closeParen:
			l.emitBuffer(CloseParen)
		case openBrace:
			l.emitBuffer(OpenBrace)
		case closeBrace:
			l.emitBuffer(CloseBrace)
		case question:
			l.emitBuffer(Question)
		case asterisk:
			l.emitBuffer(Asterisk)
		case plus:
			l.emitBuffer(Plus)
		case minus:
			l.emitBuffer(Minus)
		case slash:
			next := l.peek()
			switch next {
			case slash:
				// inline comment, ignore until end of line
				for {
					r := l.peek()
					if r == '\n' || r == utf8.RuneError {
						break
					}
				}
				l.advance()
			case asterisk:
				// block comment, ignore until */
			OUTER:
				for {
					r := l.peek()
					switch r {
					case utf8.RuneError:
						break OUTER
					case asterisk:
						next := l.peek()
						if next == slash {
							break OUTER
						}
					}
				}
				l.advance()
			default:
				l.unread()
				l.emitBuffer(Slash)
			}
		case lessThan:
			if inInclude {
				l.unread()
				if err := l.captureLiteral(lessThan, greaterThan); err != nil {
					return
				}
				inInclude = false
				continue
			}

			next := l.peek()
			if next != equal {
				l.unread()
				l.emitBuffer(LessThan)
			} else {
				l.advance()
				l.emitBuffer(LessThanEqual)
			}
		case greaterThan:
			next := l.peek()
			if next != equal {
				l.unread()
				l.emitBuffer(GreaterThan)
			} else {
				l.advance()
				l.emitBuffer(GreaterThanEqual)
			}
		case percent:
			l.emitBuffer(Percent)
		default:
			found = false
		}
		if found {
			continue
		}

		// maybe it's one of the keywords...
		found = true
		switch peeked {
		case 'i':
			l.unread()
			if err := l.expect(Keyword, includeKeyword); err == nil {
				inInclude = true
			} else {
				found = false
			}
		case 'u':
			l.unread()
			if err := l.expect(Keyword, useKeyword); err == nil {
				inInclude = true
			} else {
				found = false
			}
		case 'l':
			l.unread()
			if err := l.expect(Keyword, letKeyword); err != nil {
				found = false
			}
		case 'm':
			l.unread()
			if err := l.expect(Keyword, moduleKeyword); err != nil {
				found = false
			}
		case 'f':
			l.unread()
			if err := l.expect(Keyword, forKeyword); err != nil {
				if err := l.expect(Keyword, functionKeyword); err != nil {
					found = false
				}
			}
		case dquote:
			l.unread()
			if err := l.captureLiteral(dquote, dquote); err != nil {
				found = false
			}
		case '-':
			l.unread()
			if err := l.captureNumeric(); err != nil {
				found = false
			}
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			l.unread()
			if err := l.captureNumeric(); err != nil {
				found = false
			}
		default:
			found = false
		}
		if found {
			continue
		}
		l.unread()

		// it must be an identifier, then
		l.captureIdent()
	}
	l.emit(EOF, "")
}

func (l *lexer) peekExpect(typ int, v []byte) bool {
	l.skipWhiteSpaces()
	return bytes.HasPrefix(l.src, v)
}

func (l *lexer) expect(typ int, v []byte) error {
	l.skipWhiteSpaces()
	if !bytes.Equal(v, l.src[l.pos:len(v)]) {
		return fmt.Errorf("expected %q, but was not foud", v)
	}
	l.src = l.src[l.pos+len(v):]
	l.pos = 0
	l.peekPos = nil
	l.emit(typ, string(v))
	return nil
}

func (l *lexer) captureNumericLike(sb *strings.Builder) {
	for {
		// first character
		r := l.peek()
		if !unicode.IsNumber(r) {
			l.unread()
			break
		}
		sb.WriteRune(r)
	}
}

func (l *lexer) captureNumeric() error {
	l.skipWhiteSpaces()
	var sb strings.Builder

	r := l.peek()
	if r == '-' {
		sb.WriteRune(r)
	} else {
		l.unread()
	}

	l.captureNumericLike(&sb)

	r = l.peek()
	if r == '.' {
		sb.WriteRune(r)
		l.captureNumericLike(&sb)
	} else {
		l.unread()
	}
	l.emit(Numeric, sb.String())
	l.advance()
	return nil
}

func (l *lexer) captureIdent() error {
	l.skipWhiteSpaces()
	var sb strings.Builder
	for len(l.src) > 0 {
		r := l.peek()

		// We need to allow $ in the first character because it's used in the
		// special variables like $fn
		if sb.Len() > 0 {
			if r != '_' && !unicode.IsLetter(r) && !unicode.IsNumber(r) {
				l.unread()
				break
			}
		} else {
			if r != '$' && r != '_' && !unicode.IsLetter(r) && !unicode.IsNumber(r) {
				l.unread()
				break
			}
		}
		l.advance()
		sb.WriteRune(r)
	}

	if sb.Len() > 0 {
		l.emit(Ident, sb.String())
	}
	return nil
}

func (l *lexer) captureLiteral(begin, end rune) error {
	l.skipWhiteSpaces()
	var sb strings.Builder

	if l.peek() != begin {
		return fmt.Errorf("expected %q, but was not found", begin)
	}
	for len(l.src) > 0 {
		r := l.peek()
		if r == end {
			break
		}
		sb.WriteRune(r)
	}
	l.advance()
	l.emit(Literal, sb.String())
	return nil
}

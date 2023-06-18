package parser_test

import (
	"testing"

	"github.com/lestrrat-go/openscad/parser"
)

func TestLexer(t *testing.T) {
	const src = `
module foo(a, b, c=0) {
	bar=1;
}
`

	ch := make(chan *parser.Token, 1)
	go parser.Lex(ch, []byte(src))

	for tok := range ch {
		t.Logf("%#v", tok)
	}
}

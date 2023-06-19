package openscad_test

import (
	"testing"

	"github.com/lestrrat-go/openscad"
)

func TestLexer(t *testing.T) {
	const src = `
module foo(a, b, c=0) {
	bar=1;
}
`

	ch := make(chan *openscad.Token, 1)
	go openscad.Lex(ch, []byte(src))

	for tok := range ch {
		t.Logf("%#v", tok)
	}
}

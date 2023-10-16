package parser

import (
	"strings"
	"testing"

	"github.com/sblinch/kdl-go/internal/tokenizer"
)

func asStr(tokens []tokenizer.Token) string {
	b := strings.Builder{}
	for _, t := range tokens {
		b.Write(t.Data)
	}
	return b.String()
}

func assertTokens(t *testing.T, tokens []tokenizer.Token, expect string) {
	got := asStr(tokens)
	if got != expect {
		t.Fatalf("expected %q, got %q", expect, got)
	}
}

func Test_recentTokens(t *testing.T) {
	r := recentTokens{}

	r.Add(tokenizer.Token{ID: tokenizer.BareIdentifier, Data: []byte("node")})
	assertTokens(t, r.Get(), "node")

	r.Add(tokenizer.Token{ID: tokenizer.Whitespace, Data: []byte(" ")})
	assertTokens(t, r.Get(), "node ")

	r.Add(tokenizer.Token{ID: tokenizer.BareIdentifier, Data: []byte("alpha")})
	assertTokens(t, r.Get(), "node alpha")

	r.Add(tokenizer.Token{ID: tokenizer.Whitespace, Data: []byte(" ")})
	assertTokens(t, r.Get(), "node alpha ")

	r.Add(tokenizer.Token{ID: tokenizer.BareIdentifier, Data: []byte("beta")})
	assertTokens(t, r.Get(), "node alpha beta")

	r.Add(tokenizer.Token{ID: tokenizer.Whitespace, Data: []byte(" ")})
	assertTokens(t, r.Get(), " alpha beta ")

	r.Add(tokenizer.Token{ID: tokenizer.BareIdentifier, Data: []byte("charlie")})
	assertTokens(t, r.Get(), "alpha beta charlie")

	r.Add(tokenizer.Token{ID: tokenizer.Whitespace, Data: []byte(" ")})
	assertTokens(t, r.Get(), " beta charlie ")

	r.Add(tokenizer.Token{ID: tokenizer.BareIdentifier, Data: []byte("delta")})
	assertTokens(t, r.Get(), "beta charlie delta")

	r.Add(tokenizer.Token{ID: tokenizer.Newline, Data: []byte("\n")})
	assertTokens(t, r.Get(), " charlie delta\n")
}

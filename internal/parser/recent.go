package parser

import (
	"github.com/sblinch/kdl-go/internal/tokenizer"
)

const recentTokensCount = 5

type recentTokens struct {
	tokens [recentTokensCount]tokenizer.Token
	head   int
}

func (r *recentTokens) Add(t tokenizer.Token) {
	r.head = (r.head + 1) % recentTokensCount
	r.tokens[r.head] = t
}

func (r *recentTokens) Get() []tokenizer.Token {
	consec := make([]tokenizer.Token, 0, len(r.tokens))
	n := r.head
	for i := 0; i < len(r.tokens); i++ {
		n = (n + 1) % recentTokensCount
		if r.tokens[n].ID != tokenizer.Unknown {
			consec = append(consec, r.tokens[n])
		}
	}
	return consec
}

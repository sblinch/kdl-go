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

func (r *recentTokens) TrailingNewlines() []byte {
	var white []byte
	n := r.head
	lastWasSignificant := false
	for i := 0; i < len(r.tokens)-1; i++ {
		n = (n + 1) % recentTokensCount

		id := r.tokens[n].ID
		if id == tokenizer.Newline {
			spc := r.tokens[n].Data
			if lastWasSignificant && len(spc) > 0 {
				spc = spc[1:]
			}
			if len(spc) > 0 {
				if white == nil {
					white = make([]byte, 0, 16)
				}
				white = append(white, spc...)
			}
		} else if id != tokenizer.Whitespace && len(white) > 0 {
			white = white[:0]
		}

		lastWasSignificant = !(id == tokenizer.Newline || id == tokenizer.Whitespace || id == tokenizer.SingleLineComment || id == tokenizer.MultiLineComment)
	}
	return white
}

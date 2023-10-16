package document

import (
	"testing"
)

func TestQuoteString(t *testing.T) {
	tests := []struct {
		s    string
		want string
	}{
		{"This is a test", `"This is a test"`},
		{"This \"is\" a test", `"This \"is\" a test"`},
		{"This is\ta test", `"This is\ta test"`},
		{"This is a test\t", `"This is a test\t"`},
		{"This is a test\\", `"This is a test\\"`},
	}
	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			if got := QuoteString(tt.s); got != tt.want {
				t.Errorf("QuoteString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnquoteString(t *testing.T) {
	tests := []struct {
		s       string
		want    string
		wantErr bool
	}{
		{`"This is a test"`, "This is a test", false},
		{`"This \"is\" a test"`, "This \"is\" a test", false},
		{`"This is\ta test"`, "This is\ta test", false},
		{`"This is a test\t"`, "This is a test\t", false},
		{`"This is a test\"`, "", true},
		{`""`, "", false},
		{`"`, "", true},
		{`"x"`, "x", false},
		{`"x\t"`, "x\t", false},
		{`"\tx"`, "\tx", false},
		{`"\t"`, "\t", false},
		{`"This is a test😀"`, "This is a test😀", false},
		{`"This is a test\t😀"`, "This is a test\t😀", false},
		{`"\u{0020}"`, " ", false},
		{`"\u{1000000}"`, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			got, err := UnquoteString(tt.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnquoteString() error = %v", err)
			}
			if err == nil && got != tt.want {
				t.Errorf("UnquoteString() = %v, want %v", got, tt.want)
			}
		})
	}
}

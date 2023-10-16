package document

import (
	"testing"
)

func Test_rawString(t *testing.T) {
	tests := []struct {
		s    string
		want string
	}{
		{`This is a test`, `r"This is a test"`},
		{`This "is" a test`, `r#"This "is" a test"#`},
		{`This #"is"# a test`, `r##"This #"is"# a test"##`},
	}
	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			if got := rawString(tt.s); got != tt.want {
				t.Errorf("rawString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseRawString(t *testing.T) {
	tests := []struct {
		s       string
		want    string
		wantErr bool
	}{
		{`r#"[id="node-node"]"#`, `[id="node-node"]`, false},
	}
	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			got, err := parseRawString([]byte(tt.s))
			if (err != nil) != tt.wantErr {
				t.Errorf("parseRawString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseRawString() got = %v, want %v", got, tt.want)
			}
		})
	}
}

package marshaler

import (
	"testing"
	"time"
)

func Test_parseHMSDuration(t *testing.T) {
	mustDuration := func(s string) time.Duration {
		d, err := time.ParseDuration(s)
		if err != nil {
			panic(err.Error())
		}
		return d
	}
	tests := []struct {
		hms     string
		want    time.Duration
		wantErr bool
	}{
		{"01:32:47", mustDuration("1h32m47s"), false},
		{"1:32:47", mustDuration("1h32m47s"), false},
		{"1:32:47.0", mustDuration("1h32m47s"), false},
		{"0:32:47", mustDuration("0h32m47s"), false},
		{"1:32:47.000000000", mustDuration("1h32m47s"), false},
		{"1:32:47.000000123", mustDuration("1h32m47.000000123s"), false},
	}
	for _, tt := range tests {
		t.Run(tt.hms, func(t *testing.T) {
			got, err := parseHMSDuration(tt.hms)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseHMSDuration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseHMSDuration() got = %v, want %v", got, tt.want)
			}
		})
	}
}

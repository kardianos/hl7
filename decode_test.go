package hl7

import (
	"testing"
	"time"
)

func Test_lineDecoder_parseDateTime(t *testing.T) {
	tests := []struct {
		name    string
		dt      string
		want    time.Time
		wantErr bool
	}{
		{"year only", "2006", time.Date(2006, 1, 1, 0, 0, 0, 0, time.Local), false},
		{"date only", "20060203", time.Date(2006, 2, 3, 0, 0, 0, 0, time.Local), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDateTime(tt.dt)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDateTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !got.Equal(tt.want) {
				t.Errorf("parseDateTime() got = %v, want %v", got, tt.want)
			}
		})
	}
}

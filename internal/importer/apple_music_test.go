package importer

import (
	"testing"
	"time"
)

func TestParseISO8601Time(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  time.Time
	}{
		{
			name:  "RFC3339 with nanoseconds",
			input: "2024-10-20T18:31:29.435Z",
			want:  time.Date(2024, 10, 20, 18, 31, 29, 435000000, time.UTC),
		},
		{
			name:  "RFC3339 without nanoseconds",
			input: "2024-09-25T12:37:25Z",
			want:  time.Date(2024, 9, 25, 12, 37, 25, 0, time.UTC),
		},
		{
			name:  "empty string",
			input: "",
			want:  time.Time{},
		},
		{
			name:  "invalid format",
			input: "not-a-timestamp",
			want:  time.Time{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseISO8601Time(tt.input)
			if got.IsZero() != tt.want.IsZero() {
				t.Errorf("parseISO8601Time(%q) = %v, want zero=%v", tt.input, got, tt.want.IsZero())
			}
			if !got.IsZero() && !got.Equal(tt.want) {
				t.Errorf("parseISO8601Time(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseAppleMusicDuration(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int32
	}{
		{"zero", "0", 0},
		{"one second", "1000", 1},
		{"1484 ms", "1484", 1},
		{"193124 ms", "193124", 193},
		{"263813 ms", "263813", 263},
		{"empty", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseAppleMusicDuration(tt.input)
			if got != tt.want {
				t.Errorf("parseAppleMusicDuration(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

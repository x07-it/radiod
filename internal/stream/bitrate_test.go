package stream

import "testing"

// TestParseBitrate verifies conversion from human-readable bitrate to bytes per second.
func TestParseBitrate(t *testing.T) {
	tests := []struct {
		in   string
		want int
	}{
		{"96k", 12000},
		{"128000", 16000},
		{"bad", 16000}, // fallback to 128k
	}
	for _, tt := range tests {
		if got := parseBitrate(tt.in); got != tt.want {
			t.Errorf("parseBitrate(%q)=%d want %d", tt.in, got, tt.want)
		}
	}
}

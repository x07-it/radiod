package stream

import (
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

// parseBitrate converts values like "96k" or "128000" to bytes per second.
// If parsing fails, it falls back to 128kbps.
func parseBitrate(s string) int {
	br := strings.ToLower(strings.TrimSpace(s))
	multiplier := 1
	if strings.HasSuffix(br, "k") {
		multiplier = 1000
		br = strings.TrimSuffix(br, "k")
	}

	n, err := strconv.Atoi(br)
	if err != nil || n <= 0 {
		logrus.WithField("value", s).WithError(err).Warn("invalid bitrate, using 128k")
		n = 128
		multiplier = 1000
	}
	bitsPerSec := n * multiplier
	return bitsPerSec / 8
}

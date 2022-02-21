package utils

import (
	"crypto/sha256"
	"fmt"
	"math"
	"time"
)

// HashData generates a sha256 checksum of given data.
func HashData(data []byte) ([]byte, error) {
	h := sha256.New()
	_, err := h.Write(data)
	if err != nil {
		return nil, fmt.Errorf("failed to generate md5 hash: %w", err)
	}
	return h.Sum(nil), nil
}

// TimeRound will round down a time to the nearest minute.
func TimeRound(input time.Time) time.Time {
	return time.Date(
		input.Year(),
		input.Month(),
		input.Day(),
		input.Hour(),
		int(math.Floor(float64(input.Minute())/10)*10)+1,
		0,
		0,
		input.Location(),
	)
}

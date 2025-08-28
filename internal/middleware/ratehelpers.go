package middleware

import (
	"time"

	"golang.org/x/time/rate"
)

// PerMinute converts a requests-per-minute integer to rate.Limit
func PerMinute(n int) rate.Limit {
	if n <= 0 {
		return rate.Limit(0)
	}
	return rate.Every(time.Minute / time.Duration(n))
}

// PerSecond converts requests-per-second to rate.Limit
func PerSecond(n int) rate.Limit {
	if n <= 0 {
		return rate.Limit(0)
	}
	return rate.Limit(n)
}

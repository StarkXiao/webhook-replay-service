package retry

import (
	"math"
	"net/http"
	"strings"
	"time"
)

type Decision struct {
	Retry  bool
	Delay  time.Duration
	Reason string
}
type Classifier struct {
	Retryable                  map[int]bool
	RetryNetwork, RetryTimeout bool
}

func DefaultClassifier() Classifier {
	return Classifier{Retryable: map[int]bool{408: true, 425: true, 429: true, 500: true, 502: true, 503: true, 504: true}, RetryNetwork: true, RetryTimeout: true}
}
func (c Classifier) Classify(code int, err error, p Policy, attempt int) Decision {
	if err != nil {
		return Decision{Retry: c.RetryNetwork, Delay: p.Delay(attempt), Reason: err.Error()}
	}
	if c.Retryable[code] {
		return Decision{Retry: true, Delay: p.Delay(attempt), Reason: http.StatusText(code)}
	}
	return Decision{Reason: http.StatusText(code)}
}
func Exponential(initial, max time.Duration, attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}
	if initial <= 0 {
		initial = time.Second
	}
	if max <= 0 {
		max = 5 * time.Minute
	}
	value := float64(initial) * math.Pow(2, float64(attempt))
	if value > float64(max) {
		return max
	}
	return time.Duration(value)
}
func ParseRetryAfter(value string, now time.Time) time.Duration {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	if n, err := time.ParseDuration(value + "s"); err == nil {
		return n
	}
	if at, err := http.ParseTime(value); err == nil && at.After(now) {
		return at.Sub(now)
	}
	return 0
}
func BackoffTable(p Policy, n int) []time.Duration {
	out := make([]time.Duration, n)
	for i := range out {
		out[i] = p.Delay(i)
	}
	return out
}
func AttemptsAllowed(current, max int) bool { return max < 0 || current < max }

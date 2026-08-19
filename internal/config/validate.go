package config

import (
	"fmt"
	"strings"
	"time"
)

func (c Config) Validate() error {
	if strings.TrimSpace(c.Port) == "" {
		return fmt.Errorf("port is required")
	}
	if c.MaxRetries < 0 {
		return fmt.Errorf("max retries cannot be negative")
	}
	if c.InitialBackoff <= 0 || c.MaxBackoff < c.InitialBackoff {
		return fmt.Errorf("invalid backoff configuration")
	}
	if c.WorkerCount <= 0 {
		return fmt.Errorf("worker count must be positive")
	}
	return nil
}
func (c Config) RetryWindow() time.Duration {
	total := time.Duration(0)
	for i := 0; i < c.MaxRetries; i++ {
		d := c.InitialBackoff << i
		if d > c.MaxBackoff {
			d = c.MaxBackoff
		}
		total += d
	}
	return total
}

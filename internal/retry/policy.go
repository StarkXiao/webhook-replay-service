package retry

import "time"

type Policy struct{ Initial, Max time.Duration }

func (p Policy) Delay(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}
	d := p.Initial
	for i := 0; i < attempt; i++ {
		if d >= p.Max/2 {
			return p.Max
		}
		d *= 2
	}
	if d > p.Max {
		return p.Max
	}
	return d
}
func (p Policy) Should(status int, networkErr bool) bool {
	return networkErr || status == 408 || status == 425 || status == 429 || status >= 500
}

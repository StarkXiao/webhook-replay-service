package delivery

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Signature struct {
	Header    string
	Secret    []byte
	Tolerance time.Duration
}

func (s Signature) Sign(body []byte, at time.Time) string {
	mac := hmac.New(sha256.New, s.Secret)
	_, _ = mac.Write([]byte(fmt.Sprintf("%d.", at.Unix())))
	_, _ = mac.Write(body)
	return fmt.Sprintf("t=%d,v1=%s", at.Unix(), hex.EncodeToString(mac.Sum(nil)))
}
func (s Signature) Verify(header string, body []byte, now time.Time) error {
	var ts int64
	var value string
	for _, p := range strings.Split(header, ",") {
		x := strings.SplitN(strings.TrimSpace(p), "=", 2)
		if len(x) != 2 {
			continue
		}
		if x[0] == "t" {
			_, _ = fmt.Sscan(x[1], &ts)
		}
		if x[0] == "v1" {
			value = x[1]
		}
	}
	if ts == 0 || value == "" {
		return fmt.Errorf("malformed signature")
	}
	if s.Tolerance > 0 && (now.Sub(time.Unix(ts, 0)) > s.Tolerance || time.Unix(ts, 0).Sub(now) > s.Tolerance) {
		return fmt.Errorf("signature expired")
	}
	want := s.Sign(body, time.Unix(ts, 0))
	want = strings.TrimPrefix(strings.Split(want, ",")[1], "v1=")
	got, e := hex.DecodeString(value)
	if e != nil {
		return fmt.Errorf("invalid signature encoding")
	}
	expected, _ := hex.DecodeString(want)
	if !hmac.Equal(got, expected) {
		return fmt.Errorf("signature mismatch")
	}
	return nil
}
func ApplyHeaders(req *http.Request, headers map[string]string) {
	for k, v := range headers {
		if strings.EqualFold(k, "host") || strings.EqualFold(k, "content-length") {
			continue
		}
		req.Header.Set(k, v)
	}
}
func ResponseSummary(status int, body []byte, max int) string {
	if max <= 0 {
		max = 512
	}
	if len(body) > max {
		body = body[:max]
	}
	return fmt.Sprintf("status=%d body=%q", status, string(body))
}

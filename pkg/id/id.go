package id

import (
	"crypto/rand"
	"encoding/hex"
)

// New returns a cryptographically random identifier. crypto/rand.Read only
// returns an error when it cannot fill the supplied buffer.
func New() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand unavailable: " + err.Error())
	}
	return hex.EncodeToString(b)
}

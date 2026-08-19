package domain

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

type Cursor struct {
	CreatedAt time.Time `json:"created_at"`
	ID        string    `json:"id"`
	Direction string    `json:"direction"`
}

func (c Cursor) Encode() string {
	b, _ := json.Marshal(c)
	return base64.RawURLEncoding.EncodeToString(b)
}
func DecodeCursor(value string) (Cursor, error) {
	if value == "" {
		return Cursor{}, nil
	}
	b, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return Cursor{}, fmt.Errorf("invalid cursor: %w", err)
	}
	var c Cursor
	if err = json.Unmarshal(b, &c); err != nil {
		return Cursor{}, fmt.Errorf("invalid cursor: %w", err)
	}
	if c.Direction != "" && c.Direction != "next" && c.Direction != "previous" {
		return Cursor{}, fmt.Errorf("invalid cursor direction")
	}
	return c, nil
}
func NewCursor(e WebhookEvent, direction string) Cursor {
	return Cursor{CreatedAt: e.CreatedAt, ID: e.ID, Direction: direction}
}
func (c Cursor) Empty() bool { return c.ID == "" && c.CreatedAt.IsZero() }

package transport

import (
	"encoding/json"
	"net/http"
)

func Decode(r *http.Request, v any) error { return json.NewDecoder(r.Body).Decode(v) }
func Encode(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

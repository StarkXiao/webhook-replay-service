package handler

import (
	"errors"
	"net/http"

	"github.com/StarkXiao/webhook-replay-service/internal/repository"
	"github.com/StarkXiao/webhook-replay-service/internal/validator"
)

// classifyReceiveError maps a service.Receive error to an HTTP status and a
// stable error code. It lets the API layer distinguish duplicate writes,
// repository unavailability, and parameter validation failures without parsing
// the error string. Scoped to Receive errors only — ClaimTask reuses
// repository.ErrDuplicate with different semantics and must not share this path.
func classifyReceiveError(err error) (status int, code string) {
	var verrs validator.Errors
	if errors.As(err, &verrs) {
		return http.StatusBadRequest, "validation_failed"
	}
	if errors.Is(err, repository.ErrDuplicate) {
		return http.StatusConflict, "duplicate_event"
	}
	// Any other error originates from the repository layer (storage
	// unavailable, connection loss, constraint, etc.). Surface it as a
	// retryable 503 rather than masking it as a client-side 400.
	return http.StatusServiceUnavailable, "repository_unavailable"
}

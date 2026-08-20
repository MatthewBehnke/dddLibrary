package rest

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/matthewbehnke/dddlibrary/internal/lending/domain"
)

// errorResponse is the JSON body returned for any non-2xx result.
type errorResponse struct {
	Error string `json:"error"`
}

// statusForError maps a domain error to an HTTP status code. This is the single
// place transport concerns meet domain vocabulary; the core stays HTTP-agnostic.
func statusForError(err error) int {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound // 404
	case errors.Is(err, domain.ErrBookUnavailable):
		return http.StatusConflict // 409
	case errors.Is(err, domain.ErrLoanLimitReached),
		errors.Is(err, domain.ErrMembershipInactive):
		return http.StatusUnprocessableEntity // 422
	default:
		return http.StatusInternalServerError // 500
	}
}

// writeJSON encodes v as JSON with the given status.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError maps err to a status and writes a JSON error body.
func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, errorResponse{Error: err.Error()})
}

package domain

import "errors"

// Domain errors are typed sentinels. They express rule violations in domain
// terms and know nothing about transport. Adapters (e.g. HTTP) map each to a
// protocol-specific status; the core never imports net/http.
var (
	// ErrBookUnavailable means the requested Book is already on loan.
	ErrBookUnavailable = errors.New("book is not available")
	// ErrLoanLimitReached means the Member already holds the maximum number
	// of active loans.
	ErrLoanLimitReached = errors.New("member has reached the active loan limit")
	// ErrMembershipInactive means the Member's membership is not active.
	ErrMembershipInactive = errors.New("member's membership is not active")
	// ErrNotFound means a requested aggregate does not exist.
	ErrNotFound = errors.New("not found")
	// ErrInvalidDueDate means a due date was not strictly after its reference
	// instant.
	ErrInvalidDueDate = errors.New("due date must be in the future")
)

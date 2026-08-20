package domain

import "time"

// LoanPeriod is how long a book may be borrowed before it is due back.
const LoanPeriod = 14 * 24 * time.Hour

// DueDate is a value object: an instant by which a borrowed Book must be
// returned. It is immutable and always strictly after the moment it was
// created relative to.
type DueDate struct {
	t time.Time
}

// NewDueDate builds a DueDate, rejecting any instant not strictly after now.
func NewDueDate(due, now time.Time) (DueDate, error) {
	if !due.After(now) {
		return DueDate{}, ErrInvalidDueDate
	}
	return DueDate{t: due}, nil
}

// DueDateFrom returns the DueDate for a loan starting at now, i.e. now plus
// LoanPeriod. It is always valid.
func DueDateFrom(now time.Time) DueDate {
	return DueDate{t: now.Add(LoanPeriod)}
}

// Time returns the underlying instant.
func (d DueDate) Time() time.Time { return d.t }

// IsZero reports whether the DueDate is the zero value.
func (d DueDate) IsZero() bool { return d.t.IsZero() }

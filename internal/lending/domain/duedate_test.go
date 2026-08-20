package domain

import (
	"errors"
	"testing"
	"time"
)

func TestNewDueDate_rejectsNonFuture(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	cases := map[string]time.Time{
		"in the past":  now.Add(-time.Hour),
		"equal to now": now,
	}
	for name, due := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := NewDueDate(due, now); !errors.Is(err, ErrInvalidDueDate) {
				t.Fatalf("want ErrInvalidDueDate, got %v", err)
			}
		})
	}
}

func TestNewDueDate_acceptsFuture(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	due := now.Add(time.Hour)

	d, err := NewDueDate(due, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !d.Time().Equal(due) {
		t.Fatalf("want %v, got %v", due, d.Time())
	}
}

func TestDueDateFrom_isNowPlusLoanPeriod(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	d := DueDateFrom(now)

	want := now.Add(LoanPeriod)
	if !d.Time().Equal(want) {
		t.Fatalf("want %v, got %v", want, d.Time())
	}
}

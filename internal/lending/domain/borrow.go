package domain

import "time"

// Borrow is a domain service coordinating the act of lending: it spans three
// aggregates (Member, Book, Loan) that no single one of them owns, so the rule
// lives here rather than on any one entity.
//
// It enforces, in order:
//  1. the Member's membership is active (ErrMembershipInactive),
//  2. the Member is under the active-loan limit (ErrLoanLimitReached),
//  3. the Book is available (ErrBookUnavailable).
//
// On success it mutates book and member state and returns a new active Loan
// due LoanPeriod after now. On failure it mutates nothing.
func Borrow(book *Book, member *Member, now time.Time) (*Loan, error) {
	if err := member.CanBorrow(); err != nil {
		return nil, err
	}
	if err := book.markOnLoan(); err != nil {
		return nil, err
	}
	member.recordBorrow()

	loan := &Loan{
		id:         NewLoanID(),
		bookID:     book.ID(),
		memberID:   member.ID(),
		borrowedAt: now,
		dueDate:    DueDateFrom(now),
	}
	return loan, nil
}

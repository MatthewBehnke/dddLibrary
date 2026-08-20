package domain

import "time"

// Loan is an aggregate root recording that a Member has borrowed a Book, when,
// and when it is due back. A nil returnedAt means the loan is still active.
type Loan struct {
	id         LoanID
	bookID     BookID
	memberID   MemberID
	borrowedAt time.Time
	dueDate    DueDate
	returnedAt *time.Time
}

// ReconstituteLoan rebuilds a Loan from persisted state. Repositories use this;
// application code should not.
func ReconstituteLoan(id LoanID, bookID BookID, memberID MemberID, borrowedAt time.Time, dueDate DueDate, returnedAt *time.Time) *Loan {
	return &Loan{
		id:         id,
		bookID:     bookID,
		memberID:   memberID,
		borrowedAt: borrowedAt,
		dueDate:    dueDate,
		returnedAt: returnedAt,
	}
}

func (l *Loan) ID() LoanID            { return l.id }
func (l *Loan) BookID() BookID        { return l.bookID }
func (l *Loan) MemberID() MemberID    { return l.memberID }
func (l *Loan) BorrowedAt() time.Time { return l.borrowedAt }
func (l *Loan) DueDate() DueDate      { return l.dueDate }

// ReturnedAt returns the return instant and whether the loan has been returned.
func (l *Loan) ReturnedAt() (time.Time, bool) {
	if l.returnedAt == nil {
		return time.Time{}, false
	}
	return *l.returnedAt, true
}

// IsActive reports whether the loan is still outstanding.
func (l *Loan) IsActive() bool { return l.returnedAt == nil }

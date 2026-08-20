// Package app holds the application layer: use cases that orchestrate domain
// aggregates through the repository ports. It is delivery-agnostic — no HTTP,
// gRPC, or SQL types appear here — so any driving adapter can invoke it.
package app

import (
	"context"
	"time"

	"github.com/matthewbehnke/dddlibrary/internal/lending/domain"
)

// Clock supplies the current time, injected so use cases are deterministic
// under test.
type Clock func() time.Time

// BorrowBook is the use case for a Member borrowing a Book.
type BorrowBook struct {
	books   domain.BookRepository
	members domain.MemberRepository
	loans   domain.LoanRepository
	now     Clock
}

// NewBorrowBook wires the use case with its ports and a clock. If now is nil,
// time.Now is used.
func NewBorrowBook(books domain.BookRepository, members domain.MemberRepository, loans domain.LoanRepository, now Clock) *BorrowBook {
	if now == nil {
		now = time.Now
	}
	return &BorrowBook{books: books, members: members, loans: loans, now: now}
}

// BorrowBookCommand carries the resolved identities to borrow. Parsing raw
// input into typed IDs is the driving adapter's job.
type BorrowBookCommand struct {
	BookID   domain.BookID
	MemberID domain.MemberID
}

// BorrowBookResult is the use case output, in plain types for any adapter to
// render.
type BorrowBookResult struct {
	LoanID  domain.LoanID
	DueDate time.Time
}

// Execute loads the aggregates, applies the Borrow domain service, and persists
// the results. It surfaces domain errors (ErrNotFound, ErrBookUnavailable,
// ErrLoanLimitReached, ErrMembershipInactive) unchanged for the adapter to map.
func (uc *BorrowBook) Execute(ctx context.Context, cmd BorrowBookCommand) (BorrowBookResult, error) {
	book, err := uc.books.FindByID(ctx, cmd.BookID)
	if err != nil {
		return BorrowBookResult{}, err
	}
	member, err := uc.members.FindByID(ctx, cmd.MemberID)
	if err != nil {
		return BorrowBookResult{}, err
	}

	loan, err := domain.Borrow(book, member, uc.now())
	if err != nil {
		return BorrowBookResult{}, err
	}

	if err := uc.loans.Save(ctx, loan); err != nil {
		return BorrowBookResult{}, err
	}
	if err := uc.books.Save(ctx, book); err != nil {
		return BorrowBookResult{}, err
	}
	if err := uc.members.Save(ctx, member); err != nil {
		return BorrowBookResult{}, err
	}

	return BorrowBookResult{LoanID: loan.ID(), DueDate: loan.DueDate().Time()}, nil
}

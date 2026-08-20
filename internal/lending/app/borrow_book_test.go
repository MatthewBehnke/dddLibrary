package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/matthewbehnke/dddlibrary/internal/lending/adapters/memory"
	"github.com/matthewbehnke/dddlibrary/internal/lending/app"
	"github.com/matthewbehnke/dddlibrary/internal/lending/domain"
)

var fixedNow = time.Date(2026, 5, 10, 9, 0, 0, 0, time.UTC)

type fixtures struct {
	books   *memory.BookRepository
	members *memory.MemberRepository
	loans   *memory.LoanRepository
	uc      *app.BorrowBook
}

func newFixtures(t *testing.T) fixtures {
	t.Helper()
	books := memory.NewBookRepository()
	members := memory.NewMemberRepository()
	loans := memory.NewLoanRepository()
	uc := app.NewBorrowBook(books, members, loans, func() time.Time { return fixedNow })
	return fixtures{books: books, members: members, loans: loans, uc: uc}
}

func seedBook(t *testing.T, f fixtures, available bool) domain.BookID {
	t.Helper()
	b := domain.ReconstituteBook(domain.NewBookID(), "SICP", available)
	if err := f.books.Save(context.Background(), b); err != nil {
		t.Fatalf("seed book: %v", err)
	}
	return b.ID()
}

func seedMember(t *testing.T, f fixtures, status domain.MembershipStatus, active int) domain.MemberID {
	t.Helper()
	m := domain.ReconstituteMember(domain.NewMemberID(), "Ada", status, active)
	if err := f.members.Save(context.Background(), m); err != nil {
		t.Fatalf("seed member: %v", err)
	}
	return m.ID()
}

func TestBorrowBook_success(t *testing.T) {
	f := newFixtures(t)
	bookID := seedBook(t, f, true)
	memberID := seedMember(t, f, domain.MembershipActive, 0)

	res, err := f.uc.Execute(context.Background(), app.BorrowBookCommand{BookID: bookID, MemberID: memberID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if want := fixedNow.Add(domain.LoanPeriod); !res.DueDate.Equal(want) {
		t.Errorf("due date = %v, want %v", res.DueDate, want)
	}
	if f.loans.Count() != 1 {
		t.Errorf("loans stored = %d, want 1", f.loans.Count())
	}
	// The book must be persisted as unavailable.
	got, err := f.books.FindByID(context.Background(), bookID)
	if err != nil {
		t.Fatalf("reload book: %v", err)
	}
	if got.IsAvailable() {
		t.Error("book should be persisted as unavailable")
	}
	// The member's active-loan count must be persisted.
	gotM, err := f.members.FindByID(context.Background(), memberID)
	if err != nil {
		t.Fatalf("reload member: %v", err)
	}
	if gotM.ActiveLoans() != 1 {
		t.Errorf("member active loans = %d, want 1", gotM.ActiveLoans())
	}
}

func TestBorrowBook_bookNotFound(t *testing.T) {
	f := newFixtures(t)
	memberID := seedMember(t, f, domain.MembershipActive, 0)

	_, err := f.uc.Execute(context.Background(), app.BorrowBookCommand{BookID: domain.NewBookID(), MemberID: memberID})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestBorrowBook_memberNotFound(t *testing.T) {
	f := newFixtures(t)
	bookID := seedBook(t, f, true)

	_, err := f.uc.Execute(context.Background(), app.BorrowBookCommand{BookID: bookID, MemberID: domain.NewMemberID()})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestBorrowBook_propagatesDomainRuleViolation(t *testing.T) {
	f := newFixtures(t)
	bookID := seedBook(t, f, false) // unavailable
	memberID := seedMember(t, f, domain.MembershipActive, 0)

	_, err := f.uc.Execute(context.Background(), app.BorrowBookCommand{BookID: bookID, MemberID: memberID})
	if !errors.Is(err, domain.ErrBookUnavailable) {
		t.Fatalf("want ErrBookUnavailable, got %v", err)
	}
	if f.loans.Count() != 0 {
		t.Error("no loan should be stored when borrowing fails")
	}
}

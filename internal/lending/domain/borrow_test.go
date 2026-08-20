package domain

import (
	"errors"
	"testing"
	"time"
)

var borrowNow = time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)

func activeMember() *Member {
	return ReconstituteMember(NewMemberID(), "Ada", MembershipActive, 0)
}

func TestBorrow_success(t *testing.T) {
	book := NewBook("Domain-Driven Design")
	member := activeMember()

	loan, err := Borrow(book, member, borrowNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if loan.BookID() != book.ID() || loan.MemberID() != member.ID() {
		t.Error("loan should reference the borrowed book and member")
	}
	if !loan.BorrowedAt().Equal(borrowNow) {
		t.Errorf("borrowedAt = %v, want %v", loan.BorrowedAt(), borrowNow)
	}
	if want := borrowNow.Add(LoanPeriod); !loan.DueDate().Time().Equal(want) {
		t.Errorf("dueDate = %v, want %v", loan.DueDate().Time(), want)
	}
	if !loan.IsActive() {
		t.Error("a fresh loan should be active")
	}
	if book.IsAvailable() {
		t.Error("book should be marked unavailable after borrowing")
	}
	if member.ActiveLoans() != 1 {
		t.Errorf("member active loans = %d, want 1", member.ActiveLoans())
	}
}

func TestBorrow_rejectsInactiveMembership(t *testing.T) {
	book := NewBook("Refactoring")
	member := ReconstituteMember(NewMemberID(), "Grace", MembershipInactive, 0)

	_, err := Borrow(book, member, borrowNow)

	if !errors.Is(err, ErrMembershipInactive) {
		t.Fatalf("want ErrMembershipInactive, got %v", err)
	}
	if !book.IsAvailable() {
		t.Error("book must stay available when borrowing fails")
	}
}

func TestBorrow_rejectsAtLoanLimit(t *testing.T) {
	book := NewBook("The Pragmatic Programmer")
	member := ReconstituteMember(NewMemberID(), "Linus", MembershipActive, MaxActiveLoans)

	_, err := Borrow(book, member, borrowNow)

	if !errors.Is(err, ErrLoanLimitReached) {
		t.Fatalf("want ErrLoanLimitReached, got %v", err)
	}
	if !book.IsAvailable() {
		t.Error("book must stay available when borrowing fails")
	}
}

func TestBorrow_rejectsUnavailableBook(t *testing.T) {
	book := ReconstituteBook(NewBookID(), "Clean Architecture", false)
	member := activeMember()

	_, err := Borrow(book, member, borrowNow)

	if !errors.Is(err, ErrBookUnavailable) {
		t.Fatalf("want ErrBookUnavailable, got %v", err)
	}
	if member.ActiveLoans() != 0 {
		t.Error("member loan count must not change when borrowing fails")
	}
}

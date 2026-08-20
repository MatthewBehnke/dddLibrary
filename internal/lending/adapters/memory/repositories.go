// Package memory provides in-memory implementations of the domain repository
// ports. It backs fast unit tests and can serve as a zero-dependency dev mode.
// Each repository is safe for concurrent use.
package memory

import (
	"context"
	"sync"

	"github.com/matthewbehnke/dddlibrary/internal/lending/domain"
)

// BookRepository is an in-memory domain.BookRepository.
type BookRepository struct {
	mu    sync.RWMutex
	books map[domain.BookID]*domain.Book
}

// NewBookRepository returns an empty in-memory book repository.
func NewBookRepository() *BookRepository {
	return &BookRepository{books: make(map[domain.BookID]*domain.Book)}
}

func (r *BookRepository) FindByID(_ context.Context, id domain.BookID) (*domain.Book, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	b, ok := r.books[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return domain.ReconstituteBook(b.ID(), b.Title(), b.IsAvailable()), nil
}

func (r *BookRepository) Save(_ context.Context, book *domain.Book) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.books[book.ID()] = domain.ReconstituteBook(book.ID(), book.Title(), book.IsAvailable())
	return nil
}

// MemberRepository is an in-memory domain.MemberRepository.
type MemberRepository struct {
	mu      sync.RWMutex
	members map[domain.MemberID]*domain.Member
}

// NewMemberRepository returns an empty in-memory member repository.
func NewMemberRepository() *MemberRepository {
	return &MemberRepository{members: make(map[domain.MemberID]*domain.Member)}
}

func (r *MemberRepository) FindByID(_ context.Context, id domain.MemberID) (*domain.Member, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	m, ok := r.members[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return domain.ReconstituteMember(m.ID(), m.Name(), m.Status(), m.ActiveLoans()), nil
}

func (r *MemberRepository) Save(_ context.Context, member *domain.Member) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.members[member.ID()] = domain.ReconstituteMember(member.ID(), member.Name(), member.Status(), member.ActiveLoans())
	return nil
}

// LoanRepository is an in-memory domain.LoanRepository.
type LoanRepository struct {
	mu    sync.RWMutex
	loans map[domain.LoanID]*domain.Loan
}

// NewLoanRepository returns an empty in-memory loan repository.
func NewLoanRepository() *LoanRepository {
	return &LoanRepository{loans: make(map[domain.LoanID]*domain.Loan)}
}

func (r *LoanRepository) Save(_ context.Context, loan *domain.Loan) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.loans[loan.ID()] = loan
	return nil
}

// Count reports how many loans are stored. Test helper.
func (r *LoanRepository) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.loans)
}

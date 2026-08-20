// Package postgres provides pgx-backed implementations of the domain repository
// ports. It is a driven adapter: it depends inward on the domain, never the
// reverse. All aggregate identities are stored as native uuid columns.
package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/matthewbehnke/dddlibrary/internal/lending/domain"
)

// BookRepository is a pgx-backed domain.BookRepository.
type BookRepository struct {
	pool *pgxpool.Pool
}

// NewBookRepository builds a BookRepository over the given pool.
func NewBookRepository(pool *pgxpool.Pool) *BookRepository { return &BookRepository{pool: pool} }

func (r *BookRepository) FindByID(ctx context.Context, id domain.BookID) (*domain.Book, error) {
	const q = `SELECT id, title, available FROM books WHERE id = $1`
	var (
		bookID    uuid.UUID
		title     string
		available bool
	)
	err := r.pool.QueryRow(ctx, q, uuid.UUID(id)).Scan(&bookID, &title, &available)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return domain.ReconstituteBook(domain.BookID(bookID), title, available), nil
}

func (r *BookRepository) Save(ctx context.Context, book *domain.Book) error {
	const q = `
		INSERT INTO books (id, title, available)
		VALUES ($1, $2, $3)
		ON CONFLICT (id) DO UPDATE SET title = EXCLUDED.title, available = EXCLUDED.available`
	_, err := r.pool.Exec(ctx, q, uuid.UUID(book.ID()), book.Title(), book.IsAvailable())
	return err
}

// MemberRepository is a pgx-backed domain.MemberRepository.
type MemberRepository struct {
	pool *pgxpool.Pool
}

// NewMemberRepository builds a MemberRepository over the given pool.
func NewMemberRepository(pool *pgxpool.Pool) *MemberRepository { return &MemberRepository{pool: pool} }

func (r *MemberRepository) FindByID(ctx context.Context, id domain.MemberID) (*domain.Member, error) {
	const q = `SELECT id, name, status, active_loans FROM members WHERE id = $1`
	var (
		memberID    uuid.UUID
		name        string
		status      string
		activeLoans int
	)
	err := r.pool.QueryRow(ctx, q, uuid.UUID(id)).Scan(&memberID, &name, &status, &activeLoans)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return domain.ReconstituteMember(domain.MemberID(memberID), name, domain.MembershipStatus(status), activeLoans), nil
}

func (r *MemberRepository) Save(ctx context.Context, member *domain.Member) error {
	const q = `
		INSERT INTO members (id, name, status, active_loans)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, status = EXCLUDED.status, active_loans = EXCLUDED.active_loans`
	_, err := r.pool.Exec(ctx, q, uuid.UUID(member.ID()), member.Name(), string(member.Status()), member.ActiveLoans())
	return err
}

// LoanRepository is a pgx-backed domain.LoanRepository.
type LoanRepository struct {
	pool *pgxpool.Pool
}

// NewLoanRepository builds a LoanRepository over the given pool.
func NewLoanRepository(pool *pgxpool.Pool) *LoanRepository { return &LoanRepository{pool: pool} }

func (r *LoanRepository) Save(ctx context.Context, loan *domain.Loan) error {
	const q = `
		INSERT INTO loans (id, book_id, member_id, borrowed_at, due_date, returned_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET returned_at = EXCLUDED.returned_at`

	var returnedAt *time.Time
	if t, ok := loan.ReturnedAt(); ok {
		returnedAt = &t
	}

	_, err := r.pool.Exec(ctx, q,
		uuid.UUID(loan.ID()),
		uuid.UUID(loan.BookID()),
		uuid.UUID(loan.MemberID()),
		loan.BorrowedAt(),
		loan.DueDate().Time(),
		returnedAt,
	)
	return err
}

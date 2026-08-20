package domain

import "context"

// The repository ports are the persistence-facing interfaces the domain owns.
// Adapters (in-memory, Postgres) implement them; the dependency rule points
// inward, so the domain never imports those adapters.
//
// FindByID implementations must return ErrNotFound when the aggregate is
// absent, so callers can branch on a domain error rather than a driver error.

// BookRepository loads and stores Book aggregates.
type BookRepository interface {
	FindByID(ctx context.Context, id BookID) (*Book, error)
	Save(ctx context.Context, book *Book) error
}

// MemberRepository loads and stores Member aggregates.
type MemberRepository interface {
	FindByID(ctx context.Context, id MemberID) (*Member, error)
	Save(ctx context.Context, member *Member) error
}

// LoanRepository stores Loan aggregates.
type LoanRepository interface {
	Save(ctx context.Context, loan *Loan) error
}

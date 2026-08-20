//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/matthewbehnke/dddlibrary/internal/lending/adapters/postgres"
	"github.com/matthewbehnke/dddlibrary/internal/lending/app"
	"github.com/matthewbehnke/dddlibrary/internal/lending/domain"
	"github.com/matthewbehnke/dddlibrary/internal/platform/migrate"
)

// setup starts a throwaway Postgres, applies migrations, and returns a pool.
func setup(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()

	ctr, err := tcpostgres.Run(ctx, "postgres:16-alpine",
		tcpostgres.WithDatabase("lending"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}
	testcontainers.CleanupContainer(t, ctr)

	dsn, err := ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("connection string: %v", err)
	}
	if err := migrate.Run(dsn); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("open pool: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func TestPostgres_borrowBookEndToEnd(t *testing.T) {
	ctx := context.Background()
	pool := setup(t)

	books := postgres.NewBookRepository(pool)
	members := postgres.NewMemberRepository(pool)
	loans := postgres.NewLoanRepository(pool)

	book := domain.NewBook("Structure and Interpretation of Computer Programs")
	if err := books.Save(ctx, book); err != nil {
		t.Fatalf("save book: %v", err)
	}
	member := domain.NewMember("Ada Lovelace")
	if err := members.Save(ctx, member); err != nil {
		t.Fatalf("save member: %v", err)
	}

	now := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	uc := app.NewBorrowBook(books, members, loans, func() time.Time { return now })

	res, err := uc.Execute(ctx, app.BorrowBookCommand{BookID: book.ID(), MemberID: member.ID()})
	if err != nil {
		t.Fatalf("borrow: %v", err)
	}
	if want := now.Add(domain.LoanPeriod); !res.DueDate.Equal(want) {
		t.Errorf("due date = %v, want %v", res.DueDate, want)
	}

	// Book must be persisted as unavailable.
	gotBook, err := books.FindByID(ctx, book.ID())
	if err != nil {
		t.Fatalf("reload book: %v", err)
	}
	if gotBook.IsAvailable() {
		t.Error("book should be unavailable in the database after borrowing")
	}

	// Member's active-loan count must be persisted.
	gotMember, err := members.FindByID(ctx, member.ID())
	if err != nil {
		t.Fatalf("reload member: %v", err)
	}
	if gotMember.ActiveLoans() != 1 {
		t.Errorf("active loans = %d, want 1", gotMember.ActiveLoans())
	}

	// The loan row must exist.
	var count int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM loans WHERE id = $1`, res.LoanID.String()).Scan(&count); err != nil {
		t.Fatalf("count loans: %v", err)
	}
	if count != 1 {
		t.Errorf("loan rows = %d, want 1", count)
	}
}

func TestPostgres_findMissingReturnsNotFound(t *testing.T) {
	ctx := context.Background()
	pool := setup(t)
	books := postgres.NewBookRepository(pool)

	_, err := books.FindByID(ctx, domain.NewBookID())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

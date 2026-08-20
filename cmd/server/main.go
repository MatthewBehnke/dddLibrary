// Command server is the composition root for the lending service. It is the one
// place that knows about every concrete adapter: it reads config, applies
// migrations, opens the Postgres pool, and wires repositories → use case →
// HTTP handler → router → server by hand (no DI framework).
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/matthewbehnke/dddlibrary/internal/lending/adapters/postgres"
	"github.com/matthewbehnke/dddlibrary/internal/lending/adapters/rest"
	"github.com/matthewbehnke/dddlibrary/internal/lending/app"
	"github.com/matthewbehnke/dddlibrary/internal/platform/config"
	"github.com/matthewbehnke/dddlibrary/internal/platform/migrate"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("server exited with error", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Apply schema migrations before serving.
	if err := migrate.Run(cfg.DatabaseURL); err != nil {
		return err
	}
	logger.Info("migrations applied")

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		return err
	}

	// Manual wiring: adapters (driven) → use case → adapter (driving).
	books := postgres.NewBookRepository(pool)
	members := postgres.NewMemberRepository(pool)
	loans := postgres.NewLoanRepository(pool)

	borrow := app.NewBorrowBook(books, members, loans, time.Now)

	router := rest.NewRouter(rest.NewLoanHandler(borrow))

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Serve until an interrupt, then shut down gracefully.
	shutdownCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	serveErr := make(chan error, 1)
	go func() {
		logger.Info("listening", "addr", cfg.HTTPAddr)
		serveErr <- srv.ListenAndServe()
	}()

	select {
	case err := <-serveErr:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-shutdownCtx.Done():
		logger.Info("shutting down")
		graceCtx, graceCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer graceCancel()
		return srv.Shutdown(graceCtx)
	}
}

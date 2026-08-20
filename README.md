# dddlibrary

A small Go outline of a library **lending** service built with Domain-Driven
Design and hexagonal (ports & adapters) clean architecture. It is a *tracer
bullet*: the full structure is in place and one use case — **borrow a book** —
is wired end-to-end (HTTP → application → domain → Postgres) so the architecture
actually compiles and serves a real request.

## Architecture

Dependencies point inward; the domain knows nothing about HTTP or SQL.

```
        driving adapter            core                     driven adapters
     ┌──────────────────┐   ┌───────────────────┐   ┌───────────────────────┐
HTTP │ adapters/rest     │→ │ app (use cases)    │→ │ domain ports           │
     │  POST /loans      │   │  BorrowBook       │   │  Book/Member/LoanRepo  │
     └──────────────────┘   │ domain (entities, │   ├───────────────────────┤
     (future: gRPC beside)   │  VOs, invariants) │   │ adapters/postgres (pgx)│
                             └───────────────────┘   │ adapters/memory (tests)│
                                                      └───────────────────────┘
```

- **`internal/lending/domain`** — `Book`, `Member`, `Loan` aggregates; `DueDate`
  and `MembershipStatus` value objects; typed UUID IDs; the `Borrow` domain
  service enforcing the invariants; typed domain errors; and the repository
  **ports** (interfaces).
- **`internal/lending/app`** — the `BorrowBook` use case. Delivery-agnostic (see
  [ADR 0002](docs/adr/0002-delivery-agnostic-application-layer.md)).
- **`internal/lending/adapters`** — `rest` (HTTP driving adapter), `postgres`
  (pgx driven adapter), `memory` (in-memory repos for fast tests / dev).
- **`internal/platform`** — `config` (env loading) and `migrate` (embedded
  golang-migrate runner).
- **`cmd/server`** — the composition root: manual wiring, migrate-on-boot,
  graceful shutdown.

See [`CONTEXT.md`](CONTEXT.md) for the domain glossary and
[`docs/adr/`](docs/adr) for the two architecture decisions.

## The one endpoint

`POST /loans` — a Member borrows a Book.

```json
{ "book_id": "<uuid>", "member_id": "<uuid>" }
```

Responses:

| Situation                     | Status | Domain error            |
| ----------------------------- | ------ | ----------------------- |
| Borrowed                      | 201    | —                       |
| Malformed body / bad UUID     | 400    | —                       |
| Book or member unknown        | 404    | `ErrNotFound`           |
| Book already on loan          | 409    | `ErrBookUnavailable`    |
| Member at loan limit (5)      | 422    | `ErrLoanLimitReached`   |
| Membership inactive           | 422    | `ErrMembershipInactive` |

`GET /healthz` returns `200 ok`.

## Requirements

- Go 1.26+
- Docker (for local Postgres and the integration tests)

## Quickstart

```sh
make db-up      # start Postgres via docker compose
make run        # migrate on boot, then serve on :8080
```

The server reads `DATABASE_URL` (required) and `HTTP_ADDR` (default `:8080`).

## Tests

```sh
make test              # fast: domain + app unit tests + HTTP httptest
make test-integration  # pgx adapter against a throwaway Postgres (testcontainers)
```

Run `make help` for all targets.

## Tooling / quality gate

Quality is enforced by [`pre-commit`](https://pre-commit.com): on every
`git commit` the full bar runs — `gofmt`, `go vet`, `golangci-lint`,
`go-arch-lint`, a `go mod tidy` check, and the fast test suite. Each check is
also a plain `make` target, so `pre-commit` is a thin adapter over them and you
can run any check by hand. See
[ADR 0003](docs/adr/0003-pre-commit-as-the-sole-quality-gate.md) for why this is
the sole gate.

### One-time setup

Install the `pre-commit` binary out-of-band (it is not a Go tool):

```sh
brew install pre-commit      # macOS
pipx install pre-commit      # or, cross-platform
```

Then register the git hooks:

```sh
make hooks-install   # runs `pre-commit install`
```

The Go linters (`golangci-lint`, `go-arch-lint`) are pinned and installed
automatically via `go install` the first time their `make` target runs — no
manual step needed.

### Running the checks

```sh
make pre-commit   # run the whole bar across all files (no commit needed)

make fmt          # gofmt -w (auto-format in place)
make vet          # go vet
make golangci     # golangci-lint only
make arch-lint    # go-arch-lint only (ADR 0001/0002 boundaries)
make lint         # both linters at once
make tidy-check   # fail if go.mod / go.sum are not tidy
make test         # fast suite
```

Every check runs over the whole module (`./...`), and the linter versions are
pinned in the `Makefile`, so a fresh clone reaches an identical green bar. The
checks work with or without `pre-commit` installed.

## Deliberate scope limits

This is an outline, not a finished service. Known simplifications:

- **Only the borrow slice is implemented.** Creating books/members and returning
  loans are out of scope; seed rows directly for now.
- **No unit-of-work / transaction across aggregates.** `BorrowBook` saves the
  loan, book, and member in sequence. A production version would wrap these in a
  single transaction so a mid-sequence failure cannot leave partial state.
- **No pagination, auth, or read endpoints.**

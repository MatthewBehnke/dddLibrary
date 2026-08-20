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
  (pgx driven adapter), `memory` (in-memory repos for fast tests / dev). The
  `rest` adapter is **spec-first**: [`api/openapi.yaml`](api/openapi.yaml) is the
  single source of truth, and `rest/openapi` holds the `oapi-codegen`-generated
  server interface and DTOs plus the `kin-openapi` request-validator middleware.
  The hand-written `rest` package *implements* that generated interface (see
  [ADR 0004](docs/adr/0004-spec-first-openapi-codegen-for-the-http-adapter.md)).
- **`internal/platform`** — `config` (env loading) and `migrate` (embedded
  golang-migrate runner).
- **`cmd/server`** — the composition root: manual wiring, migrate-on-boot,
  graceful shutdown.

See [`CONTEXT.md`](CONTEXT.md) for the domain glossary and
[`docs/adr/`](docs/adr) for the architecture decisions.

## The API

The full contract lives in [`api/openapi.yaml`](api/openapi.yaml). Only **Borrow**
(`POST /loans`) is wired to a use case today; the liveness probe answers
directly, and every other declared operation returns `501 Not Implemented` until
its slice is built, so what is built stays distinguishable from what is merely
promised.

| Method & path              | Operation      | Status                         |
| -------------------------- | -------------- | ------------------------------ |
| `POST /loans`              | Borrow a Book  | wired                          |
| `GET /healthz`             | Liveness       | wired (`200`)                  |
| `POST /loans/{id}/return`  | Return a Loan  | `501`                          |
| `GET /loans/{id}`          | Get a Loan     | `501`                          |
| `GET /loans`               | List Loans     | `501`                          |
| `POST /books`              | Create a Book  | `501`                          |
| `GET /books/{id}`          | Get a Book     | `501`                          |
| `GET /books`               | List Books     | `501`                          |
| `POST /members`            | Create a Member| `501`                          |
| `GET /members/{id}`        | Get a Member   | `501`                          |
| `GET /members`             | List Members   | `501`                          |

Requests are validated against the spec by a `kin-openapi` middleware before any
business logic runs, and every non-2xx response — validation and domain alike —
shares one `{ "error": "..." }` shape. Collection endpoints declare a `limit` and
an opaque cursor and return a `{ "data": [...], "next_cursor": ... }` envelope.

### Borrow — `POST /loans`

```json
{ "book_id": "<uuid>", "member_id": "<uuid>" }
```

Success is `201` with the new Loan's identifier and Due Date. Other responses:

| Situation                     | Status | Domain error            |
| ----------------------------- | ------ | ----------------------- |
| Borrowed                      | 201    | —                       |
| Malformed body / bad UUID     | 400    | — (rejected by validator) |
| Book or member unknown        | 404    | `ErrNotFound`           |
| Book already on loan          | 409    | `ErrBookUnavailable`    |
| Member at loan limit (5)      | 422    | `ErrLoanLimitReached`   |
| Membership inactive           | 422    | `ErrMembershipInactive` |

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
`go-arch-lint`, a `go mod tidy` check, a spec/codegen drift check
(`generate-check`), and the fast test suite. Each check is also a plain `make`
target, so `pre-commit` is a thin adapter over them and you can run any check by
hand. See
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
make generate     # regenerate the OpenAPI server code from api/openapi.yaml
make generate-check # fail if the committed generated code drifts from the spec
make test         # fast suite
```

Every check runs over the whole module (`./...`), and the linter and
`oapi-codegen` versions are pinned in the `Makefile`, so a fresh clone reaches an
identical green bar and reproduces identical generated output. The generated code
is committed, so a fresh clone builds without the code generator installed. The
checks work with or without `pre-commit` installed.

## Deliberate scope limits

This is an outline, not a finished service. Known simplifications:

- **Only the borrow slice is implemented.** Every other operation in the spec
  (create Book/Member, return a Loan, the reads and collections) is declared but
  returns `501`; seed rows directly for now.
- **No unit-of-work / transaction across aggregates.** `BorrowBook` saves the
  loan, book, and member in sequence. A production version would wrap these in a
  single transaction so a mid-sequence failure cannot leave partial state.
- **Pagination is contract-only.** Collections declare the `limit`/cursor
  envelope but are not implemented; auth and rate limiting are out of scope.

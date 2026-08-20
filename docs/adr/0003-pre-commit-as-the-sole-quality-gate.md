# pre-commit as the sole quality gate

Quality is enforced by the [`pre-commit`](https://pre-commit.com) framework and
nothing else — there is no CI yet. On every `git commit` the full bar runs:
`gofmt`, `go vet`, `golangci-lint`, `go-arch-lint`, a `go mod tidy` check, and
the fast test suite. `make build` is excluded (the test compile already covers
it) and the Docker-dependent integration tests are excluded (committing must not
require a running daemon). CI is a deliberate follow-up, not a gap we forgot.

Every hook is a thin shell-out to a `make` target: `.pre-commit-config.yaml` is
a single `repo: local` block whose hooks all use `language: system` with
`entry: make <target>`. We rejected pulling in upstream hook repos (e.g.
`pre-commit/pre-commit-hooks`) so there is exactly one source of truth — every
check is runnable with or without `pre-commit` installed (`make golangci`,
`make test`, …), and the framework is a thin adapter over the `make` targets
rather than a second, divergent gate. The checks are unusual in that each runs
over the whole module (`./...`), not just staged files, because that is how the
Go tools work and it stops a staged change from silently breaking an unstaged
package. All hooks run at the single `pre-commit` stage; there is no pre-push
stage. The linters are pinned (`golangci-lint v2.13.0`, `go-arch-lint v1.17.0`)
so the bar does not drift.

We adopted `go-arch-lint` specifically to make [ADR 0001](0001-hexagonal-ports-and-adapters.md)
and [ADR 0002](0002-delivery-agnostic-application-layer.md) build-failing rather
than review-enforced: `.go-arch-lint.yml` encodes the component graph and the
inward-only dependency rules, so a domain import of an adapter, or a pgx import
inside the app layer, fails the commit. One sharp edge is worth recording:
`go-arch-lint` governs internal components and external (go.mod) vendors only —
it *cannot* restrict standard-library imports. ADR 0002's rule that the app
layer must not import `net/http` (stdlib) is therefore enforced by
`golangci-lint`'s `depguard` linter instead, while the pgx half of that rule
stays in the archfile's vendor allowlist. `deepScan` (AST-level usage analysis,
on by default) is turned off because at the composition root it misreads
legitimate dependency injection — wiring a Postgres repository into the app use
case — as an adapter-to-app dependency; import-level checks enforce the ADRs
without that false positive.

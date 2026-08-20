# Spec-first OpenAPI codegen for the HTTP adapter

The HTTP delivery layer is spec-first: a hand-authored `api/openapi.yaml` is the
single source of truth for the lending API, and Go is generated from it with
[`oapi-codegen`](https://github.com/oapi-codegen/oapi-codegen) in **strict-server**
mode targeting the standard-library `net/http` router. The generated code
defines a `StrictServerInterface` plus request/response types; the hand-written
`adapters/rest` package *implements* that interface and translates its DTOs to
and from the `app` use cases. This replaces the previously hand-written handler,
router, and request-parsing code. The core is unaffected: this decision lives
entirely in the driving adapter and does not amend
[ADR 0002](0002-delivery-agnostic-application-layer.md) — the `app` layer still
accepts plain commands and returns plain results with no transport types.

We chose `oapi-codegen` (std-http, strict-server) over `ogen` and `go-swagger`
because it maps onto the existing stdlib `ServeMux` with the least churn while
giving exactly the "adapter implements a generated interface" shape we want.
`ogen` is more complete (its own router, typed everything, built-in validation)
but imposes its own runtime and router, which is disproportionate for this
service; `go-swagger` is heavyweight and dated. Request validation is done by a
[`kin-openapi`](https://github.com/getkin/kin-openapi) validator middleware
loaded from the spec, so the contract is enforced rather than merely documented;
the manual `json.Decode` + `domain.ParseBookID`/`ParseMemberID` guards are gone.
A small custom error handler reshapes validator failures into the same
`{"error": "..."}` body as domain errors, so every response honors the spec's
shared `Error` schema. Domain errors keep their existing mapping to status codes
(`statusForError`: 404/409/422/500). Spec formats map idiomatically —
`format: uuid` to `uuid.UUID`, `format: date-time` to `time.Time` — and the
adapter converts to domain typed IDs with `domain.BookID(...)` conversions.

The spec covers the full intended API surface, not just the implemented borrow
slice: create and get-by-id for Book, Member, and Loan; cursor-paginated
collections (`?limit=` plus an opaque cursor, responses wrapped as
`{data, next_cursor}`); `POST /loans/{id}/return` for returning a loan; and
`GET /healthz`. Only `POST /loans` (the `BorrowBook` use case) is wired; every
other generated operation returns `501 Not Implemented` until its slice is
built. Publishing the whole contract up front — including the pagination scheme,
which is expensive to change once clients depend on it — is the point: the cost
now is only the spec and stub methods the compiler forces us to acknowledge, and
`501` keeps built and unbuilt operations honestly distinguishable rather than
faking functionality.

The generated code is committed and kept truthful the same way `go.mod` is: a
`//go:generate` directive and a `make generate` target regenerate it, and a
`make generate-check` (regenerate, then `git diff --exit-code`, mirroring
`tidy-check`) runs in the `pre-commit` bar so the spec and the generated code
cannot silently diverge (see [ADR 0003](0003-pre-commit-as-the-sole-quality-gate.md)).
A fresh clone therefore builds without the codegen tool installed, and the
`oapi-codegen` version is pinned in the `Makefile` alongside the linters so the
output cannot drift.

The generated code lives in its own leaf package,
`internal/lending/adapters/rest/openapi`, which is the only place that touches
the new vendors — `oapi-codegen`'s runtime, `kin-openapi`, and (via typed DTO
fields) `uuid`. `.go-arch-lint.yml` gains a matching `adapter-rest-openapi`
component whose `canUse` allowlists exactly those three vendors; the hand-written
`adapter-rest` component keeps its zero-vendor allowlist and merely
`mayDependOn` the generated package (the `domain.BookID(field)` conversions do
not name `uuid`, so `rest` never imports it). This isolates the entire
generated-code dependency footprint behind one boundary and keeps the hexagonal
seam crisp. The alternative — hand-maintaining handlers, request parsing, and an
out-of-band API document — is simpler for one endpoint but drifts from the
contract exactly as the API grows, which is the failure a spec-first generator
exists to prevent.

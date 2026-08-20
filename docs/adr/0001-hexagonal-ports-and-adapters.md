# Hexagonal architecture (ports & adapters)

The codebase is structured as ports & adapters under `internal/lending`: the
`domain` package holds entities, value objects, and the repository interfaces
(ports) it owns; the `app` package holds delivery-agnostic use cases; and
`adapters/*` packages (rest, postgres, memory) depend inward on those ports.
Dependencies always point toward the domain — the domain imports no adapter, no
framework, and no SQL or HTTP types.

We chose this over a flat/standard layout so the lending rules live in one
dependency-free core that is trivially unit-testable (the `memory` adapter
stands in for Postgres in fast tests), and so infrastructure choices are swap-in
details rather than pervasive assumptions. The cost is more packages and
constructor wiring than a naive layout; that boilerplate is the price of the
dependency rule.

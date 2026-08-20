# Delivery-agnostic application layer (gRPC-ready)

The `app` use cases accept plain commands (typed domain IDs) and return plain
results — they never touch `net/http`, `context`-carried request data, or any
transport type. HTTP lives entirely in the `adapters/rest` driving adapter,
which parses requests, calls the use case, and maps typed domain errors to
status codes.

We decided this now, while HTTP is the only delivery mechanism, specifically so
a future gRPC adapter can sit beside `rest` and invoke the same use cases with
no change to the core. The alternative — letting handlers hold business logic or
passing `http.Request` inward — is simpler for a single transport but would have
to be unpicked the moment a second one arrives. Keeping the seam from day one is
cheap; retrofitting it is not.

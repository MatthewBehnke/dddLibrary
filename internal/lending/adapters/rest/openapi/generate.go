// Package openapi holds the spec-generated HTTP server contract for the lending
// API together with the kin-openapi request-validation middleware. It is the
// single boundary that touches the transport-codegen dependencies (the
// oapi-codegen runtime, kin-openapi, and — via the generated uuid-typed DTO
// fields — google/uuid); the hand-written adapters/rest package implements the
// generated StrictServerInterface and never imports those libraries directly.
//
// The Go in openapi.gen.go is generated from api/openapi.yaml and committed, so
// a fresh clone builds without the generator. Run `make generate` to refresh it.
package openapi

//go:generate oapi-codegen -config config.yaml ../../../../../api/openapi.yaml

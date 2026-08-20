package openapi

import (
	"net/http"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

// ErrorHandler renders a request-validation failure. The rest adapter supplies
// one so that validator failures share the adapter's {"error": ...} body; the
// message text remains the validator's.
type ErrorHandler func(w http.ResponseWriter, r *http.Request, err error)

// NewValidator builds middleware that validates every incoming request against
// the embedded OpenAPI spec before it reaches a handler. Requests that violate
// the contract — malformed bodies, missing required fields, badly-formatted
// identifiers — are rejected via onError (a 400) and never reach the adapter.
// Requests to paths outside the spec fall through to next, which 404s.
//
// This is the sole seam where kin-openapi is used; the rest adapter only sees
// the returned middleware and its own ErrorHandler.
func NewValidator(onError ErrorHandler) (func(http.Handler) http.Handler, error) {
	spec, err := GetSwagger()
	if err != nil {
		return nil, err
	}
	// The router matches on path only; clearing servers avoids host/base-path
	// matching against the (documentation-only) servers list.
	spec.Servers = nil

	router, err := gorillamux.NewRouter(spec)
	if err != nil {
		return nil, err
	}

	options := &openapi3filter.Options{AuthenticationFunc: openapi3filter.NoopAuthenticationFunc}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			route, pathParams, err := router.FindRoute(r)
			if err != nil {
				// Not a spec'd route (unknown path or method); let the mux
				// produce its own 404/405.
				next.ServeHTTP(w, r)
				return
			}

			input := &openapi3filter.RequestValidationInput{
				Request:    r,
				PathParams: pathParams,
				Route:      route,
				Options:    options,
			}
			if err := openapi3filter.ValidateRequest(r.Context(), input); err != nil {
				onError(w, r, err)
				return
			}

			next.ServeHTTP(w, r)
		})
	}, nil
}

// Package rest is the HTTP driving adapter. It implements the spec-generated
// StrictServerInterface (see the openapi subpackage): it translates generated
// request DTOs into application use-case calls and maps results and domain
// errors back to HTTP. The API's shape is defined once in api/openapi.yaml and
// the server follows from it; this package holds only the hand-written glue.
//
// Only Borrow (POST /loans) and the liveness probe are wired; every other
// declared operation returns 501 until its slice is built. A future gRPC adapter
// would sit beside this package, calling the same use case.
package rest

import (
	"context"
	"errors"
	"net/http"

	"github.com/matthewbehnke/dddlibrary/internal/lending/adapters/rest/openapi"
	"github.com/matthewbehnke/dddlibrary/internal/lending/app"
	"github.com/matthewbehnke/dddlibrary/internal/lending/domain"
)

// errNotImplemented marks an operation that is declared in the spec but not yet
// wired to a use case. The response error handler maps it to 501.
var errNotImplemented = errors.New("not implemented")

// Server implements the spec-generated StrictServerInterface. Only BorrowBook
// and Healthz are wired; the remaining operations delegate to the 501 path.
type Server struct {
	borrow *app.BorrowBook
}

var _ openapi.StrictServerInterface = (*Server)(nil)

// NewHandler assembles the full HTTP handler for the lending adapter: the
// spec-generated strict server (carrying the domain-error and 501 status
// mapping) wrapped by the kin-openapi request validator, whose failures are
// reshaped into the adapter's shared {"error": ...} body. This is the single
// wiring seam: both cmd/server and the HTTP tests build the handler here, so
// the validator middleware and error handling are always in the tested path.
func NewHandler(borrow *app.BorrowBook) (http.Handler, error) {
	strict := openapi.NewStrictHandlerWithOptions(&Server{borrow: borrow}, nil, openapi.StrictHTTPServerOptions{
		// Strict binding failures that slipped past the validator (e.g. an
		// undecodable body) share the adapter's error shape as a 400.
		RequestErrorHandlerFunc: func(w http.ResponseWriter, _ *http.Request, err error) {
			writeError(w, http.StatusBadRequest, err)
		},
		// A handler returning an error is mapped here: 501 for unimplemented
		// operations, otherwise the domain-error status mapping.
		ResponseErrorHandlerFunc: func(w http.ResponseWriter, _ *http.Request, err error) {
			if errors.Is(err, errNotImplemented) {
				writeError(w, http.StatusNotImplemented, err)
				return
			}
			writeError(w, statusForError(err), err)
		},
	})

	handler := openapi.HandlerFromMux(strict, http.NewServeMux())

	validate, err := openapi.NewValidator(func(w http.ResponseWriter, _ *http.Request, err error) {
		writeError(w, http.StatusBadRequest, err)
	})
	if err != nil {
		return nil, err
	}
	return validate(handler), nil
}

// BorrowBook handles POST /loans: a Member borrows a Book. It is the only
// operation wired to a use case. Domain errors are returned unchanged for the
// response error handler to map (404/409/422/500).
func (s *Server) BorrowBook(ctx context.Context, request openapi.BorrowBookRequestObject) (openapi.BorrowBookResponseObject, error) {
	res, err := s.borrow.Execute(ctx, app.BorrowBookCommand{
		BookID:   domain.BookID(request.Body.BookId),
		MemberID: domain.MemberID(request.Body.MemberId),
	})
	if err != nil {
		return nil, err
	}
	return openapi.BorrowBook201JSONResponse{
		LoanId:  res.LoanID.String(),
		DueDate: res.DueDate,
	}, nil
}

// Healthz handles GET /healthz: liveness, requiring no use case.
func (s *Server) Healthz(context.Context, openapi.HealthzRequestObject) (openapi.HealthzResponseObject, error) {
	return openapi.Healthz200JSONResponse{Status: "ok"}, nil
}

// The operations below are declared in the spec but not yet wired to a use
// case; each returns 501 until its slice is built.

func (s *Server) CreateBook(context.Context, openapi.CreateBookRequestObject) (openapi.CreateBookResponseObject, error) {
	return nil, errNotImplemented
}

func (s *Server) GetBook(context.Context, openapi.GetBookRequestObject) (openapi.GetBookResponseObject, error) {
	return nil, errNotImplemented
}

func (s *Server) ListBooks(context.Context, openapi.ListBooksRequestObject) (openapi.ListBooksResponseObject, error) {
	return nil, errNotImplemented
}

func (s *Server) CreateMember(context.Context, openapi.CreateMemberRequestObject) (openapi.CreateMemberResponseObject, error) {
	return nil, errNotImplemented
}

func (s *Server) GetMember(context.Context, openapi.GetMemberRequestObject) (openapi.GetMemberResponseObject, error) {
	return nil, errNotImplemented
}

func (s *Server) ListMembers(context.Context, openapi.ListMembersRequestObject) (openapi.ListMembersResponseObject, error) {
	return nil, errNotImplemented
}

func (s *Server) GetLoan(context.Context, openapi.GetLoanRequestObject) (openapi.GetLoanResponseObject, error) {
	return nil, errNotImplemented
}

func (s *Server) ListLoans(context.Context, openapi.ListLoansRequestObject) (openapi.ListLoansResponseObject, error) {
	return nil, errNotImplemented
}

func (s *Server) ReturnLoan(context.Context, openapi.ReturnLoanRequestObject) (openapi.ReturnLoanResponseObject, error) {
	return nil, errNotImplemented
}

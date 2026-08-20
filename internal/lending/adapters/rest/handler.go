// Package rest is the HTTP driving adapter. It translates HTTP requests into
// application use-case calls and domain errors into HTTP status codes. It is
// the only place that imports net/http; the core knows nothing about it. A
// future gRPC adapter would sit beside this package, calling the same use case.
package rest

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/matthewbehnke/dddlibrary/internal/lending/app"
	"github.com/matthewbehnke/dddlibrary/internal/lending/domain"
)

// LoanHandler serves loan-related endpoints backed by the BorrowBook use case.
type LoanHandler struct {
	borrow *app.BorrowBook
}

// NewLoanHandler builds a LoanHandler around the BorrowBook use case.
func NewLoanHandler(borrow *app.BorrowBook) *LoanHandler {
	return &LoanHandler{borrow: borrow}
}

// borrowRequest is the POST /loans request body.
type borrowRequest struct {
	BookID   string `json:"book_id"`
	MemberID string `json:"member_id"`
}

// borrowResponse is the POST /loans success body.
type borrowResponse struct {
	LoanID  string `json:"loan_id"`
	DueDate string `json:"due_date"`
}

// Borrow handles POST /loans: it borrows a book for a member.
func (h *LoanHandler) Borrow(w http.ResponseWriter, r *http.Request) {
	var req borrowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid JSON body"))
		return
	}

	bookID, err := domain.ParseBookID(req.BookID)
	if err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid book_id"))
		return
	}
	memberID, err := domain.ParseMemberID(req.MemberID)
	if err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid member_id"))
		return
	}

	res, err := h.borrow.Execute(r.Context(), app.BorrowBookCommand{BookID: bookID, MemberID: memberID})
	if err != nil {
		writeError(w, statusForError(err), err)
		return
	}

	writeJSON(w, http.StatusCreated, borrowResponse{
		LoanID:  res.LoanID.String(),
		DueDate: res.DueDate.UTC().Format(time.RFC3339),
	})
}

package rest_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/matthewbehnke/dddlibrary/internal/lending/adapters/memory"
	"github.com/matthewbehnke/dddlibrary/internal/lending/adapters/rest"
	"github.com/matthewbehnke/dddlibrary/internal/lending/app"
	"github.com/matthewbehnke/dddlibrary/internal/lending/domain"
)

type harness struct {
	server  *httptest.Server
	books   *memory.BookRepository
	members *memory.MemberRepository
}

func newHarness(t *testing.T) harness {
	t.Helper()
	books := memory.NewBookRepository()
	members := memory.NewMemberRepository()
	loans := memory.NewLoanRepository()
	uc := app.NewBorrowBook(books, members, loans, func() time.Time {
		return time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	})
	handler, err := rest.NewHandler(uc)
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return harness{server: srv, books: books, members: members}
}

func (h harness) seedBook(t *testing.T, available bool) domain.BookID {
	t.Helper()
	b := domain.ReconstituteBook(domain.NewBookID(), "SICP", available)
	if err := h.books.Save(context.Background(), b); err != nil {
		t.Fatal(err)
	}
	return b.ID()
}

func (h harness) seedMember(t *testing.T, status domain.MembershipStatus, active int) domain.MemberID {
	t.Helper()
	m := domain.ReconstituteMember(domain.NewMemberID(), "Ada", status, active)
	if err := h.members.Save(context.Background(), m); err != nil {
		t.Fatal(err)
	}
	return m.ID()
}

func (h harness) postLoan(t *testing.T, body string) *http.Response {
	t.Helper()
	resp, err := http.Post(h.server.URL+"/loans", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })
	return resp
}

func jsonBody(bookID, memberID string) string {
	b, _ := json.Marshal(map[string]string{"book_id": bookID, "member_id": memberID})
	return string(b)
}

func TestBorrow_created(t *testing.T) {
	h := newHarness(t)
	bookID := h.seedBook(t, true)
	memberID := h.seedMember(t, domain.MembershipActive, 0)

	resp := h.postLoan(t, jsonBody(bookID.String(), memberID.String()))

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
	var got struct {
		LoanID  string `json:"loan_id"`
		DueDate string `json:"due_date"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.LoanID == "" {
		t.Error("expected a loan_id")
	}
	if want := "2026-06-15T00:00:00Z"; got.DueDate != want {
		t.Errorf("due_date = %q, want %q", got.DueDate, want)
	}
}

func TestBorrow_statusMapping(t *testing.T) {
	tests := []struct {
		name     string
		body     func(h harness, t *testing.T) string
		wantCode int
	}{
		{
			name:     "invalid json",
			body:     func(harness, *testing.T) string { return "{not json" },
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "invalid book id",
			body:     func(harness, *testing.T) string { return jsonBody("not-a-uuid", domain.NewMemberID().String()) },
			wantCode: http.StatusBadRequest,
		},
		{
			name: "book not found",
			body: func(h harness, t *testing.T) string {
				m := h.seedMember(t, domain.MembershipActive, 0)
				return jsonBody(domain.NewBookID().String(), m.String())
			},
			wantCode: http.StatusNotFound,
		},
		{
			name: "book unavailable",
			body: func(h harness, t *testing.T) string {
				b := h.seedBook(t, false)
				m := h.seedMember(t, domain.MembershipActive, 0)
				return jsonBody(b.String(), m.String())
			},
			wantCode: http.StatusConflict,
		},
		{
			name: "loan limit reached",
			body: func(h harness, t *testing.T) string {
				b := h.seedBook(t, true)
				m := h.seedMember(t, domain.MembershipActive, domain.MaxActiveLoans)
				return jsonBody(b.String(), m.String())
			},
			wantCode: http.StatusUnprocessableEntity,
		},
		{
			name: "membership inactive",
			body: func(h harness, t *testing.T) string {
				b := h.seedBook(t, true)
				m := h.seedMember(t, domain.MembershipInactive, 0)
				return jsonBody(b.String(), m.String())
			},
			wantCode: http.StatusUnprocessableEntity,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := newHarness(t)
			resp := h.postLoan(t, tc.body(h, t))
			if resp.StatusCode != tc.wantCode {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tc.wantCode)
			}
		})
	}
}

// TestBorrow_validationErrorShape proves a spec-violating request is rejected
// with 400 before any business logic runs, and that the validator's failure is
// reshaped into the adapter's shared {"error": ...} body.
func TestBorrow_validationErrorShape(t *testing.T) {
	h := newHarness(t)

	resp := h.postLoan(t, jsonBody("not-a-uuid", domain.NewMemberID().String()))

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
	var got struct {
		Error string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Error == "" {
		t.Error("expected a non-empty error message in the shared error shape")
	}
}

// TestUnimplementedOperation_returns501 proves a declared-but-unwired operation
// (POST /books) returns 501 rather than fabricating a result, and honors the
// shared error shape.
func TestUnimplementedOperation_returns501(t *testing.T) {
	h := newHarness(t)

	resp, err := http.Post(h.server.URL+"/books", "application/json", strings.NewReader(`{"title":"SICP"}`))
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })

	if resp.StatusCode != http.StatusNotImplemented {
		t.Fatalf("status = %d, want 501", resp.StatusCode)
	}
	var got struct {
		Error string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Error == "" {
		t.Error("expected a non-empty error message in the shared error shape")
	}
}

// TestHealthz_ok proves the liveness probe answers 200 through the real wiring
// (it is modeled in-spec but needs no use case, so it is not a 501 stub).
func TestHealthz_ok(t *testing.T) {
	h := newHarness(t)

	resp, err := http.Get(h.server.URL + "/healthz")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	var got struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Status != "ok" {
		t.Errorf("status = %q, want %q", got.Status, "ok")
	}
}

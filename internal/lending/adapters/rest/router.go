package rest

import "net/http"

// NewRouter builds the HTTP routing table for the lending context using the
// stdlib ServeMux method+path patterns (Go 1.22+).
func NewRouter(loans *LoanHandler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /loans", loans.Borrow)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	return mux
}

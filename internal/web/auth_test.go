package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ben/ikite-go/internal/config"
)

func TestSettingsRequiresPass(t *testing.T) {
	cfg := &config.Config{SettingsPass: "secret-guid"}
	s, err := New(cfg, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	h := s.Handler()

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/settings", nil))
	if rr.Code != http.StatusNotFound {
		t.Fatalf("no pass: status %d", rr.Code)
	}

	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/settings?pass=wrong", nil))
	if rr.Code != http.StatusNotFound {
		t.Fatalf("wrong pass: status %d", rr.Code)
	}
}

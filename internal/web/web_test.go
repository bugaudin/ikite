package web

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ben/ikite-go/internal/config"
)

func TestTemplatesParse(t *testing.T) {
	cfg := &config.Config{}
	s, err := New(cfg, nil, slog.Default())
	if err != nil {
		t.Fatal(err)
	}
	if s.tmpl.Lookup("index.html") == nil {
		t.Fatal("index.html missing")
	}
	if s.tmpl.Lookup("graph.html") == nil {
		t.Fatal("graph.html missing")
	}
	if s.tmpl.Lookup("settings.html") == nil {
		t.Fatal("settings.html missing")
	}
	if s.tmpl.Lookup("camera.html") == nil {
		t.Fatal("camera.html missing")
	}
	if s.tmpl.Lookup("prediction.html") == nil {
		t.Fatal("prediction.html missing")
	}
}

func TestHealthz(t *testing.T) {
	cfg := &config.Config{}
	s, err := New(cfg, nil, slog.Default())
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d", rr.Code)
	}
}

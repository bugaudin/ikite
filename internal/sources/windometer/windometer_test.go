package windometer

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ben/ikite-go/internal/begetproxy"
)

func TestFetch(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("slugs") != khSlug {
			t.Fatalf("slugs: %s", r.URL.Query().Get("slugs"))
		}
		_, _ = w.Write([]byte(`{"ok":true,"results":{"pick-up-surf":{"Angle":250,"Speed":11,"Gust":12,"recorded_at":1784190067,"stale":false}}}`))
	}))
	defer upstream.Close()

	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true,"results":{"pick-up-surf":{"Angle":250,"Speed":11,"Gust":12,"recorded_at":1784190067,"stale":false}}}`))
	}))
	defer proxy.Close()

	client := New(begetproxy.New(proxy.URL), upstream.URL+"?slugs="+khSlug)
	now := time.Unix(1784190067, 0).In(time.UTC)
	reading, _, err := client.Fetch(now)
	if err != nil {
		t.Fatal(err)
	}
	if reading.Location != "kh" || reading.Wind != 11 || reading.Gust != 12 || reading.WindDir != 250 {
		t.Fatalf("reading: %+v", reading)
	}
}

func TestFetchStale(t *testing.T) {
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"ok":true,"results":{"pick-up-surf":{"Angle":90,"Speed":8,"Gust":10,"stale":true}}}`))
	}))
	defer proxy.Close()

	client := New(begetproxy.New(proxy.URL), "http://example/live?slugs="+khSlug)
	reading, _, err := client.Fetch(time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if reading.Wind != 0 || reading.Gust != 0 || reading.WindDir != 0 {
		t.Fatalf("stale should zero reading: %+v", reading)
	}
}

package begetproxy

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestJSONKeys(t *testing.T) {
	b, err := json.Marshal(Request{
		URL:     "https://example.com",
		Method:  "GET",
		Headers: map[string]string{"Accept": "*/*"},
	})
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"url", "method", "headers"} {
		if _, ok := m[key]; !ok {
			t.Fatalf("missing json key %q in %s", key, string(b))
		}
	}
	if _, ok := m["URL"]; ok {
		t.Fatalf("capitalized URL in %s", string(b))
	}
}

func TestClientDo(t *testing.T) {
	const testSecret = "test-secret"

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Test") != "1" {
			t.Fatalf("missing header: %+v", r.Header)
		}
		_, _ = io.WriteString(w, `{"ok":true}`)
	}))
	defer upstream.Close()

	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method: %s", r.Method)
		}
		if r.Header.Get("X-Proxy-Secret") != testSecret {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		body, _ := io.ReadAll(r.Body)
		var req Request
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatal(err)
		}
		if req.URL != upstream.URL {
			t.Fatalf("url: %s", req.URL)
		}
		upReq, err := http.NewRequest(http.MethodGet, req.URL, nil)
		if err != nil {
			t.Fatal(err)
		}
		for k, v := range req.Headers {
			upReq.Header.Set(k, v)
		}
		resp, err := http.DefaultClient.Do(upReq)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		out, _ := io.ReadAll(resp.Body)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(out)
	}))
	defer proxy.Close()

	client := New(proxy.URL, testSecret)
	out, err := client.Get(upstream.URL, map[string]string{"X-Test": "1"})
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != `{"ok":true}` {
		t.Fatalf("body: %s", out)
	}
}

package begetproxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client forwards HTTP requests through a generic Beget proxy_post.php (POST JSON).
// Set URL to "direct" to call upstream hosts without the proxy (emergency / local only).
type Client struct {
	HTTP   *http.Client
	URL    string
	Secret string
}

func New(url, secret string) *Client {
	return &Client{
		HTTP:   &http.Client{Timeout: 90 * time.Second},
		URL:    strings.TrimRight(url, "/"),
		Secret: secret,
	}
}

// Request describes an upstream HTTP call to execute via the proxy.
type Request struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body,omitempty"`
}

func (c *Client) Do(req Request) ([]byte, error) {
	if c == nil || c.URL == "" {
		return nil, fmt.Errorf("beget proxy URL not configured")
	}
	if req.URL == "" {
		return nil, fmt.Errorf("beget proxy: missing target url")
	}
	if req.Method == "" {
		req.Method = http.MethodGet
	}
	if req.Headers == nil {
		req.Headers = map[string]string{}
	}

	if c.URL == "direct" {
		return c.doDirect(req)
	}
	if c.Secret == "" {
		return nil, fmt.Errorf("beget proxy secret not configured")
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest(http.MethodPost, c.URL, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Proxy-Secret", c.Secret)
	httpReq.Header.Set("Accept", "application/json, text/plain, */*")
	httpReq.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	if origin := proxyOrigin(c.URL); origin != "" {
		httpReq.Header.Set("Origin", origin)
		httpReq.Header.Set("Referer", origin+"/")
	}

	resp, err := c.HTTP.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return body, fmt.Errorf("beget proxy upstream status %d", resp.StatusCode)
	}
	if len(body) == 0 {
		return nil, fmt.Errorf("beget proxy: empty response")
	}
	return body, nil
}

func (c *Client) doDirect(req Request) ([]byte, error) {
	var bodyReader io.Reader
	if req.Body != "" {
		bodyReader = strings.NewReader(req.Body)
	}
	httpReq, err := http.NewRequest(req.Method, req.URL, bodyReader)
	if err != nil {
		return nil, err
	}
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	resp, err := c.HTTP.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return body, fmt.Errorf("upstream status %d", resp.StatusCode)
	}
	if len(body) == 0 {
		return nil, fmt.Errorf("upstream: empty response")
	}
	return body, nil
}

func (c *Client) Get(url string, headers map[string]string) ([]byte, error) {
	return c.Do(Request{URL: url, Method: http.MethodGet, Headers: headers})
}

func proxyOrigin(proxyURL string) string {
	u, err := url.Parse(proxyURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return ""
	}
	return u.Scheme + "://" + u.Host
}

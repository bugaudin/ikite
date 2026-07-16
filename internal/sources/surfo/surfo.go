package surfo

import (
	"encoding/json"
	"fmt"

	"github.com/ben/ikite-go/internal/begetproxy"
)

const userAgent = "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/50.0.2661.102 Safari/537.36"

type Client struct {
	Proxy       *begetproxy.Client
	UpstreamURL string
}

func New(proxy *begetproxy.Client, upstreamURL string) *Client {
	return &Client{
		Proxy:       proxy,
		UpstreamURL: upstreamURL,
	}
}

type liveJSON struct {
	Report string `json:"report"`
}

func (c *Client) FetchReport() (string, error) {
	body, err := c.Proxy.Get(c.UpstreamURL, map[string]string{
		"user-agent": userAgent,
		"accept":     "application/json",
	})
	if err != nil {
		return "", err
	}

	var parsed liveJSON
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", err
	}
	if parsed.Report == "" {
		return "", fmt.Errorf("no report in response")
	}
	return parsed.Report, nil
}

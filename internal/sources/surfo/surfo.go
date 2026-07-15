package surfo

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const userAgent = "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/50.0.2661.102 Safari/537.36"

type Client struct {
	HTTP *http.Client
	URL  string
}

func New(url string) *Client {
	return &Client{
		HTTP: &http.Client{Timeout: 30 * time.Second},
		URL:  url,
	}
}

type liveJSON struct {
	Report string `json:"report"`
}

func (c *Client) FetchReport() (string, error) {
	req, err := http.NewRequest(http.MethodGet, c.URL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("surfo live status %d", resp.StatusCode)
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

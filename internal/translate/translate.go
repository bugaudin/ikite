package translate

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	HTTP *http.Client
}

func New() *Client {
	return &Client{
		HTTP: &http.Client{Timeout: 30 * time.Second},
	}
}

type mmResp struct {
	ResponseData struct {
		TranslatedText string `json:"translatedText"`
	} `json:"responseData"`
}

func (c *Client) HebrewToEnglish(text string) (string, error) {
	q := url.Values{}
	q.Set("q", text)
	q.Set("langpair", "he|en")
	endpoint := "https://api.mymemory.translated.net/get?" + q.Encode()

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; ikite-go/1.0)")

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
		return "", fmt.Errorf("mymemory status %d", resp.StatusCode)
	}

	var parsed mmResp
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", err
	}
	if parsed.ResponseData.TranslatedText == "" {
		return "", fmt.Errorf("empty translation")
	}
	return parsed.ResponseData.TranslatedText, nil
}

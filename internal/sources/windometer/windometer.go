package windometer

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ben/ikite-go/internal/models"
)

const userAgent = "Mozilla/5.0 (compatible; ikite-go/1.0)"

type Client struct {
	HTTP *http.Client
}

func New() *Client {
	return &Client{
		HTTP: &http.Client{Timeout: 30 * time.Second},
	}
}

type historyResp struct {
	Records []record `json:"Records"`
}

type record struct {
	Speed float64 `json:"Speed"`
	Gust  float64 `json:"Gust"`
	Angle float64 `json:"Angle"`
}

type liveResp struct {
	Speed float64 `json:"Speed"`
	Gust  float64 `json:"Gust"`
	Angle float64 `json:"Angle"`
}

func (c *Client) FetchHistory(now time.Time) (*models.WindReading, string, error) {
	url := fmt.Sprintf("https://www.windometer.info/updates/history.json?v=%d", now.UnixMilli())
	body, err := c.get(url)
	if err != nil {
		return nil, "", err
	}

	var parsed historyResp
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, string(body), err
	}
	if len(parsed.Records) == 0 {
		return nil, string(body), fmt.Errorf("windometer: empty records")
	}

	cur := parsed.Records[0]
	wind, gust, dir := cur.Speed, cur.Gust, cur.Angle

	// Stuck-sensor detection: two identical consecutive readings.
	if len(parsed.Records) > 1 {
		prev := parsed.Records[1]
		if cur.Speed == prev.Speed && cur.Gust == prev.Gust && cur.Angle == prev.Angle {
			wind, gust, dir = 0, 0, 0
		}
	}

	return &models.WindReading{
		Period:   now,
		Location: "kh",
		Wind:     wind,
		Gust:     gust,
		WindDir:  dir,
	}, string(body), nil
}

func (c *Client) FetchLive() (*models.WindReading, error) {
	url := fmt.Sprintf("https://www.windometer.info/updates/live.json?v=%d", time.Now().UnixMilli())
	body, err := c.get(url)
	if err != nil {
		return nil, err
	}
	var parsed liveResp
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	return &models.WindReading{
		Wind:    parsed.Speed,
		Gust:    parsed.Gust,
		WindDir: parsed.Angle,
	}, nil
}

func (c *Client) get(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return body, fmt.Errorf("windometer status %d", resp.StatusCode)
	}
	return body, nil
}

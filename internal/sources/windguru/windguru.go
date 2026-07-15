package windguru

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ben/ikite-go/internal/models"
)

const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/111.0.0.0 Safari/537.36"

type Client struct {
	HTTP       *http.Client
	StationURL string // sprintf template with %d, or proxy base URL
}

func New(stationURLTemplate string) *Client {
	return &Client{
		HTTP:       &http.Client{Timeout: 60 * time.Second},
		StationURL: stationURLTemplate,
	}
}

type weatherResp struct {
	Weather struct {
		WindMin       float64 `json:"wind_min"`
		WindMax       float64 `json:"wind_max"`
		WindDirection float64 `json:"wind_direction"`
		Temperature   float64 `json:"temperature"`
	} `json:"weather"`
}

func (c *Client) fetchURL(stationID int) string {
	if strings.Contains(c.StationURL, "%d") {
		return fmt.Sprintf(c.StationURL, stationID)
	}
	target := fmt.Sprintf("https://www.windguru.net/int/iapi.php?q=station&id_station=%d&weather=false", stationID)
	base := strings.TrimRight(c.StationURL, "/")
	return base + "?url=" + url.QueryEscape(target)
}

func (c *Client) Fetch(stationID int) (*models.WindReading, string, error) {
	fetchURL := c.fetchURL(stationID)

	req, err := http.NewRequest(http.MethodGet, fetchURL, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, string(body), fmt.Errorf("windguru station %d: status %d", stationID, resp.StatusCode)
	}
	if len(body) == 0 {
		return nil, "", fmt.Errorf("windguru station %d: empty response (upload deploy/beget/wg_station.php to Beget)", stationID)
	}
	if body[0] == '<' {
		return nil, string(body), fmt.Errorf("windguru station %d: blocked (HTML response)", stationID)
	}

	var parsed weatherResp
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, string(body), fmt.Errorf("decode windguru %d: %w", stationID, err)
	}

	temp := parsed.Weather.Temperature
	return &models.WindReading{
		Wind:    parsed.Weather.WindMin,
		Gust:    parsed.Weather.WindMax,
		WindDir: parsed.Weather.WindDirection,
		Temp:    &temp,
	}, string(body), nil
}

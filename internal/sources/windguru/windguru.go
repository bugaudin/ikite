package windguru

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/ben/ikite-go/internal/begetproxy"
	"github.com/ben/ikite-go/internal/models"
)

const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/111.0.0.0 Safari/537.36"

const (
	wgOrigin  = "https://www.windguru.cz"
	wgReferer = "https://www.windguru.cz/"
)

func wgHeaders(netHost bool) map[string]string {
	h := map[string]string{
		"accept":          "*/*",
		"accept-language": "en-US,en;q=0.9",
		"origin":          wgOrigin,
		"referer":         wgReferer,
		"user-agent":      userAgent,
	}
	if netHost {
		h["authority"] = "www.windguru.net"
	}
	return h
}

type Client struct {
	Proxy *begetproxy.Client
}

func New(proxy *begetproxy.Client) *Client {
	return &Client{Proxy: proxy}
}

type weatherResp struct {
	Weather struct {
		WindMin       float64 `json:"wind_min"`
		WindMax       float64 `json:"wind_max"`
		WindDirection float64 `json:"wind_direction"`
		Temperature   float64 `json:"temperature"`
	} `json:"weather"`
}

func stationURL(stationID int) string {
	return fmt.Sprintf("https://www.windguru.net/int/iapi.php?q=station&id_station=%d&weather=false", stationID)
}

func (c *Client) Fetch(stationID int) (*models.WindReading, string, error) {
	body, err := c.Proxy.Get(stationURL(stationID), wgHeaders(true))
	if err != nil {
		return nil, string(body), err
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

func forecastSpotURL(spotID int) string {
	return fmt.Sprintf("https://www.windguru.cz/int/iapi.php?q=forecast_spot&id_spot=%d", spotID)
}

func forecastModelURL(spotID int, mr modelRun) string {
	params := url.Values{
		"q":           {"forecast"},
		"id_model":    {fmt.Sprint(mr.IDModel)},
		"rundef":      {mr.Rundef},
		"id_spot":     {fmt.Sprint(spotID)},
		"cachefix":    {mr.Cachefix},
		"WGCACHEABLE": {"21600"},
	}
	return "https://www.windguru.net/int/iapi.php?" + params.Encode()
}

package kyhistory

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ben/ikite-go/internal/models"
	"golang.org/x/net/html"
)

const userAgent = "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/50.0.2661.102 Safari/537.36"

var (
	ktRe     = regexp.MustCompile(`(?i)\s*kt$`)
	rotateRe = regexp.MustCompile(`(?i)rotate\(\s*([\d.]+)deg\)`)
	numRe    = regexp.MustCompile(`[^0-9.\-+]`)
)

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

type Row struct {
	Time        string
	Temperature float64
	Wind        float64
	Gust        float64
	Direction   float64
}

// AlertStats is computed from the first 5 history rows (same as PHP).
type AlertStats struct {
	WindMax float64
	WindMin float64
	GustMax float64
	Temp    float64
	Msg     string
}

func (c *Client) Fetch(now time.Time) ([]models.WindReading, AlertStats, error) {
	req, err := http.NewRequest(http.MethodGet, c.URL, nil)
	if err != nil {
		return nil, AlertStats{}, err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, AlertStats{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, AlertStats{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, AlertStats{}, fmt.Errorf("ky history status %d", resp.StatusCode)
	}

	rows, err := parseTable(string(body))
	if err != nil {
		return nil, AlertStats{}, err
	}

	day := now.Format("2006-01-02")
	stats := AlertStats{WindMin: 99}
	var readings []models.WindReading

	for i, row := range rows {
		period, err := time.ParseInLocation("2006-01-02 15:04:05", day+" "+row.Time+":00", now.Location())
		if err != nil {
			continue
		}
		temp := row.Temperature
		readings = append(readings, models.WindReading{
			Period:   period,
			Location: "ky",
			Wind:     row.Wind,
			Gust:     row.Gust,
			WindDir:  row.Direction,
			Temp:     &temp,
		})

		if i < 5 {
			if i == 0 {
				stats.Temp = row.Temperature
			}
			if row.Wind > stats.WindMax {
				stats.WindMax = row.Wind
			}
			if row.Wind < stats.WindMin {
				stats.WindMin = row.Wind
			}
			if row.Gust > stats.GustMax {
				stats.GustMax = row.Gust
			}
			stats.Msg = fmt.Sprintf("%.0f - %.0f, %.0fC", stats.WindMin, stats.GustMax, stats.Temp)
		}
	}

	return readings, stats, nil
}

func parseTable(rawHTML string) ([]Row, error) {
	doc, err := html.Parse(strings.NewReader(rawHTML))
	if err != nil {
		return nil, err
	}

	var rows []Row
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "tr" {
			cells := tableCells(n)
			if len(cells) >= 5 {
				timeStr := strings.TrimSpace(cells[0].Text)
				if timeStr != "" && timeStr != "Time" && !strings.EqualFold(timeStr, "שעה") {
					wind, _ := strconv.ParseFloat(ktRe.ReplaceAllString(strings.TrimSpace(cells[3].Text), ""), 64)
					gust, _ := strconv.ParseFloat(ktRe.ReplaceAllString(strings.TrimSpace(cells[4].Text), ""), 64)
					tempStr := numRe.ReplaceAllString(strings.TrimSpace(cells[1].Text), "")
					temp, _ := strconv.ParseFloat(tempStr, 64)
					rows = append(rows, Row{
						Time:        timeStr,
						Temperature: temp,
						Wind:        wind,
						Gust:        gust,
						Direction:   cells[2].Direction,
					})
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return rows, nil
}

type cell struct {
	Text      string
	Direction float64
}

func tableCells(tr *html.Node) []cell {
	var cells []cell
	for td := tr.FirstChild; td != nil; td = td.NextSibling {
		if td.Type != html.ElementNode || (td.Data != "td" && td.Data != "th") {
			continue
		}
		c := cell{Text: textContent(td), Direction: directionFromImg(td)}
		cells = append(cells, c)
	}
	return cells
}

func textContent(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var b strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		b.WriteString(textContent(c))
	}
	return b.String()
}

func directionFromImg(n *html.Node) float64 {
	var found float64
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "img" {
			for _, a := range node.Attr {
				if a.Key == "style" {
					if m := rotateRe.FindStringSubmatch(a.Val); len(m) == 2 {
						found, _ = strconv.ParseFloat(m[1], 64)
					}
				}
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return found
}

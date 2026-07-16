package kyhistory

import (
	"os"
	"testing"
	"time"
)

func TestParseAPIWind(t *testing.T) {
	body, err := os.ReadFile("testdata/api_wind_sample.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	loc := time.FixedZone("IST", 3*3600)
	rows, err := parseAPIWind(body, loc)
	if err != nil {
		t.Fatalf("parseAPIWind: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("got %d rows", len(rows))
	}
	r := rows[1]
	if r.Wind != 13 || r.Gust != 20 || r.WindDir != 262 {
		t.Fatalf("wind/gust/dir: %+v", r)
	}
	if r.Temp == nil || *r.Temp != 32.5 {
		t.Fatalf("temp: %+v", r.Temp)
	}
	if r.Humidity == nil || *r.Humidity != 56.9 {
		t.Fatalf("humidity: %+v", r.Humidity)
	}
	if r.Pressure == nil || *r.Pressure != 1009.6 {
		t.Fatalf("pressure: %+v", r.Pressure)
	}
	if r.Period.Format("2006-01-02 15:04:05") != "2026-07-16 00:01:00" {
		t.Fatalf("period: %s", r.Period)
	}
}

func TestParseAPIPeriod(t *testing.T) {
	loc := time.FixedZone("IST", 3*3600)
	ts, err := parseAPIPeriod(`16\07\2026`, "08:06", loc)
	if err != nil {
		t.Fatal(err)
	}
	if ts.Format("2006-01-02 15:04:05") != "2026-07-16 08:06:00" {
		t.Fatalf("got %s", ts)
	}
}

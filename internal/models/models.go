package models

import "time"

type WindReading struct {
	Period    time.Time
	Location  string
	Wind      float64
	Gust      float64
	WindDir   float64
	Temp      *float64
	Humidity  *float64
	Pressure  *float64
}

type Forecast struct {
	Period   time.Time
	Location string
	ReportHe string
	ReportEn string
}

type HomeWind struct {
	Datetime   time.Time
	Wind       float64
	WindSensor float64
}

// Spot is a dashboard column; id matches wind_data.location.
type Spot struct {
	ID                 string
	Name               string
	WindguruStationID  *int
	WindguruID         *int
	SortOrder          int
	Visible            bool
	Collect            bool
	CollectIntervalMin int
	CollectStartHour   int
	CollectEndHour     int
}

type WindForecastRow struct {
	ForecastDate time.Time
	Location     string
	WindguruID   int
	IDModel      int
	Model        string
	Period       time.Time
	Wind         *float64
	Gust         *float64
	WindDir      *float64
	Temp         *float64
}

// ValidForecastHours are allowed AI forecast start/stop hours (0–24).
// Stop 24 means through end of day; otherwise stop is exclusive (stop 22 → no run from 22:00).
func ValidForecastHours() []int {
	h := make([]int, 25)
	for i := range h {
		h[i] = i
	}
	return h
}

func NormalizeForecastHour(v, fallback int) int {
	if v >= 0 && v <= 24 {
		return v
	}
	return fallback
}

// ForecastInWindow reports whether the forecast job may run at nowHour.
func ForecastInWindow(startHour, endHour, nowHour int) bool {
	startHour = NormalizeForecastHour(startHour, 8)
	endHour = NormalizeForecastHour(endHour, 22)
	if startHour > endHour && endHour < 24 {
		startHour, endHour = 8, 22
	}
	if nowHour < startHour {
		return false
	}
	if endHour >= 24 {
		return true
	}
	return nowHour < endHour
}

// ValidCollectIntervals are allowed update intervals (minutes).
var ValidCollectIntervals = []int{1, 5, 10, 15, 30}

// ValidCollectHours are allowed start/stop hours (inclusive).
func ValidCollectHours() []int {
	h := make([]int, 0, 17)
	for i := 6; i <= 22; i++ {
		h = append(h, i)
	}
	return h
}

func NormalizeCollectInterval(v int) int {
	for _, n := range ValidCollectIntervals {
		if v == n {
			return v
		}
	}
	return 5
}

func NormalizeCollectHour(v, fallback int) int {
	if v >= 6 && v <= 22 {
		return v
	}
	return fallback
}

// CollectSkipReason returns why collection should not run, without needing the
// last-collect time. Empty string means the spot may collect (interval still applies).
func (sp Spot) CollectSkipReason(now time.Time) string {
	if !sp.Collect {
		return "collect disabled"
	}
	hour := now.Hour()
	if hour < sp.CollectStartHour || hour >= sp.CollectEndHour {
		return "outside hours"
	}
	return ""
}

// ShouldCollectAt reports whether a reading should be collected now.
// Start hour is inclusive; stop hour is exclusive (stop 22 → no collection from 22:00).
// Example: start 8 / stop 22 → 08:00:00 through 21:59:59.
func (sp Spot) ShouldCollectAt(now time.Time, lastCollect time.Time) (bool, string) {
	if reason := sp.CollectSkipReason(now); reason != "" {
		return false, reason
	}
	if sp.CollectIntervalMin > 0 && !lastCollect.IsZero() {
		minWait := time.Duration(sp.CollectIntervalMin) * time.Minute
		if now.Sub(lastCollect) < minWait {
			return false, "interval not elapsed"
		}
	}
	return true, ""
}

func CardinalDirection(angle float64) string {
	directions := []string{"N", "NE", "E", "SE", "S", "SW", "W", "NW"}
	idx := int((angle/45.0)+0.5) % 8
	if idx < 0 {
		idx += 8
	}
	return directions[idx]
}

func WindCSSClass(wind float64) string {
	switch {
	case wind >= 18:
		return "wind18"
	case wind >= 16:
		return "wind16"
	case wind >= 14:
		return "wind14"
	case wind >= 10:
		return "wind10"
	case wind >= 8:
		return "wind8"
	default:
		return ""
	}
}

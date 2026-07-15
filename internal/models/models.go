package models

import "time"

type WindReading struct {
	Period   time.Time
	Location string
	Wind     float64
	Gust     float64
	WindDir  float64
	Temp     *float64
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
	ID                string
	Name              string
	WindguruStationID *int
	SortOrder         int
	Visible           bool
	Collect           bool
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

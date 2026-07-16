package prediction

import (
	"testing"
	"time"

	"github.com/ben/ikite-go/internal/store"
)

func TestGoodWindowOverlap(t *testing.T) {
	rec := store.PredictionRecord{GoodStartHr: 11, GoodEndHr: 16}
	actual := dayActual{
		hours: map[int]store.HourlyWindStat{
			10: {AvgWind: 8},
			11: {AvgWind: 12},
			12: {AvgWind: 13},
			13: {AvgWind: 14},
			14: {AvgWind: 12},
			15: {AvgWind: 11},
			16: {AvgWind: 9},
			17: {AvgWind: 8},
		},
	}
	got := goodWindowOverlap(rec, actual)
	if got < 0.8 {
		t.Fatalf("overlap %.2f", got)
	}
}

func TestRecordFromForecast(t *testing.T) {
	rec := recordFromForecast(&Result{
		Date:        "2026-07-16",
		GeneratedAt: time.Date(2026, 7, 16, 11, 0, 0, 0, time.UTC),
		SimilarDays: 5,
	}, forecastBundle{
		peak:      hourForecast{AvgWind: 13.2, AvgGust: 15.1, MaxGust: 17.8, AvgDir: 250},
		peakStart: 13,
		peakEnd:   14,
		goodStart: 10,
		goodEnd:   16,
		windDown:  17,
	})
	if rec.PeakStartHr != 13 || rec.PeakWind != 13.2 || rec.SimilarDays != 5 {
		t.Fatalf("%+v", rec)
	}
}

package prediction

import (
	"fmt"
	"math"
	"time"

	"github.com/ben/ikite-go/internal/store"
)

// HistoryScore summarizes how past saved predictions matched actual wind_data.
type HistoryScore struct {
	EvaluatedDays     int     `json:"evaluated_days"`
	AvgPeakWindError  float64 `json:"avg_peak_wind_error"`
	AvgPeakHourError  float64 `json:"avg_peak_hour_error"`
	PeakInWindowPct   float64 `json:"peak_in_window_pct"`
	GoodWindowOverlap float64 `json:"good_window_overlap"`
	Summary           string  `json:"summary"`
}

type dayActual struct {
	peakHr   int
	peakWind float64
	peakGust float64
	peakDir  float64
	hours    map[int]store.HourlyWindStat
}

func evaluateHistory(st *store.Store, beforeDay time.Time, loc *time.Location) (*HistoryScore, error) {
	records, err := st.ListPredictionsBefore(beforeDay)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, nil
	}

	var (
		n               int
		sumWindErr      float64
		sumHourErr      float64
		sumGoodOverlap  float64
		peakInWindow    int
	)

	for _, rec := range records {
		actual, ok, err := actualForDay(st, rec.TargetDate.In(loc))
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		n++
		sumWindErr += math.Abs(actual.peakWind - rec.PeakWind)
		predPeakHr := rec.PeakStartHr
		if rec.PeakEndHr > rec.PeakStartHr {
			predPeakHr = (rec.PeakStartHr + rec.PeakEndHr) / 2
		}
		sumHourErr += math.Abs(float64(actual.peakHr - predPeakHr))
		if actual.peakHr >= rec.PeakStartHr && actual.peakHr <= rec.PeakEndHr {
			peakInWindow++
		}
		sumGoodOverlap += goodWindowOverlap(rec, actual)
	}

	if n == 0 {
		return nil, nil
	}

	score := &HistoryScore{
		EvaluatedDays:     n,
		AvgPeakWindError:  sumWindErr / float64(n),
		AvgPeakHourError:  sumHourErr / float64(n),
		PeakInWindowPct:   100 * float64(peakInWindow) / float64(n),
		GoodWindowOverlap: 100 * sumGoodOverlap / float64(n),
	}
	score.Summary = fmt.Sprintf(
		"%d past predictions: peak wind off by %.1f kt, peak hour off by %.0f h, %.0f%% peaks in window, %.0f%% good-window overlap",
		n, score.AvgPeakWindError, score.AvgPeakHourError,
		score.PeakInWindowPct, score.GoodWindowOverlap,
	)
	return score, nil
}

func actualForDay(st *store.Store, day time.Time) (dayActual, bool, error) {
	rows, err := st.KyDayHourlyStats(day)
	if err != nil {
		return dayActual{}, false, err
	}
	if len(rows) < 6 {
		return dayActual{}, false, nil
	}
	hours := map[int]store.HourlyWindStat{}
	for _, h := range rows {
		hours[h.Hour] = h
	}

	bestHr := -1
	bestWind := -1.0
	var bestGust, bestDir float64
	for hr := 9; hr <= 18; hr++ {
		h, ok := hours[hr]
		if !ok || h.AvgWind <= bestWind {
			continue
		}
		bestWind = h.AvgWind
		bestHr = hr
		bestGust = h.AvgGust
		bestDir = h.AvgDir
	}
	if bestHr < 0 {
		return dayActual{}, false, nil
	}
	return dayActual{
		peakHr:   bestHr,
		peakWind: bestWind,
		peakGust: bestGust,
		peakDir:  bestDir,
		hours:    hours,
	}, true, nil
}

func goodWindowOverlap(rec store.PredictionRecord, actual dayActual) float64 {
	var predGood, actualGood, both int
	for hr := 6; hr <= 20; hr++ {
		pred := hr >= rec.GoodStartHr && hr <= rec.GoodEndHr
		h, ok := actual.hours[hr]
		actual := ok && h.AvgWind >= goodWindKt
		if pred {
			predGood++
		}
		if actual {
			actualGood++
		}
		if pred && actual {
			both++
		}
	}
	union := predGood + actualGood - both
	if union == 0 {
		return 0
	}
	return float64(both) / float64(union)
}

func recordFromForecast(res *Result, fc forecastBundle) store.PredictionRecord {
	wLo := int(math.Floor(fc.peak.AvgWind))
	wHi := int(math.Ceil(fc.peak.AvgWind))
	if wHi <= wLo {
		wHi = wLo + 1
	}
	gLo := int(math.Round(fc.peak.AvgGust))
	gHi := int(math.Round(fc.peak.MaxGust))
	if gHi <= gLo {
		gHi = gLo + 1
	}
	day, _ := time.Parse("2006-01-02", res.Date)
	return store.PredictionRecord{
		TargetDate:  day,
		CreatedAt:   res.GeneratedAt,
		PeakStartHr: fc.peakStart,
		PeakEndHr:   fc.peakEnd,
		PeakWind:    fc.peak.AvgWind,
		PeakWindLo:  float64(wLo),
		PeakWindHi:  float64(wHi),
		PeakGust:    fc.peak.AvgGust,
		PeakGustMax: fc.peak.MaxGust,
		PeakDir:     fc.peak.AvgDir,
		GoodStartHr: fc.goodStart,
		GoodEndHr:   fc.goodEnd,
		WindDownHr:  fc.windDown,
		SimilarDays: res.SimilarDays,
	}
}

func savePrediction(st *store.Store, res *Result, fc forecastBundle) error {
	return st.SavePrediction(recordFromForecast(res, fc))
}

package store

import (
	"database/sql"
	"time"
)

// PredictionRecord is a saved KY thermal prediction for one calendar day.
type PredictionRecord struct {
	TargetDate   time.Time
	CreatedAt    time.Time
	PeakStartHr  int
	PeakEndHr    int
	PeakWind     float64
	PeakWindLo   float64
	PeakWindHi   float64
	PeakGust     float64
	PeakGustMax  float64
	PeakDir      float64
	GoodStartHr  int
	GoodEndHr    int
	WindDownHr   int
	SimilarDays  int
}

// SavePrediction upserts the prediction for target_date.
func (s *Store) SavePrediction(rec PredictionRecord) error {
	_, err := s.DB.Exec(`
		INSERT INTO prediction_history (
			target_date, created_at,
			peak_start_hr, peak_end_hr,
			peak_wind, peak_wind_lo, peak_wind_hi,
			peak_gust, peak_gust_max, peak_dir,
			good_start_hr, good_end_hr, wind_down_hr, similar_days
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			created_at = VALUES(created_at),
			peak_start_hr = VALUES(peak_start_hr),
			peak_end_hr = VALUES(peak_end_hr),
			peak_wind = VALUES(peak_wind),
			peak_wind_lo = VALUES(peak_wind_lo),
			peak_wind_hi = VALUES(peak_wind_hi),
			peak_gust = VALUES(peak_gust),
			peak_gust_max = VALUES(peak_gust_max),
			peak_dir = VALUES(peak_dir),
			good_start_hr = VALUES(good_start_hr),
			good_end_hr = VALUES(good_end_hr),
			wind_down_hr = VALUES(wind_down_hr),
			similar_days = VALUES(similar_days)`,
		rec.TargetDate.Format("2006-01-02"), rec.CreatedAt,
		rec.PeakStartHr, rec.PeakEndHr,
		rec.PeakWind, rec.PeakWindLo, rec.PeakWindHi,
		rec.PeakGust, rec.PeakGustMax, rec.PeakDir,
		rec.GoodStartHr, rec.GoodEndHr, rec.WindDownHr, rec.SimilarDays,
	)
	return err
}

// ListPredictionsBefore returns saved predictions with target_date strictly before day.
func (s *Store) ListPredictionsBefore(day time.Time) ([]PredictionRecord, error) {
	rows, err := s.DB.Query(`
		SELECT target_date, created_at,
			peak_start_hr, peak_end_hr,
			peak_wind, peak_wind_lo, peak_wind_hi,
			peak_gust, peak_gust_max, peak_dir,
			good_start_hr, good_end_hr, wind_down_hr, similar_days
		FROM prediction_history
		WHERE target_date < ?
		ORDER BY target_date DESC
		LIMIT 120`, day.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []PredictionRecord
	for rows.Next() {
		var rec PredictionRecord
		var targetDate time.Time
		if err := rows.Scan(
			&targetDate, &rec.CreatedAt,
			&rec.PeakStartHr, &rec.PeakEndHr,
			&rec.PeakWind, &rec.PeakWindLo, &rec.PeakWindHi,
			&rec.PeakGust, &rec.PeakGustMax, &rec.PeakDir,
			&rec.GoodStartHr, &rec.GoodEndHr, &rec.WindDownHr, &rec.SimilarDays,
		); err != nil {
			return nil, err
		}
		rec.TargetDate = targetDate
		out = append(out, rec)
	}
	return out, rows.Err()
}

// PredictionExists returns whether a prediction was saved for the given calendar day.
func (s *Store) PredictionExists(day time.Time) (bool, error) {
	var n int
	err := s.DB.QueryRow(`
		SELECT COUNT(*) FROM prediction_history WHERE target_date = ?`,
		day.Format("2006-01-02")).Scan(&n)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return n > 0, err
}

package store

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ben/ikite-go/internal/models"
)

type Store struct {
	DB *sql.DB
}

func New(db *sql.DB) *Store {
	return &Store{DB: db}
}

func (s *Store) InsertWind(r models.WindReading) error {
	_, err := s.DB.Exec(`
		INSERT INTO wind_data (period, location, wind, gust, wind_dir, temp, humidity, pressure)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			wind = VALUES(wind),
			gust = VALUES(gust),
			wind_dir = VALUES(wind_dir),
			temp = VALUES(temp),
			humidity = VALUES(humidity),
			pressure = VALUES(pressure)`,
		r.Period, r.Location, r.Wind, r.Gust, r.WindDir, r.Temp, r.Humidity, r.Pressure,
	)
	return err
}

func (s *Store) InsertWindLog(period time.Time, location, raw string) error {
	_, err := s.DB.Exec(`
		INSERT IGNORE INTO wind_data_log (period, location, raw)
		VALUES (?, ?, ?)`, period, location, raw)
	return err
}

func (s *Store) GetSetting(key string) (string, error) {
	var val string
	err := s.DB.QueryRow(`SELECT s_val FROM settings WHERE s_key = ?`, key).Scan(&val)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return val, err
}

func (s *Store) SetSetting(key, val string) error {
	_, err := s.DB.Exec(`REPLACE INTO settings (s_key, s_val) VALUES (?, ?)`, key, val)
	return err
}

func (s *Store) Threshold() (float64, error) {
	val, err := s.GetSetting("threshold")
	if err != nil {
		return 10, err
	}
	if val == "" {
		return 10, nil
	}
	var n float64
	_, err = fmt.Sscanf(val, "%f", &n)
	if err != nil {
		return 10, nil
	}
	return n, nil
}

func (s *Store) ForecastTelegramEnabled() (bool, error) {
	val, err := s.GetSetting("forecast_telegram")
	if err != nil {
		return true, err
	}
	return val != "no", nil
}

func (s *Store) PredictionTelegramEnabled() (bool, error) {
	val, err := s.GetSetting("prediction_telegram")
	if err != nil {
		return true, err
	}
	return val != "no", nil
}

func (s *Store) ForecastSchedule() (startHour, endHour int, err error) {
	startHour, endHour = 8, 22
	if v, err := s.GetSetting("forecast_start_hour"); err != nil {
		return startHour, endHour, err
	} else if v != "" {
		if n, e := strconv.Atoi(v); e == nil {
			startHour = models.NormalizeForecastHour(n, 8)
		}
	}
	if v, err := s.GetSetting("forecast_end_hour"); err != nil {
		return startHour, endHour, err
	} else if v != "" {
		if n, e := strconv.Atoi(v); e == nil {
			endHour = models.NormalizeForecastHour(n, 22)
		}
	}
	if startHour > endHour && endHour < 24 {
		startHour, endHour = 8, 22
	}
	return startHour, endHour, nil
}

func (s *Store) SetForecastSchedule(startHour, endHour int) error {
	startHour = models.NormalizeForecastHour(startHour, 8)
	endHour = models.NormalizeForecastHour(endHour, 22)
	if startHour > endHour && endHour < 24 {
		startHour, endHour = 8, 22
	}
	if err := s.SetSetting("forecast_start_hour", strconv.Itoa(startHour)); err != nil {
		return err
	}
	return s.SetSetting("forecast_end_hour", strconv.Itoa(endHour))
}

func splitKeys(val string) []string {
	parts := strings.Split(val, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func (s *Store) LatestGust(location string) (float64, error) {
	var gust float64
	err := s.DB.QueryRow(`
		SELECT gust FROM wind_data WHERE location = ?
		ORDER BY period DESC LIMIT 1`, location).Scan(&gust)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return gust, err
}

func (s *Store) LatestWindPeriod(location string) (time.Time, error) {
	var period time.Time
	err := s.DB.QueryRow(`
		SELECT period FROM wind_data WHERE location = ?
		ORDER BY period DESC LIMIT 1`, location).Scan(&period)
	if err == sql.ErrNoRows {
		return time.Time{}, nil
	}
	return period, err
}

func (s *Store) LatestWind(location string) (float64, error) {
	var wind float64
	err := s.DB.QueryRow(`
		SELECT wind FROM wind_data WHERE location = ?
		ORDER BY period DESC LIMIT 1`, location).Scan(&wind)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return wind, err
}

func (s *Store) ListWind(from, to time.Time) ([]models.WindReading, error) {
	rows, err := s.DB.Query(`
		SELECT period, location, wind, gust, wind_dir, temp, humidity, pressure
		FROM wind_data
		WHERE period > ? AND period < ?
		ORDER BY period DESC`, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.WindReading
	for rows.Next() {
		var r models.WindReading
		var temp, humidity, pressure sql.NullFloat64
		if err := rows.Scan(&r.Period, &r.Location, &r.Wind, &r.Gust, &r.WindDir, &temp, &humidity, &pressure); err != nil {
			return nil, err
		}
		if temp.Valid {
			t := temp.Float64
			r.Temp = &t
		}
		if humidity.Valid {
			h := humidity.Float64
			r.Humidity = &h
		}
		if pressure.Valid {
			p := pressure.Float64
			r.Pressure = &p
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) LatestForecast(location string) (*models.Forecast, error) {
	var f models.Forecast
	err := s.DB.QueryRow(`
		SELECT period, location, report_he, report_en
		FROM wind_forecast_ai
		WHERE location = ?
		ORDER BY period DESC LIMIT 1`, location).
		Scan(&f.Period, &f.Location, &f.ReportHe, &f.ReportEn)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (s *Store) InsertForecast(f models.Forecast) error {
	_, err := s.DB.Exec(`
		INSERT INTO wind_forecast_ai (period, location, report_he, report_en)
		VALUES (?, ?, ?, ?)`,
		f.Period, f.Location, f.ReportHe, f.ReportEn)
	return err
}

func (s *Store) InsertHomeWind(h models.HomeWind) error {
	_, err := s.DB.Exec(`
		INSERT IGNORE INTO wind_home (datetime, wind, wind_sensor)
		VALUES (?, ?, ?)`, h.Datetime, h.Wind, h.WindSensor)
	return err
}

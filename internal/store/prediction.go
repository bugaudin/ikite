package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// HourlyWindStat aggregates wind readings for one clock hour.
type HourlyWindStat struct {
	Hour        int
	AvgWind     float64
	MaxWind     float64
	AvgGust     float64
	MaxGust     float64
	AvgDir      float64
	AvgTemp     sql.NullFloat64
	AvgHumidity sql.NullFloat64
	AvgPressure sql.NullFloat64
	Count       int
}

// DayProfile is one calendar day of hourly KY readings.
type DayProfile struct {
	Date  time.Time
	Hours map[int]HourlyWindStat
}

const hourlySelect = `
	SELECT HOUR(period) AS hr,
		AVG(wind), MAX(wind), AVG(gust), MAX(gust), AVG(wind_dir),
		AVG(temp), AVG(humidity), AVG(pressure), COUNT(*)`

// KySummerDayProfiles returns per-day hourly aggregates for July+August KY history.
// Days on or after excludeDay are omitted (typically today).
func (s *Store) KySummerDayProfiles(excludeDay time.Time) ([]DayProfile, error) {
	summerClause, summerArgs := summerPeriodClause(2010, excludeDay.Year())
	q := `
		SELECT DATE(period) AS d, HOUR(period) AS hr,
			AVG(wind), MAX(wind), AVG(gust), MAX(gust), AVG(wind_dir),
			AVG(temp), AVG(humidity), AVG(pressure), COUNT(*)
		FROM wind_data
		WHERE location = 'ky' AND (` + summerClause + `)`
	args := append([]any{}, summerArgs...)
	if !excludeDay.IsZero() {
		q += ` AND period < ?`
		args = append(args, excludeDay)
	}
	q += ` GROUP BY DATE(period), HOUR(period) ORDER BY d, hr`

	rows, err := s.DB.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDayProfiles(rows)
}

// KyJulyDayProfiles is an alias for KySummerDayProfiles.
func (s *Store) KyJulyDayProfiles(excludeDay time.Time) ([]DayProfile, error) {
	return s.KySummerDayProfiles(excludeDay)
}

// KyDayHourlyStats returns hourly averages for one calendar day.
func (s *Store) KyDayHourlyStats(day time.Time) ([]HourlyWindStat, error) {
	start := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location())
	end := start.Add(24 * time.Hour)
	rows, err := s.DB.Query(hourlySelect+`
		FROM wind_data
		WHERE location = 'ky' AND period >= ? AND period < ?
		GROUP BY HOUR(period) ORDER BY hr`, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanHourlyStats(rows)
}

// KyLatestReading returns the most recent KY observation.
func (s *Store) KyLatestReading() (wind, gust, dir float64, temp, humidity, pressure sql.NullFloat64, err error) {
	err = s.DB.QueryRow(`
		SELECT wind, gust, wind_dir, temp, humidity, pressure
		FROM wind_data WHERE location = 'ky'
		ORDER BY period DESC LIMIT 1`).
		Scan(&wind, &gust, &dir, &temp, &humidity, &pressure)
	if err == sql.ErrNoRows {
		err = nil
	}
	return
}

// summerPeriodClause builds index-friendly range OR clauses for July and August.
func summerPeriodClause(fromYear, throughYear int) (string, []any) {
	var parts []string
	var args []any
	for y := fromYear; y <= throughYear; y++ {
		parts = append(parts, "(period >= ? AND period < ?)")
		args = append(args,
			fmt.Sprintf("%d-07-01 00:00:00", y),
			fmt.Sprintf("%d-08-01 00:00:00", y),
		)
		parts = append(parts, "(period >= ? AND period < ?)")
		args = append(args,
			fmt.Sprintf("%d-08-01 00:00:00", y),
			fmt.Sprintf("%d-09-01 00:00:00", y),
		)
	}
	return strings.Join(parts, " OR "), args
}

func scanDayProfiles(rows *sql.Rows) ([]DayProfile, error) {
	byDate := map[string]*DayProfile{}
	var order []string
	for rows.Next() {
		var day time.Time
		var st HourlyWindStat
		if err := rows.Scan(
			&day, &st.Hour, &st.AvgWind, &st.MaxWind, &st.AvgGust, &st.MaxGust,
			&st.AvgDir, &st.AvgTemp, &st.AvgHumidity, &st.AvgPressure, &st.Count,
		); err != nil {
			return nil, err
		}
		key := day.Format("2006-01-02")
		if byDate[key] == nil {
			byDate[key] = &DayProfile{
				Date:  time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location()),
				Hours: map[int]HourlyWindStat{},
			}
			order = append(order, key)
		}
		byDate[key].Hours[st.Hour] = st
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	out := make([]DayProfile, 0, len(order))
	for _, key := range order {
		out = append(out, *byDate[key])
	}
	return out, nil
}

func scanHourlyStats(rows *sql.Rows) ([]HourlyWindStat, error) {
	var out []HourlyWindStat
	for rows.Next() {
		var st HourlyWindStat
		if err := rows.Scan(
			&st.Hour, &st.AvgWind, &st.MaxWind, &st.AvgGust, &st.MaxGust,
			&st.AvgDir, &st.AvgTemp, &st.AvgHumidity, &st.AvgPressure, &st.Count,
		); err != nil {
			return nil, err
		}
		out = append(out, st)
	}
	return out, rows.Err()
}

// TotalReadings sums observation counts from day profiles.
func TotalReadings(profiles []DayProfile) int {
	n := 0
	for _, d := range profiles {
		for _, h := range d.Hours {
			n += h.Count
		}
	}
	return n
}

// BaselineHourlyStats averages each clock hour across all day profiles.
func BaselineHourlyStats(profiles []DayProfile) []HourlyWindStat {
	type acc struct {
		st HourlyWindStat
		n  int
	}
	byHour := map[int]*acc{}
	for _, day := range profiles {
		for hr, h := range day.Hours {
			a := byHour[hr]
			if a == nil {
				a = &acc{}
				byHour[hr] = a
			}
			a.st.AvgWind += h.AvgWind
			a.st.AvgGust += h.AvgGust
			a.st.AvgDir += h.AvgDir
			a.st.MaxWind = maxf(a.st.MaxWind, h.MaxWind)
			a.st.MaxGust = maxf(a.st.MaxGust, h.MaxGust)
			a.st.Count += h.Count
			if h.AvgTemp.Valid {
				if !a.st.AvgTemp.Valid {
					a.st.AvgTemp = h.AvgTemp
				} else {
					a.st.AvgTemp.Float64 += h.AvgTemp.Float64
				}
			}
			a.n++
		}
	}
	out := make([]HourlyWindStat, 0, len(byHour))
	for hr, a := range byHour {
		if a.n == 0 {
			continue
		}
		st := a.st
		st.Hour = hr
		st.AvgWind /= float64(a.n)
		st.AvgGust /= float64(a.n)
		st.AvgDir /= float64(a.n)
		if st.AvgTemp.Valid {
			st.AvgTemp.Float64 /= float64(a.n)
		}
		out = append(out, st)
	}
	return out
}

func maxf(a, b float64) float64 {
	if b > a {
		return b
	}
	return a
}

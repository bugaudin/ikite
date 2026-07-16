package store

import (
	"testing"
	"time"
)

func TestSummerPeriodClause(t *testing.T) {
	clause, args := summerPeriodClause(2024, 2026)
	if clause == "" || len(args) != 12 {
		t.Fatalf("clause=%q args=%d want 12", clause, len(args))
	}
}

func TestBaselineHourlyStats(t *testing.T) {
	profiles := []DayProfile{
		{
			Date: time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC),
			Hours: map[int]HourlyWindStat{
				13: {Hour: 13, AvgWind: 12, AvgGust: 14, AvgDir: 250, Count: 60},
			},
		},
		{
			Date: time.Date(2024, 7, 2, 0, 0, 0, 0, time.UTC),
			Hours: map[int]HourlyWindStat{
				13: {Hour: 13, AvgWind: 14, AvgGust: 16, AvgDir: 255, Count: 60},
			},
		},
	}
	base := BaselineHourlyStats(profiles)
	if len(base) != 1 || base[0].AvgWind != 13 {
		t.Fatalf("baseline %+v", base)
	}
	if TotalReadings(profiles) != 120 {
		t.Fatalf("count %d", TotalReadings(profiles))
	}
}

package models

import (
	"testing"
	"time"
)

func TestSpotShouldCollectAt(t *testing.T) {
	sp := Spot{
		Collect:            true,
		CollectIntervalMin: 5,
		CollectStartHour:   8,
		CollectEndHour:     22,
	}
	loc := time.FixedZone("IST", 3*3600)
	now := time.Date(2026, 7, 15, 10, 0, 0, 0, loc)
	last := now.Add(-6 * time.Minute)

	ok, _ := sp.ShouldCollectAt(now, last)
	if !ok {
		t.Fatal("expected collect after interval")
	}

	ok, reason := sp.ShouldCollectAt(now, now.Add(-2*time.Minute))
	if ok || reason != "interval not elapsed" {
		t.Fatalf("expected interval skip, got ok=%v reason=%q", ok, reason)
	}

	ok, reason = sp.ShouldCollectAt(time.Date(2026, 7, 15, 7, 0, 0, 0, loc), time.Time{})
	if ok || reason != "outside hours" {
		t.Fatalf("expected outside hours, got ok=%v reason=%q", ok, reason)
	}

	ok, reason = sp.ShouldCollectAt(time.Date(2026, 7, 15, 22, 30, 0, 0, loc), time.Time{})
	if ok || reason != "outside hours" {
		t.Fatalf("expected outside hours at stop hour, got ok=%v reason=%q", ok, reason)
	}

	ok, _ = sp.ShouldCollectAt(time.Date(2026, 7, 15, 21, 30, 0, 0, loc), time.Time{})
	if !ok {
		t.Fatal("expected collect during last allowed hour")
	}
}

func TestNormalizeCollectInterval(t *testing.T) {
	if got := NormalizeCollectInterval(15); got != 15 {
		t.Fatalf("got %d", got)
	}
	if got := NormalizeCollectInterval(99); got != 5 {
		t.Fatalf("got %d", got)
	}
}

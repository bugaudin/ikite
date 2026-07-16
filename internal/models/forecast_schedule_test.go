package models

import "testing"

func TestForecastInWindow(t *testing.T) {
	if !ForecastInWindow(8, 22, 10) {
		t.Fatal("expected inside window")
	}
	if ForecastInWindow(8, 22, 7) {
		t.Fatal("expected before start")
	}
	if ForecastInWindow(8, 22, 22) {
		t.Fatal("expected exclusive stop at 22")
	}
	if !ForecastInWindow(8, 22, 21) {
		t.Fatal("expected last hour before stop")
	}
	if !ForecastInWindow(8, 24, 23) {
		t.Fatal("expected end 24 through midnight")
	}
}

package prediction

import (
	"strings"
	"testing"
)

func TestTelegramMessage(t *testing.T) {
	msg := (&Result{
		Date:            "2026-07-16",
		StrongestWindow: "13:00 – 14:30",
		ExpectedPeak:    "13–14 kt sustained, gusts 16–18 kt",
		Direction:       "235–265° (WSW)",
		GoodWindow:      "10:00 – 16:30",
		WindDown:        "~17:00",
		SimilarDays:     31,
	}).TelegramMessage()
	if !strings.Contains(msg, "KY thermal prediction") || !strings.Contains(msg, "13:00 – 14:30") {
		t.Fatalf("msg %q", msg)
	}
}

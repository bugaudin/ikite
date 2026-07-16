package prediction

import (
	"fmt"
	"strings"
)

// TelegramMessage formats the prediction for the AI forecast Telegram bot.
func (r *Result) TelegramMessage() string {
	if r == nil {
		return ""
	}
	var b strings.Builder
	fmt.Fprintf(&b, "KY thermal prediction — %s\n\n", r.Date)
	fmt.Fprintf(&b, "Strongest window: %s\n", r.StrongestWindow)
	fmt.Fprintf(&b, "Expected peak: %s\n", r.ExpectedPeak)
	fmt.Fprintf(&b, "Direction: %s\n", r.Direction)
	fmt.Fprintf(&b, "Good window: %s\n", r.GoodWindow)
	fmt.Fprintf(&b, "Wind down: %s\n", r.WindDown)
	if r.SimilarDays > 0 {
		fmt.Fprintf(&b, "\n%d similar good-wind summer days\n", r.SimilarDays)
	}
	if r.History != nil && r.History.Summary != "" {
		fmt.Fprintf(&b, "\nPast accuracy: %s\n", r.History.Summary)
	}
	if r.Current != "" {
		fmt.Fprintf(&b, "Now: %s\n", strings.TrimSuffix(r.Current, " (now)"))
	}
	fmt.Fprintf(&b, "\nhttps://ikite.fyi/prediction")
	return b.String()
}

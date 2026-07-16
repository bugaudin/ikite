package collector

import (
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/ben/ikite-go/internal/config"
	"github.com/ben/ikite-go/internal/models"
	"github.com/ben/ikite-go/internal/notify/telegram"
	"github.com/ben/ikite-go/internal/sources/kyhistory"
	"github.com/ben/ikite-go/internal/sources/windguru"
	"github.com/ben/ikite-go/internal/sources/windometer"
	"github.com/ben/ikite-go/internal/store"
)

type Service struct {
	Cfg      *config.Config
	Store    *store.Store
	WG       *windguru.Client
	KY       *kyhistory.Client
	KH       *windometer.Client
	Telegram *telegram.Client
	Log      *slog.Logger
}

type Result struct {
	WindKY     float64
	WindKH     float64
	MsgKY      string
	MsgNorth   string
	MsgBG      string
	AlertSent  bool
	SavedCount int
}

func (s *Service) shouldCollectSpot(sp models.Spot, now time.Time) (bool, string) {
	if reason := sp.CollectSkipReason(now); reason != "" {
		return false, reason
	}
	last, err := s.Store.LatestWindPeriod(sp.ID)
	if err != nil {
		s.Log.Warn("latest wind period", "loc", sp.ID, "err", err)
	}
	return sp.ShouldCollectAt(now, last)
}

// RunWindguruStation fetches and saves one Windguru station (used by per-station timers).
// Windguru/Beget HTTP is only used when schedule checks pass.
func (s *Service) RunWindguruStation(now time.Time, stationID int) error {
	st, err := s.Store.SpotByWindguruID(stationID)
	if err != nil {
		return err
	}

	now = now.In(s.Cfg.Timezone)
	if ok, reason := s.shouldCollectSpot(*st, now); !ok {
		s.Log.Info("windguru skipped", "station", stationID, "loc", st.ID, "reason", reason)
		return nil
	}

	delay := 2 + rand.Intn(9) // 2–10 seconds
	s.Log.Info("windguru delay", "station", stationID, "seconds", delay)
	time.Sleep(time.Duration(delay) * time.Second)

	reading, _, err := s.WG.Fetch(stationID)
	if err != nil {
		return err
	}

	reading.Period = now
	reading.Location = st.ID
	if err := s.Store.InsertWind(*reading); err != nil {
		return fmt.Errorf("insert wg wind: %w", err)
	}

	s.Log.Info("saved windguru", "station", stationID, "loc", st.ID, "wind", reading.Wind, "gust", reading.Gust)
	return nil
}

func (s *Service) Run(now time.Time) (*Result, error) {
	now = now.In(s.Cfg.Timezone)
	hour := now.Hour()
	res := &Result{}

	threshold, err := s.Store.Threshold()
	if err != nil {
		return nil, fmt.Errorf("threshold: %w", err)
	}

	if ky, err := s.Store.SpotByID("ky"); err == nil {
		if ok, reason := s.shouldCollectSpot(*ky, now); ok {
			readings, stats, err := s.KY.Fetch(now)
			if err != nil {
				s.Log.Error("ky history fetch failed", "err", err)
			} else {
				for _, r := range readings {
					if err := s.Store.InsertWind(r); err != nil {
						s.Log.Error("insert ky wind", "err", err, "period", r.Period)
					} else {
						res.SavedCount++
					}
				}
				res.WindKY = stats.WindMax
				res.MsgKY = stats.Msg
				s.Log.Info("ky history saved", "rows", len(readings), "wind_ky", res.WindKY)
			}
		} else {
			s.Log.Info("ky skipped", "loc", ky.ID, "reason", reason)
		}
	}

	var kh *models.WindReading
	if spot, err := s.Store.SpotByID("kh"); err == nil {
		if ok, reason := s.shouldCollectSpot(*spot, now); ok {
			var err error
			kh, _, err = s.KH.Fetch(now)
			if err != nil {
				s.Log.Warn("windometer fetch failed", "err", err)
			} else if kh != nil {
				if err := s.Store.InsertWind(*kh); err != nil {
					s.Log.Error("insert kh wind", "err", err)
				} else {
					res.SavedCount++
					res.WindKH = kh.Wind
					s.Log.Info("saved kh", "wind", kh.Wind, "gust", kh.Gust)
				}
			}
		} else {
			s.Log.Info("kh skipped", "loc", spot.ID, "reason", reason)
			if w, err := s.Store.LatestWind("kh"); err == nil {
				res.WindKH = w
			}
		}
	}

	if res.WindKY == 0 {
		if w, err := s.Store.LatestWind("ky"); err == nil {
			res.WindKY = w
		}
	}

	gustST, _ := s.Store.LatestGust("st")
	gustBG, _ := s.Store.LatestGust("bg")
	gustBetzet, _ := s.Store.LatestGust("15233")

	res.MsgNorth = fmt.Sprintf("%d", int(gustST))
	if gustBetzet > 0 {
		res.MsgNorth = fmt.Sprintf("%d", int(gustBetzet))
	}
	res.MsgBG = fmt.Sprintf("%d", int(gustBG))

	if res.MsgKY == "" && res.WindKH > 0 {
		res.MsgKY = fmt.Sprintf("%.0f - %.0f", res.WindKH, res.WindKH)
	}

	shouldAlert := hour >= s.Cfg.AlertStartHour &&
		hour <= s.Cfg.AlertEndHour &&
		threshold != 999 &&
		(res.WindKY >= threshold || res.WindKH >= threshold)

	if shouldAlert && s.Telegram.Enabled() {
		msg := fmt.Sprintf("%s | %s | %s", res.MsgNorth, res.MsgKY, res.MsgBG)
		if err := s.Telegram.Send(msg); err != nil {
			s.Log.Error("telegram alert failed", "err", err)
		} else {
			res.AlertSent = true
			s.Log.Info("telegram alert sent", "msg", msg)
		}
	}

	return res, nil
}

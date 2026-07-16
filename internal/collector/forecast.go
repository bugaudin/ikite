package collector

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/ben/ikite-go/internal/config"
	"github.com/ben/ikite-go/internal/models"
	"github.com/ben/ikite-go/internal/notify/telegram"
	"github.com/ben/ikite-go/internal/sources/surfo"
	"github.com/ben/ikite-go/internal/store"
	"github.com/ben/ikite-go/internal/translate"
)

type ForecastService struct {
	Cfg      *config.Config
	Store    *store.Store
	Surfo    *surfo.Client
	Translate *translate.Client
	Telegram *telegram.Client
	Log      *slog.Logger
}

func (s *ForecastService) Run(now time.Time) error {
	now = now.In(s.Cfg.Timezone)

	start, end, err := s.Store.ForecastSchedule()
	if err != nil {
		return fmt.Errorf("forecast schedule: %w", err)
	}
	if !models.ForecastInWindow(start, end, now.Hour()) {
		s.Log.Info("forecast skipped", "reason", "outside hours", "start", start, "end", end)
		return nil
	}

	reportHe, err := s.Surfo.FetchReport()
	if err != nil {
		return fmt.Errorf("fetch surfo report: %w", err)
	}

	latest, err := s.Store.LatestForecast("ky")
	if err != nil {
		return err
	}
	if latest != nil && latest.ReportHe == reportHe {
		s.Log.Info("forecast unchanged")
		return nil
	}

	reportEn, err := s.Translate.HebrewToEnglish(reportHe)
	if err != nil {
		s.Log.Warn("translation failed, storing Hebrew as English", "err", err)
		reportEn = reportHe
	}

	f := models.Forecast{
		Period:   now,
		Location: "ky",
		ReportHe: reportHe,
		ReportEn: reportEn,
	}
	if err := s.Store.InsertForecast(f); err != nil {
		return fmt.Errorf("insert forecast: %w", err)
	}

	if s.Telegram.Enabled() {
		enabled, err := s.Store.ForecastTelegramEnabled()
		if err != nil {
			s.Log.Error("forecast telegram setting", "err", err)
			enabled = true
		}
		if enabled {
			if err := s.Telegram.Send(reportEn); err != nil {
				s.Log.Error("telegram forecast failed", "err", err)
			}
		} else {
			s.Log.Info("forecast telegram disabled in settings")
		}
	}

	s.Log.Info("forecast saved and notified")
	return nil
}

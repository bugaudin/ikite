package collector

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/ben/ikite-go/internal/config"
	"github.com/ben/ikite-go/internal/notify/telegram"
	"github.com/ben/ikite-go/internal/prediction"
	"github.com/ben/ikite-go/internal/store"
)

const predictionNotifyHour = 11

type PredictionService struct {
	Cfg      *config.Config
	Store    *store.Store
	Telegram *telegram.Client
	Log      *slog.Logger
}

type PredictionOptions struct {
	Force bool // skip Jul–Aug / 11am window (for manual runs)
}

func summerMonth(m time.Month) bool {
	return m == time.July || m == time.August
}

func (s *PredictionService) Run(now time.Time, opts PredictionOptions) error {
	now = now.In(s.Cfg.Timezone)
	if !opts.Force {
		if !summerMonth(now.Month()) {
			s.Log.Info("prediction notify skipped", "reason", "not summer", "month", now.Month())
			return nil
		}
		if now.Hour() != predictionNotifyHour {
			s.Log.Info("prediction notify skipped", "reason", "not 11am", "hour", now.Hour())
			return nil
		}
	}

	res, err := prediction.Compute(s.Store, now, s.Cfg.Timezone)
	if err != nil {
		return fmt.Errorf("prediction compute: %w", err)
	}

	if !s.Telegram.Enabled() {
		s.Log.Info("prediction computed", "similar_days", res.SimilarDays)
		return nil
	}

	enabled, err := s.Store.PredictionTelegramEnabled()
	if err != nil {
		s.Log.Error("prediction telegram setting", "err", err)
		enabled = true
	}
	if !enabled {
		s.Log.Info("prediction telegram disabled in settings")
		return nil
	}

	msg := res.TelegramMessage()
	if err := s.Telegram.Send(msg); err != nil {
		return fmt.Errorf("telegram prediction: %w", err)
	}

	s.Log.Info("prediction telegram sent", "similar_days", res.SimilarDays)
	return nil
}

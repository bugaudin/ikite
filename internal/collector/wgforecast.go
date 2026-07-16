package collector

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/ben/ikite-go/internal/config"
	"github.com/ben/ikite-go/internal/models"
	"github.com/ben/ikite-go/internal/sources/windguru"
	"github.com/ben/ikite-go/internal/store"
)

type WGForecastService struct {
	Cfg    *config.Config
	Store  *store.Store
	WG     *windguru.ForecastClient
	Log    *slog.Logger
}

type WGForecastOptions struct {
	Force bool // skip 7am window and re-fetch even if already stored today
}

func (s *WGForecastService) Run(now time.Time, opts WGForecastOptions) error {
	if s.Cfg.BegetProxyURL == "" {
		return fmt.Errorf("BEGET_PROXY_URL not set")
	}

	now = now.In(s.Cfg.Timezone)
	if !opts.Force && now.Hour() != 7 {
		s.Log.Info("wg forecast skipped", "reason", "not 7am", "hour", now.Hour())
		return nil
	}

	spots, err := s.Store.SpotsWithWindguruForecast()
	if err != nil {
		return err
	}
	if len(spots) == 0 {
		s.Log.Info("wg forecast skipped", "reason", "no spots with windguru_id")
		return nil
	}

	forecastDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, s.Cfg.Timezone)
	fetchedAt := now

	for _, sp := range spots {
		if sp.WindguruID == nil {
			continue
		}
		if !opts.Force {
			exists, err := s.Store.WindForecastAlreadyFetched(*sp.WindguruID, forecastDate)
			if err != nil {
				return err
			}
			if exists {
				s.Log.Info("wg forecast skipped", "spot", sp.ID, "windguru_id", *sp.WindguruID, "reason", "already fetched today")
				continue
			}
		}

		rows, err := s.WG.FetchSpotForecasts(*sp.WindguruID, forecastDate, s.Cfg.Timezone)
		if err != nil {
			return fmt.Errorf("spot %s windguru %d: %w", sp.ID, *sp.WindguruID, err)
		}
		for i := range rows {
			rows[i].Location = sp.ID
			rows[i].WindguruID = *sp.WindguruID
			rows[i].ForecastDate = forecastDate
		}
		if err := s.Store.ReplaceWindForecast(*sp.WindguruID, forecastDate, fetchedAt, rows); err != nil {
			return err
		}
		s.Log.Info("wg forecast saved",
			"spot", sp.ID,
			"windguru_id", *sp.WindguruID,
			"rows", len(rows),
			"models", countModels(rows),
		)
	}
	return nil
}

func countModels(rows []models.WindForecastRow) int {
	seen := map[int]bool{}
	for _, r := range rows {
		seen[r.IDModel] = true
	}
	return len(seen)
}

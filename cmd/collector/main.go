package main

import (
	"flag"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/ben/ikite-go/internal/collector"
	"github.com/ben/ikite-go/internal/config"
	"github.com/ben/ikite-go/internal/db"
	"github.com/ben/ikite-go/internal/notify/telegram"
	"github.com/ben/ikite-go/internal/sources/kyhistory"
	"github.com/ben/ikite-go/internal/sources/windguru"
	"github.com/ben/ikite-go/internal/sources/windometer"
	"github.com/ben/ikite-go/internal/store"
)

func main() {
	wgStation := flag.Int("wg-station", 0, "fetch and save a single Windguru station ID")
	flag.Parse()

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := config.Load()
	if err != nil {
		log.Error("config", "err", err)
		os.Exit(1)
	}

	sqlDB, err := db.Open(cfg.DSN)
	if err != nil {
		log.Error("db open", "err", err)
		os.Exit(1)
	}
	defer sqlDB.Close()

	if mig := os.Getenv("MIGRATE"); mig == "1" || mig == "true" {
		dir := "migrations"
		if v := os.Getenv("MIGRATIONS_PATH"); v != "" {
			dir = v
		}
		if abs, err := filepath.Abs(dir); err == nil {
			dir = abs
		}
		if err := db.MigrateDir(sqlDB, dir); err != nil {
			log.Error("migrate", "err", err)
			os.Exit(1)
		}
		st := store.New(sqlDB)
		if err := st.ApplyLegacySpotSettings(); err != nil {
			log.Error("migrate spot settings", "err", err)
			os.Exit(1)
		}
	}

	svc := &collector.Service{
		Cfg:      cfg,
		Store:    store.New(sqlDB),
		WG:       windguru.New(cfg.BegetWGStationURL),
		KY:       kyhistory.New(cfg.KYHistoryURL),
		KH:       windometer.New(),
		Telegram: telegram.New(cfg.TelegramAlertToken, cfg.TelegramAlertChatID),
		Log:      log,
	}

	now := time.Now()

	if *wgStation > 0 {
		if err := svc.RunWindguruStation(now, *wgStation); err != nil {
			log.Error("windguru station failed", "station", *wgStation, "err", err)
			os.Exit(1)
		}
		log.Info("windguru station done", "station", *wgStation)
		return
	}

	res, err := svc.Run(now)
	if err != nil {
		log.Error("collector failed", "err", err)
		os.Exit(1)
	}
	log.Info("collector done",
		"saved", res.SavedCount,
		"wind_ky", res.WindKY,
		"wind_kh", res.WindKH,
		"alert", res.AlertSent,
	)
}

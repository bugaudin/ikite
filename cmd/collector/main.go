package main

import (
	"flag"
	"log/slog"
	"os"
	"time"

	"github.com/ben/ikite-go/internal/begetproxy"
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

	proxy := begetproxy.New(cfg.BegetProxyURL)

	svc := &collector.Service{
		Cfg:      cfg,
		Store:    store.New(sqlDB),
		WG:       windguru.New(proxy),
		KY:       kyhistory.New(proxy, cfg.KYHistoryURL),
		KH:       windometer.New(proxy, cfg.WindometerLiveURL),
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

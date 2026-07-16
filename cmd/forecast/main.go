package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/ben/ikite-go/internal/begetproxy"
	"github.com/ben/ikite-go/internal/collector"
	"github.com/ben/ikite-go/internal/config"
	"github.com/ben/ikite-go/internal/db"
	"github.com/ben/ikite-go/internal/notify/telegram"
	"github.com/ben/ikite-go/internal/sources/surfo"
	"github.com/ben/ikite-go/internal/store"
	"github.com/ben/ikite-go/internal/translate"
)

func main() {
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

	svc := &collector.ForecastService{
		Cfg:       cfg,
		Store:     store.New(sqlDB),
		Surfo:     surfo.New(begetproxy.New(cfg.BegetProxyURL), cfg.SurfoLiveURL),
		Translate: translate.New(),
		Telegram:  telegram.New(cfg.TelegramAIToken, cfg.TelegramAIChatID),
		Log:       log,
	}

	if err := svc.Run(time.Now()); err != nil {
		log.Error("forecast failed", "err", err)
		os.Exit(1)
	}
}

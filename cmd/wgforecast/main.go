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
	"github.com/ben/ikite-go/internal/sources/windguru"
	"github.com/ben/ikite-go/internal/store"
)

func main() {
	force := flag.Bool("force", false, "run now (ignore 7am schedule and replace today's data)")
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

	svc := &collector.WGForecastService{
		Cfg:   cfg,
		Store: store.New(sqlDB),
		WG:    windguru.NewForecast(begetproxy.New(cfg.BegetProxyURL)),
		Log:   log,
	}

	if err := svc.Run(time.Now(), collector.WGForecastOptions{Force: *force}); err != nil {
		log.Error("wg forecast failed", "err", err)
		os.Exit(1)
	}
}

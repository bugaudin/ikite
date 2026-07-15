package main

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/ben/ikite-go/internal/config"
	"github.com/ben/ikite-go/internal/db"
	"github.com/ben/ikite-go/internal/store"
	"github.com/ben/ikite-go/internal/web"
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

	if mig := os.Getenv("MIGRATE"); mig == "1" || mig == "true" {
		dir := getenv("MIGRATIONS_PATH", "migrations")
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
		log.Info("migrations applied")
	}

	st := store.New(sqlDB)
	srv, err := web.New(cfg, st, log)
	if err != nil {
		log.Error("web init", "err", err)
		os.Exit(1)
	}

	log.Info("listening", "addr", cfg.HTTPAddr)
	if err := http.ListenAndServe(cfg.HTTPAddr, srv.Handler()); err != nil {
		log.Error("server", "err", err)
		os.Exit(1)
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func Open(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return db, nil
}

func Migrate(db *sql.DB, path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return execStatements(db, string(raw))
}

func MigrateDir(db *sql.DB, dir string) error {
	paths, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		return err
	}
	sort.Strings(paths)
	for _, path := range paths {
		if err := Migrate(db, path); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
	}
	return nil
}

func execStatements(db *sql.DB, raw string) error {
	for _, stmt := range splitSQL(raw) {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("migrate statement failed: %w\nSQL: %s", err, stmt)
		}
	}
	return nil
}

func splitSQL(s string) []string {
	parts := strings.Split(s, ";")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

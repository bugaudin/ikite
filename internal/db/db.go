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
	if err := ensureMigrationsTable(db); err != nil {
		return fmt.Errorf("schema_migrations: %w", err)
	}

	paths, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		return err
	}
	sort.Strings(paths)

	if err := seedLegacyMigrations(db, paths); err != nil {
		return fmt.Errorf("seed legacy migrations: %w", err)
	}

	for _, path := range paths {
		name := filepath.Base(path)
		applied, err := migrationApplied(db, name)
		if err != nil {
			return err
		}
		if applied {
			continue
		}
		if err := Migrate(db, path); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		if err := markMigrationApplied(db, name); err != nil {
			return fmt.Errorf("%s: record applied: %w", name, err)
		}
	}
	return nil
}

func ensureMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		name VARCHAR(255) NOT NULL,
		applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (name)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`)
	return err
}

func migrationApplied(db *sql.DB, name string) (bool, error) {
	var n int
	err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE name = ?`, name).Scan(&n)
	return n > 0, err
}

func markMigrationApplied(db *sql.DB, name string) error {
	_, err := db.Exec(`INSERT INTO schema_migrations (name) VALUES (?)`, name)
	return err
}

// seedLegacyMigrations marks migrations as applied on databases that were
// migrated before schema_migrations existed.
func seedLegacyMigrations(db *sql.DB, paths []string) error {
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	legacy := map[string]string{
		"003_spot_collect.sql":  "collect",
		"004_spot_schedule.sql": "collect_interval_min",
	}
	for _, path := range paths {
		name := filepath.Base(path)
		col, ok := legacy[name]
		if !ok {
			continue
		}
		exists, err := columnExists(db, "spots", col)
		if err != nil {
			return err
		}
		if !exists {
			continue
		}
		if err := markMigrationApplied(db, name); err != nil {
			return err
		}
	}

	// If spots exists, 001 and 002 were applied on legacy installs.
	hasSpots, err := tableExists(db, "spots")
	if err != nil {
		return err
	}
	if !hasSpots {
		return nil
	}
	for _, path := range paths {
		name := filepath.Base(path)
		if strings.HasPrefix(name, "001_") || strings.HasPrefix(name, "002_") {
			if err := markMigrationApplied(db, name); err != nil {
				return err
			}
		}
	}
	return nil
}

func tableExists(db *sql.DB, table string) (bool, error) {
	var n int
	err := db.QueryRow(
		`SELECT COUNT(*) FROM information_schema.TABLES
		 WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?`,
		table,
	).Scan(&n)
	return n > 0, err
}

func columnExists(db *sql.DB, table, column string) (bool, error) {
	var n int
	err := db.QueryRow(
		`SELECT COUNT(*) FROM information_schema.COLUMNS
		 WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?`,
		table, column,
	).Scan(&n)
	return n > 0, err
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

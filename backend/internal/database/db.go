package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/user/kareelio/backend/internal/config"
)

func Connect(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	log.Println("Connected to PostgreSQL")
	return pool, nil
}

func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	entries, err := os.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("unable to read migrations directory: %w", err)
	}

	type migration struct {
		name  string
		query string
	}

	var upMigrations []migration
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".up.sql") {
			continue
		}
		content, err := os.ReadFile("migrations/" + name)
		if err != nil {
			return fmt.Errorf("unable to read migration %s: %w", name, err)
		}
		upMigrations = append(upMigrations, migration{name: name, query: string(content)})
	}

	for _, m := range upMigrations {
		log.Printf("Running migration: %s", m.name)
		if _, err := pool.Exec(ctx, m.query); err != nil {
			return fmt.Errorf("migration %s failed: %w", m.name, err)
		}
	}

	return nil
}

func SeedAdmin(ctx context.Context, pool *pgxpool.Pool, email, passwordHash string) error {
	var exists bool
	err := pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE role = 'admin' AND is_active = true)").Scan(&exists)
	if err != nil {
		return fmt.Errorf("unable to check admin: %w", err)
	}
	if exists {
		return nil
	}

	_, err = pool.Exec(ctx,
		`INSERT INTO users (email, display_name, password_hash, role, is_active, email_verified_at, language, theme)
		 VALUES ($1, 'Admin', $2, 'admin', true, NOW(), 'system', 'system')
		 ON CONFLICT (email) DO NOTHING`,
		email, passwordHash,
	)
	if err != nil {
		return fmt.Errorf("unable to seed admin: %w", err)
	}

	log.Println("Default admin user created")
	return nil
}

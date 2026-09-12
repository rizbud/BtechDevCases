package config

import (
	"context"
	"fmt"
	"log"
	"time"

	"btech-wallet/migrations"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	dbPingRetries = 10
	dbPingDelay   = 2 * time.Second
)

func newPool(ctx context.Context, dbURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, err
	}

	if err := waitForDB(ctx, pool); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}

func waitForDB(ctx context.Context, pool *pgxpool.Pool) error {
	var err error
	for i := 1; i <= dbPingRetries; i++ {
		if err = pool.Ping(ctx); err == nil {
			return nil
		}
		log.Printf("[Config.waitForDB] Database not ready (attempt %d/%d): %v", i, dbPingRetries, err)
		time.Sleep(dbPingDelay)
	}
	return fmt.Errorf("database not reachable after %d attempts: %w", dbPingRetries, err)
}

func runMigrations(pool *pgxpool.Pool) error {
	d, err := iofs.New(migrations.Migrations, ".")
	if err != nil {
		log.Printf("[Config.runMigrations] Failed to create migration source: %v", err)
		return fmt.Errorf("failed to create migration source: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, pool.Config().ConnString())
	if err != nil {
		log.Printf("[Config.runMigrations] Failed to create migrate instance: %v", err)
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Printf("[Config.runMigrations] Failed to run migrations: %v", err)
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database migrations applied successfully")
	return nil
}

func InitDB(ctx context.Context, dbURL string) (*pgxpool.Pool, error) {
	pool, err := newPool(ctx, dbURL)
	if err != nil {
		return nil, err
	}

	if err := runMigrations(pool); err != nil {
		return nil, err
	}

	return pool, nil
}

package main

import (
	"context"
	"database/sql"
	"flag"
	"log/slog"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/chhoumann/blogaggregator/internal/database"
	"github.com/chhoumann/blogaggregator/internal/scraper"
	"github.com/joho/godotenv"

	// Import the pq driver - it needs to register itself with the database/sql package
	_ "github.com/lib/pq"
)

type config struct {
	port int
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  time.Duration
	}
}

type application struct {
	config  config
	logger  *slog.Logger
	wg      sync.WaitGroup
	DB      *database.Queries
	scraper *scraper.Scraper
}

func main() {
	godotenv.Load(".env")
	var cfg config

	flag.IntVar(&cfg.port, "port", getEnvAsInt("PORT", 4000), "Server port to listen on")
	flag.StringVar(&cfg.db.dsn, "db-dsn", getEnvAsString("DATABASE_URL", ""), "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", getEnvAsInt("DB_MAX_OPEN_CONNS", 25), "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", getEnvAsInt("DB_MAX_IDLE_CONNS", 25), "PostgreSQL max idle connections")
	flag.DurationVar(&cfg.db.maxIdleTime, "db-max-idle-time", getEnvAsDuration("DB_MAX_IDLE_TIME", 15*time.Minute), "PostgreSQL max idle time")

	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	db, err := openDB(cfg)
	if err != nil {
		logger.Error("error opening db", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	logger.Info("database connection pool established")

	dbQueries := database.New(db)

	scraperLogger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	scraper := scraper.NewScraper(dbQueries, scraperLogger)

	app := &application{
		config:  cfg,
		logger:  logger,
		DB:      dbQueries,
		scraper: scraper,
	}

	// Start the scraper in a separate goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go app.scraper.Run(ctx, 1*time.Minute, 10) // Run every 1 minute, fetch 10 feeds at a time

	if err := app.serve(); err != nil {
		logger.Error("error starting server", "error", err)
		os.Exit(1)
	}
}

func getEnvAsString(key string, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return fallback
}

func getEnvAsFloat64(key string, fallback float64) float64 {
	if value, exists := os.LookupEnv(key); exists {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return fallback
}

func getEnvAsBool(key string, fallback bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return fallback
}

func getEnvAsDuration(key string, fallback time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return fallback
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)
	db.SetConnMaxIdleTime(cfg.db.maxIdleTime)

	// Create context with 5s timeout. If we can't connect in 5s, we cancel the context and return an error.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

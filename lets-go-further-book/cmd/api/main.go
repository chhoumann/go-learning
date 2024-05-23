package main

import (
	"context"
	"database/sql"
	"flag"
	"log/slog"
	"os"
	"time"

	"greenlight.bagerbach.com/internal/data"

	// Import the pq driver - it needs to register itself with the database/sql package
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// Will be generated automatically later.
const version = "1.0.0"

type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  time.Duration
	}
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
}

type application struct {
	config config
	logger *slog.Logger
	models data.Models
}

func main() {
	err := godotenv.Load()
	if err != nil {
		slog.Error("error loading .env file", "error", err)
		os.Exit(1)
	}

	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "Server port to listen on")
	flag.StringVar(&cfg.env, "env", "development", "Application environment (development|staging|production)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("DATABASE_URL"), "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "Maximum number of open connections to the database")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "Maximum number of idle connections to the database")
	// DurationVar lets us pass in any value acceptable to time.ParseDuration(), e.g. 300ms, 5s, 2h45m.
	flag.DurationVar(&cfg.db.maxIdleTime, "db-max-idle-time", 15*time.Minute, "Maximum idle time for a connection to the database")
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limit to apply to requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Burst limit to apply to requests")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiting")

	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	db, err := openDB(cfg)
	if err != nil {
		logger.Error("error opening db", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	logger.Info("database connection pool established")

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
	}

	if err := app.serve(); err != nil {
		logger.Error("error starting server", "error", err)
		os.Exit(1)
	}
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

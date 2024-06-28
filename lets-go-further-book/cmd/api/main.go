package main

import (
	"context"
	"database/sql"
	"expvar"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"greenlight.bagerbach.com/internal/data"
	"greenlight.bagerbach.com/internal/mailer"
	"greenlight.bagerbach.com/internal/vcs"

	// Import the pq driver - it needs to register itself with the database/sql package
	_ "github.com/lib/pq"
)

var (
	version = vcs.Version()
)

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
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
	cors struct {
		trustedOrigins []string
	}
}

type application struct {
	config config
	logger *slog.Logger
	models data.Models
	mailer mailer.Mailer
	wg     sync.WaitGroup
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", getEnvAsInt("PORT", 4000), "Server port to listen on")
	flag.StringVar(&cfg.env, "env", getEnvAsString("ENV", "development"), "Application environment (development|staging|production)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", getEnvAsString("DATABASE_URL", ""), "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", getEnvAsInt("DB_MAX_OPEN_CONNS", 25), "Maximum number of open connections to the database")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", getEnvAsInt("DB_MAX_IDLE_CONNS", 25), "Maximum number of idle connections to the database")
	flag.DurationVar(&cfg.db.maxIdleTime, "db-max-idle-time", getEnvAsDuration("DB_MAX_IDLE_TIME", 15*time.Minute), "Maximum idle time for a connection to the database")
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", getEnvAsFloat64("LIMITER_RPS", 2), "Rate limit to apply to requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", getEnvAsInt("LIMITER_BURST", 4), "Burst limit to apply to requests")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", getEnvAsBool("LIMITER_ENABLED", true), "Enable rate limiting")
	flag.StringVar(&cfg.smtp.host, "smtp-host", getEnvAsString("SMTP_HOST", ""), "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", getEnvAsInt("SMTP_PORT", 25), "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", getEnvAsString("SMTP_USERNAME", ""), "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", getEnvAsString("SMTP_PASSWORD", ""), "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", getEnvAsString("SMTP_SENDER", ""), "SMTP sender")
	flag.StringVar(&version, "version", getEnvAsString("VERSION", "unknown"), "Version")

	flag.Func("cors-trusted-origins", "Trusted CORS origins (space separated)", func(s string) error {
		cfg.cors.trustedOrigins = strings.Fields(s)
		return nil
	})

	displayVersion := flag.Bool("version", false, "Display version and exit")

	flag.Parse()

	if *displayVersion {
		fmt.Printf("Version: %s\n", version)
		os.Exit(0)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	db, err := openDB(cfg)
	if err != nil {
		logger.Error("error opening db", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	logger.Info("database connection pool established")

	expvar.NewString("version").Set(version)
	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))
	expvar.Publish("database", expvar.Func(func() any {
		return db.Stats()
	}))
	expvar.Publish("timestamp", expvar.Func(func() any {
		return time.Now().Unix()
	}))

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
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
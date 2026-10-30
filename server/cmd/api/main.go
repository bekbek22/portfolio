// Command api runs the portfolio HTTP API.
//
// Usage:
//
//	api [serve]      start the HTTP server (default)
//	api healthcheck  exit 0 if the local server reports healthy (for container HEALTHCHECK)
//	api migrate up | down [n] | version
//	                 apply, roll back (default 1 step) or inspect the embedded SQL migrations
//	api seed         load the placeholder projects and tags (safe to re-run)
package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/joho/godotenv"

	"github.com/bekbek101/portfolio/server/internal/config"
	"github.com/bekbek101/portfolio/server/internal/database"
	"github.com/bekbek101/portfolio/server/internal/database/sqlc"
	"github.com/bekbek101/portfolio/server/internal/httpapi"
	"github.com/bekbek101/portfolio/server/internal/project"
	"github.com/bekbek101/portfolio/server/internal/seed"
)

const shutdownTimeout = 10 * time.Second

func main() {
	// A local .env is optional; real environments set variables directly.
	if err := godotenv.Load(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		fmt.Fprintln(os.Stderr, "load .env:", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, os.Args[1:], os.Getenv); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, getenv func(string) string) error {
	cfg, err := config.Load(getenv)
	if err != nil {
		return err
	}

	cmd := "serve"
	if len(args) > 0 {
		cmd = args[0]
	}
	switch cmd {
	case "serve":
		return serve(ctx, cfg, newLogger(cfg.Env))
	case "healthcheck":
		return healthcheck(ctx, cfg.Port)
	case "migrate":
		return runMigrate(cfg.DatabaseURL, args[1:])
	case "seed":
		return runSeed(ctx, cfg.DatabaseURL)
	default:
		return fmt.Errorf("unknown command %q (want: serve, healthcheck, migrate, seed)", cmd)
	}
}

func newLogger(env config.Env) *slog.Logger {
	if env == config.EnvProduction {
		return slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

func serve(ctx context.Context, cfg config.Config, log *slog.Logger) error {
	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	httpCfg := httpapi.DefaultConfig()
	httpCfg.TrustedProxyHops = cfg.TrustedProxyHops
	app, err := httpapi.New(httpapi.Deps{
		DB:       pool,
		Projects: project.NewService(sqlc.New(pool)),
		Logger:   log,
	}, httpCfg)
	if err != nil {
		return err
	}

	listenErr := make(chan error, 1)
	go func() {
		addr := fmt.Sprintf(":%d", cfg.Port)
		log.Info("server starting", "addr", addr, "env", cfg.Env)
		listenErr <- app.Listen(addr, fiber.ListenConfig{DisableStartupMessage: true})
	}()

	select {
	case err := <-listenErr:
		return fmt.Errorf("listen: %w", err)
	case <-ctx.Done():
	}

	log.Info("shutdown signal received, draining connections", "timeout", shutdownTimeout.String())
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	log.Info("server stopped")
	return nil
}

func healthcheck(ctx context.Context, port int) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	url := fmt.Sprintf("http://127.0.0.1:%d/api/v1/health", port)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("healthcheck: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("healthcheck: status %d", res.StatusCode)
	}
	return nil
}

func runMigrate(databaseURL string, args []string) (err error) {
	if len(args) == 0 {
		return errors.New("usage: api migrate up | down [n] | version")
	}

	mg, err := database.NewMigrator(databaseURL)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, mg.Close()) }()

	switch args[0] {
	case "up":
		err = mg.Up()
	case "down":
		steps := 1
		if len(args) > 1 {
			if steps, err = strconv.Atoi(args[1]); err != nil {
				return fmt.Errorf("migrate down: invalid step count %q", args[1])
			}
		}
		err = mg.Down(steps)
	case "version":
	default:
		return fmt.Errorf("unknown migrate command %q (want: up, down, version)", args[0])
	}
	if err != nil {
		return err
	}

	version, dirty, err := mg.Version()
	if err != nil {
		return fmt.Errorf("read migration version: %w", err)
	}
	fmt.Printf("schema version: %d (dirty: %t)\n", version, dirty)
	return nil
}

func runSeed(ctx context.Context, databaseURL string) error {
	pool, err := database.NewPool(ctx, databaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	if err := seed.Run(ctx, pool); err != nil {
		return err
	}
	fmt.Println("seed complete")
	return nil
}

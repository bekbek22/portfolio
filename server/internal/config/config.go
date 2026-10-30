// Package config loads and validates runtime configuration from the environment.
package config

import (
	"errors"
	"fmt"
	"strconv"
)

type Env string

const (
	EnvDevelopment Env = "development"
	EnvProduction  Env = "production"
)

type Config struct {
	Env         Env
	Port        int
	DatabaseURL string
	// TrustedProxyHops is the number of reverse proxies in front of the server
	// (0 = use the socket address as the client IP).
	TrustedProxyHops int
}

// Load reads configuration through getenv (os.Getenv in production, a map in tests)
// and reports every invalid value at once.
func Load(getenv func(string) string) (Config, error) {
	cfg := Config{Env: EnvDevelopment, Port: 4000}
	var errs []error

	switch env := Env(getenv("APP_ENV")); env {
	case "":
	case EnvDevelopment, EnvProduction:
		cfg.Env = env
	default:
		errs = append(errs, fmt.Errorf("APP_ENV: must be %q or %q, got %q", EnvDevelopment, EnvProduction, env))
	}

	if raw := getenv("PORT"); raw != "" {
		port, err := strconv.Atoi(raw)
		if err != nil || port < 1 || port > 65535 {
			errs = append(errs, fmt.Errorf("PORT: must be 1-65535, got %q", raw))
		}
		cfg.Port = port
	}

	cfg.DatabaseURL = getenv("DATABASE_URL")
	if cfg.DatabaseURL == "" {
		errs = append(errs, errors.New("DATABASE_URL: required"))
	}

	if raw := getenv("TRUSTED_PROXY_HOPS"); raw != "" {
		hops, err := strconv.Atoi(raw)
		if err != nil || hops < 0 {
			errs = append(errs, fmt.Errorf("TRUSTED_PROXY_HOPS: must be >= 0, got %q", raw))
		}
		cfg.TrustedProxyHops = hops
	}

	if err := errors.Join(errs...); err != nil {
		return Config{}, fmt.Errorf("invalid configuration:\n%w", err)
	}
	return cfg, nil
}

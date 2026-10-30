// Package httpapi wires the Fiber application: middleware, validation and the handlers
// generated from api/openapi.yaml.
package httpapi

import (
	"fmt"
	"log/slog"
	"runtime/debug"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"

	"github.com/bekbek101/portfolio/server/internal/httpapi/gen"
)

const basePath = "/api/v1"

type Deps struct {
	DB     Pinger
	Logger *slog.Logger
}

type Config struct {
	TrustedProxyHops int           // reverse proxies in front of the server
	RequestTimeout   time.Duration // deadline for each request's context
	RateLimitMax     int           // requests per client IP per window
	RateLimitWindow  time.Duration
}

func DefaultConfig() Config {
	return Config{
		RequestTimeout:  5 * time.Second,
		RateLimitMax:    120,
		RateLimitWindow: time.Minute,
	}
}

// New builds the application. Middleware order:
//
//	request ID → logger (renders errors) → recover → deadline → client IP
//	→ [/api/v1] rate limit → OpenAPI validation → generated strict handlers
func New(deps Deps, cfg Config) (*fiber.App, error) {
	log := deps.Logger
	onError := errorHandler(log)

	app := fiber.New(fiber.Config{AppName: "portfolio-api", ErrorHandler: onError})
	app.Use(
		requestid.New(),
		requestLogger(log, onError),
		recover.New(recover.Config{
			EnableStackTrace: true,
			StackTraceHandler: func(c fiber.Ctx, e any) {
				log.ErrorContext(c.Context(), "panic recovered", "panic", e, "stack", string(debug.Stack()))
			},
		}),
		requestDeadline(cfg.RequestTimeout),
		clientIP(cfg.TrustedProxyHops),
	)

	spec, err := gen.GetSpec()
	if err != nil {
		return nil, fmt.Errorf("load embedded OpenAPI spec: %w", err)
	}
	validatorSpec, err := gen.GetSpec() // the validator mutates its copy
	if err != nil {
		return nil, fmt.Errorf("load embedded OpenAPI spec: %w", err)
	}
	validate, err := openAPIValidator(validatorSpec, basePath)
	if err != nil {
		return nil, err
	}

	v1 := app.Group(basePath)
	v1.Use(
		limiter.New(limiter.Config{
			Max:          cfg.RateLimitMax,
			Expiration:   cfg.RateLimitWindow,
			KeyGenerator: ClientIP,
			LimitReached: func(fiber.Ctx) error { return errRateLimited },
			// Container and platform health probes must never be throttled.
			Next: func(c fiber.Ctx) bool { return c.Path() == basePath+"/health" },
		}),
		validate,
	)
	v1.Get("/openapi.json", func(c fiber.Ctx) error { return c.JSON(spec) })
	gen.RegisterHandlers(v1, gen.NewStrictHandler(&handlers{db: deps.DB}, nil))

	return app, nil
}

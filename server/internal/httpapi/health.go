package httpapi

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3"
)

// Pinger is the slice of the database pool the health check needs.
type Pinger interface {
	Ping(ctx context.Context) error
}

type healthResponse struct {
	Status string `json:"status"`
	DB     string `json:"db"`
}

const healthDBTimeout = 2 * time.Second

func healthHandler(db Pinger) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Fiber's request context carries no deadline, so bound the DB call explicitly.
		ctx, cancel := context.WithTimeout(c.Context(), healthDBTimeout)
		defer cancel()

		if err := db.Ping(ctx); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(healthResponse{Status: "degraded", DB: "down"})
		}
		return c.JSON(healthResponse{Status: "ok", DB: "up"})
	}
}

// Package httpapi wires the Fiber application: routes and middleware.
package httpapi

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

type Deps struct {
	DB Pinger
}

func New(deps Deps) *fiber.App {
	app := fiber.New(fiber.Config{AppName: "portfolio-api"})
	app.Use(recover.New())

	v1 := app.Group("/api/v1")
	v1.Get("/health", healthHandler(deps.DB))

	return app
}

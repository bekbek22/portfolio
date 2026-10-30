package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
)

// requestLogger logs one line per request. It renders errors itself (through onError)
// so the logged status is the one the client actually receives.
func requestLogger(log *slog.Logger, onError fiber.ErrorHandler) fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()

		if err := c.Next(); err != nil {
			if hErr := onError(c, err); hErr != nil {
				c.Status(http.StatusInternalServerError)
			}
		}

		status := c.Response().StatusCode()
		level := slog.LevelInfo
		if status >= http.StatusInternalServerError {
			level = slog.LevelError
		}
		log.LogAttrs(context.Background(), level, "request",
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.Int("status", status),
			slog.Duration("duration", time.Since(start)),
			slog.String("request_id", requestid.FromContext(c)),
			slog.String("client_ip", ClientIP(c)),
		)
		return nil
	}
}

// requestDeadline bounds every request. Fiber's request context is a bare Background
// (no deadline, no cancellation on disconnect), so handlers would otherwise wait on the
// database forever.
func requestDeadline(timeout time.Duration) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.Context(), timeout)
		defer cancel()
		c.SetContext(ctx)
		return c.Next()
	}
}

type localsKey int

const clientIPKey localsKey = iota

// clientIP resolves the caller's address once per request; read it with ClientIP.
func clientIP(trustedHops int) fiber.Handler {
	return func(c fiber.Ctx) error {
		ip := resolveClientIP(c.Get(fiber.HeaderXForwardedFor), c.RequestCtx().RemoteIP().String(), trustedHops)
		// Fiber strings point into fasthttp's reusable buffers; the rate limiter keeps this
		// value as a map key beyond the request, so it must own its memory.
		c.Locals(clientIPKey, strings.Clone(ip))
		return c.Next()
	}
}

// ClientIP returns the address resolved by the clientIP middleware.
func ClientIP(c fiber.Ctx) string {
	ip, _ := c.Locals(clientIPKey).(string)
	return ip
}

// resolveClientIP picks the caller's IP. Each proxy appends the address it received the
// request from to X-Forwarded-For, so with trustedHops proxies in front of us the client
// is trustedHops entries from the right. Entries further left are client-controlled.
func resolveClientIP(forwardedFor, remoteIP string, trustedHops int) string {
	if trustedHops > 0 && forwardedFor != "" {
		hops := strings.Split(forwardedFor, ",")
		if i := len(hops) - trustedHops; i >= 0 {
			if ip := strings.TrimSpace(hops[i]); ip != "" {
				return ip
			}
		}
	}
	return remoteIP
}

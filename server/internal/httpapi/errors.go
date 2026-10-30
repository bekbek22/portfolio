package httpapi

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gofiber/fiber/v3"

	"github.com/bekbek101/portfolio/server/internal/httpapi/gen"
)

// Sentinel errors handlers return; errorHandler maps them to HTTP responses.
var (
	ErrNotFound       = errors.New("resource not found")
	ErrNotImplemented = errors.New("not implemented")
	errRateLimited    = errors.New("rate limit exceeded")
)

// errorHandler renders every error as the API's Error envelope. Unexpected errors are
// logged with their detail and returned as a generic 500 so internals never leak.
// (Every request, including expected errors, is logged once by requestLogger.)
func errorHandler(log *slog.Logger) fiber.ErrorHandler {
	return func(c fiber.Ctx, err error) error {
		status, detail := toErrorDetail(err)
		if detail.Code == gen.ErrorCodeInternal {
			log.ErrorContext(c.Context(), "request failed", "error", err, "path", c.Path())
		}
		return c.Status(status).JSON(gen.Error{Error: detail})
	}
}

func toErrorDetail(err error) (int, gen.ErrorDetail) {
	var vErr *validationError
	var fErr *fiber.Error

	switch {
	case errors.As(err, &vErr):
		if vErr.malformed {
			return http.StatusBadRequest, gen.ErrorDetail{Code: gen.ErrorCodeValidation, Message: vErr.message}
		}
		return http.StatusUnprocessableEntity, gen.ErrorDetail{
			Code: gen.ErrorCodeValidation, Message: "Invalid request", FieldErrors: &vErr.fieldErrors,
		}
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound, gen.ErrorDetail{Code: gen.ErrorCodeNotFound, Message: "Resource not found"}
	case errors.Is(err, ErrNotImplemented):
		return http.StatusNotImplemented, gen.ErrorDetail{Code: gen.ErrorCodeNotImplemented, Message: "Not implemented yet"}
	case errors.Is(err, errRateLimited):
		return http.StatusTooManyRequests, gen.ErrorDetail{Code: gen.ErrorCodeRateLimited, Message: "Too many requests, try again later"}
	case errors.As(err, &fErr):
		switch {
		case fErr.Code == http.StatusNotFound:
			return http.StatusNotFound, gen.ErrorDetail{Code: gen.ErrorCodeNotFound, Message: "Route not found"}
		case fErr.Code == http.StatusMethodNotAllowed:
			return http.StatusMethodNotAllowed, gen.ErrorDetail{Code: gen.ErrorCodeMethodNotAllowed, Message: "Method not allowed"}
		case fErr.Code < http.StatusInternalServerError:
			return fErr.Code, gen.ErrorDetail{Code: gen.ErrorCodeValidation, Message: fErr.Message}
		}
	}
	return http.StatusInternalServerError, gen.ErrorDetail{Code: gen.ErrorCodeInternal, Message: "Internal server error"}
}

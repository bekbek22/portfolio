package httpapi

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
)

// validationError is a request that failed OpenAPI validation.
type validationError struct {
	malformed   bool                // 400: the request could not be interpreted at all
	message     string              // used when malformed
	fieldErrors map[string][]string // 422: messages per parameter name
}

func (e *validationError) Error() string {
	if e.malformed {
		return "malformed request: " + e.message
	}
	return fmt.Sprintf("invalid request: %v", e.fieldErrors)
}

// openAPIValidator validates requests under basePath against the spec before they reach
// the generated handlers. Requests for paths or methods the spec doesn't define are
// passed through so Fiber answers them with 404/405.
func openAPIValidator(doc *openapi3.T, basePath string) (fiber.Handler, error) {
	// Routes are matched on the path below basePath, so ignore the spec's servers.
	doc.Servers = nil
	router, err := gorillamux.NewRouter(doc)
	if err != nil {
		return nil, fmt.Errorf("build OpenAPI router: %w", err)
	}
	opts := &openapi3filter.Options{MultiError: true}

	return func(c fiber.Ctx) error {
		req, err := adaptor.ConvertRequest(c, false)
		if err != nil {
			return &validationError{malformed: true, message: "could not read request"}
		}
		req.URL.Path = strings.TrimPrefix(req.URL.Path, basePath)

		route, pathParams, err := router.FindRoute(req)
		if err != nil {
			if errors.Is(err, routers.ErrPathNotFound) || errors.Is(err, routers.ErrMethodNotAllowed) {
				return c.Next()
			}
			return fmt.Errorf("find OpenAPI route: %w", err)
		}

		err = openapi3filter.ValidateRequest(context.Background(), &openapi3filter.RequestValidationInput{
			Request:    req,
			PathParams: pathParams,
			Route:      route,
			Options:    opts,
		})
		if err != nil {
			return toValidationError(err)
		}
		return c.Next()
	}, nil
}

// toValidationError groups kin-openapi's errors by parameter name.
func toValidationError(err error) error {
	var all openapi3.MultiError
	if !errors.As(err, &all) {
		all = openapi3.MultiError{err}
	}

	fieldErrors := map[string][]string{}
	for _, e := range all {
		var reqErr *openapi3filter.RequestError
		if !errors.As(e, &reqErr) || reqErr.Parameter == nil {
			// Not tied to a parameter (e.g. an unreadable body): reject as malformed.
			return &validationError{malformed: true, message: e.Error()}
		}
		name := reqErr.Parameter.Name
		fieldErrors[name] = append(fieldErrors[name], parameterMessage(reqErr))
	}
	return &validationError{fieldErrors: fieldErrors}
}

// parameterMessage returns the most specific human-readable reason kin-openapi gives.
func parameterMessage(reqErr *openapi3filter.RequestError) string {
	var schemaErr *openapi3.SchemaError
	if errors.As(reqErr.Err, &schemaErr) {
		return schemaErr.Reason
	}
	var parseErr *openapi3filter.ParseError
	if errors.As(reqErr.Err, &parseErr) {
		return parseErr.Reason
	}
	if reqErr.Reason != "" {
		return reqErr.Reason
	}
	return reqErr.Error()
}

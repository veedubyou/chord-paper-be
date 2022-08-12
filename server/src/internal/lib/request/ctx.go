package request

import (
	"context"
	"github.com/cockroachdb/errors"
	"github.com/labstack/echo/v4"
	"github.com/veedubyou/chord-paper-be/server/src/internal/errors/api"
	"github.com/veedubyou/chord-paper-be/server/src/internal/errors/auth"
	"github.com/veedubyou/chord-paper-be/shared/lib/env"
)

func Context(c echo.Context) context.Context {
	switch env.Get() {
	case env.Production, env.Test:
		return c.Request().Context()

	case env.Development:
		// opt to not use the request context in development situations
		// to avoid timeouts during debugging
		return context.Background()

	default:
		panic("Unrecognized environment")
	}
}

func AuthHeader(c echo.Context) (string, *api.Error) {
	header := c.Request().Header.Get("authorization")
	if header == "" {
		err := errors.New("No authorization header found")
		return "", api.CommitError(err,
			auth.BadAuthorizationHeaderCode,
			"The request is unauthorized. Please try logging in again")
	}

	return header, nil
}

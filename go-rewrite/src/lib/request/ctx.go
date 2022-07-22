package request

import (
	"context"
	"github.com/labstack/echo/v4"
	"github.com/veedubyou/chord-paper-be/go-rewrite/src/lib/env"
)

func Context(c echo.Context) context.Context {
	switch env.Get() {
	case env.Production:
		return c.Request().Context()

	case env.Development:
		// opt to not use the request context in development situations
		// to avoid timeouts during debugging
		return context.Background()

	default:
		panic("Unrecognized environment")
	}
}

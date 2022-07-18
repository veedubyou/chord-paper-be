package usergateway

import (
	"github.com/cockroachdb/errors/markers"
	"github.com/labstack/echo/v4"
	"github.com/veedubyou/chord-paper-be/go-rewrite/src/gateway_errors"
	userusecase "github.com/veedubyou/chord-paper-be/go-rewrite/src/user/usecase"
	"net/http"
)

type Gateway struct {
	usecase userusecase.Usecase
}

func NewGateway(usecase userusecase.Usecase) Gateway {
	return Gateway{
		usecase: usecase,
	}
}

func (g Gateway) Login(c echo.Context) error {
	authHeader := c.Request().Header.Get("authorization")
	user, err := g.usecase.Login(c.Request().Context(), authHeader)
	if err != nil {
		switch {
		case markers.Is(err, userusecase.BadAuthorizationHeaderMark):
			return gateway_errors.ErrorResponse(c, NewBadAuthorizationHeaderError(err))
		case markers.Is(err, userusecase.NoAccountMark):
			return gateway_errors.ErrorResponse(c, NewNoAccountError(err))
		case markers.Is(err, userusecase.NotAuthorizedMark):
			return gateway_errors.ErrorResponse(c, NewNotAuthorizedError(err))
		default:
			return gateway_errors.ErrorResponse(c, gateway_errors.NewInternalError(err))
		}
	}

	userJSON := userJSON{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}

	return c.JSON(http.StatusOK, userJSON)
}

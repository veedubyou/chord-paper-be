package usergateway

import (
	"github.com/labstack/echo/v4"
	"github.com/veedubyou/chord-paper-be/src/server/internal/errors/gateway"
	"github.com/veedubyou/chord-paper-be/src/server/internal/lib/request"
	"github.com/veedubyou/chord-paper-be/src/server/internal/user/usecase"
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
	ctx := request.Context(c)

	authHeader := c.Request().Header.Get("authorization")
	user, apiErr := g.usecase.Login(ctx, authHeader)
	if apiErr != nil {
		return gateway.ErrorResponse(c, apiErr)
	}

	userJSON := UserJSON{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}

	return c.JSON(http.StatusOK, userJSON)
}

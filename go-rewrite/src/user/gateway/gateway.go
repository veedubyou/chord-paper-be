package usergateway

import (
	"github.com/labstack/echo/v4"
	"github.com/veedubyou/chord-paper-be/go-rewrite/src/errors/gateway"
	"github.com/veedubyou/chord-paper-be/go-rewrite/src/lib/request"
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
	ctx := request.Context(c)

	authHeader := c.Request().Header.Get("authorization")
	user, apiErr := g.usecase.Login(ctx, authHeader)
	if apiErr != nil {
		return gateway.ErrorResponse(c, apiErr)
	}

	userJSON := userJSON{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}

	return c.JSON(http.StatusOK, userJSON)
}

package trackgateway

import (
	"github.com/labstack/echo/v4"
	"github.com/veedubyou/chord-paper-be/go-rewrite/src/errors/api"
	"github.com/veedubyou/chord-paper-be/go-rewrite/src/errors/gateway"
	"github.com/veedubyou/chord-paper-be/go-rewrite/src/lib/request"
	trackusecase "github.com/veedubyou/chord-paper-be/go-rewrite/src/track/usecase"
	"net/http"
)

type Gateway struct {
	usecase trackusecase.Usecase
}

func NewGateway(usecase trackusecase.Usecase) Gateway {
	return Gateway{
		usecase: usecase,
	}
}

func (g Gateway) GetTrackList(c echo.Context, songID string) error {
	ctx := request.Context(c)

	tracklist, apiErr := g.usecase.GetTrackList(ctx, songID)
	if apiErr != nil {
		apiErr = api.WrapError(apiErr, "Failed to get track list")
		return gateway.ErrorResponse(c, apiErr)
	}

	return c.JSON(http.StatusOK, tracklist)
}

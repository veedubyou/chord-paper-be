package trackgateway

import (
	"github.com/cockroachdb/errors/markers"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/veedubyou/chord-paper-be/go-rewrite/src/gateway_errors"
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

func (g Gateway) GetTrackList(c echo.Context, songIDStr string) error {
	songID, err := uuid.Parse(songIDStr)
	if err != nil {
		return gateway_errors.ErrorResponse(c, NewInvalidIDError(err))
	}

	tracklist, err := g.usecase.GetTrackList(c.Request().Context(), songID)
	if err != nil {
		switch {
		case markers.Is(err, trackusecase.DefaultErrorMark):
		default:
			return gateway_errors.ErrorResponse(c, gateway_errors.NewInternalError(err))
		}
	}

	return c.JSON(http.StatusOK, tracklist)
}

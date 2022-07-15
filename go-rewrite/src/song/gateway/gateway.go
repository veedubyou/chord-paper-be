package songgateway

import (
	"github.com/google/uuid"
	"github.com/guregu/dynamo"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/veedubyou/chord-paper-be/go-rewrite/src/gateway_errors"
	songusecase "github.com/veedubyou/chord-paper-be/go-rewrite/src/song/usecase"
	"net/http"
)

type Gateway struct {
	usecase songusecase.Usecase
}

func NewGateway(usecase songusecase.Usecase) Gateway {
	return Gateway{
		usecase: usecase,
	}
}

func (g Gateway) GetSong(c echo.Context, songIDStr string) error {
	songID, err := uuid.Parse(songIDStr)
	if err != nil {
		err = errors.Wrap(err, "Failed to parse song ID")
		gatewayErr := gateway_errors.NewInvalidIDError(err)
		return gateway_errors.ErrorResponse(c, gatewayErr)
	}

	song, err := g.usecase.GetSong(c.Request().Context(), songID)
	if err != nil {
		err = errors.Wrap(err, "Failed to get song")

		if errors.Is(err, dynamo.ErrNotFound) {
			gatewayErr := gateway_errors.NewSongNotFoundError(err)
			return gateway_errors.ErrorResponse(c, gatewayErr)
		}

		gatewayErr := gateway_errors.NewInternalError(err)
		return gateway_errors.ErrorResponse(c, gatewayErr)
	}

	return c.JSON(http.StatusOK, song)
}

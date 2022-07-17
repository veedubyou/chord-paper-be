package songgateway

import (
	"github.com/cockroachdb/errors/markers"
	"github.com/google/uuid"
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
		return gateway_errors.ErrorResponse(c, NewInvalidIDError(err))
	}

	song, err := g.usecase.GetSong(c.Request().Context(), songID)
	if err != nil {
		switch {
		case markers.Is(err, songusecase.SongNotFoundMark):
			return gateway_errors.ErrorResponse(c, NewSongNotFoundError(err))

		case markers.Is(err, songusecase.DefaultErrorMark):
		default:
			return gateway_errors.ErrorResponse(c, gateway_errors.NewInternalError(err))
		}
	}

	return c.JSON(http.StatusOK, song)
}

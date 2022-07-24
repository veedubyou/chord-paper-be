package trackgateway

import (
	"github.com/cockroachdb/errors"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/veedubyou/chord-paper-be/go-rewrite/src/errors/api"
	"github.com/veedubyou/chord-paper-be/go-rewrite/src/errors/gateway"
	"github.com/veedubyou/chord-paper-be/go-rewrite/src/lib/request"
	songerrors "github.com/veedubyou/chord-paper-be/go-rewrite/src/song/errors"
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
	ctx := request.Context(c)

	songID, err := uuid.Parse(songIDStr)
	if err != nil {
		err = errors.Wrap(err, "Failed to parse song ID for track list")
		apiErr := api.CommitError(err,
			songerrors.InvalidIDCode,
			"The song ID provided for the track list is invalid")
		return gateway.ErrorResponse(c, apiErr)
	}

	tracklist, apiErr := g.usecase.GetTrackList(ctx, songID)
	if apiErr != nil {
		apiErr = api.WrapError(apiErr, "Failed to get track list")
		return gateway.ErrorResponse(c, apiErr)
	}

	return c.JSON(http.StatusOK, tracklist)
}

package trackgateway

import (
	"github.com/cockroachdb/errors"
	"github.com/labstack/echo/v4"
	"github.com/veedubyou/chord-paper-be/src/server/internal/errors/api"
	"github.com/veedubyou/chord-paper-be/src/server/internal/errors/gateway"
	"github.com/veedubyou/chord-paper-be/src/server/internal/lib/request"
	"github.com/veedubyou/chord-paper-be/src/server/internal/track/errors"
	"github.com/veedubyou/chord-paper-be/src/server/internal/track/usecase"
	"github.com/veedubyou/chord-paper-be/src/shared/track/entity"
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

func (g Gateway) SetTrackList(c echo.Context, songID string) error {
	ctx := request.Context(c)
	authHeader, apiErr := request.AuthHeader(c)
	if apiErr != nil {
		return gateway.ErrorResponse(c, apiErr)
	}

	tracklist := trackentity.TrackList{}
	err := c.Bind(&tracklist)
	if err != nil {
		err = errors.Wrap(err, "Failed to bind request body to tracklist object")
		apiErr := api.CommitError(err,
			trackerrors.BadTracklistDataCode,
			"The tracklist data received was malformed. Please contact the developer")
		return gateway.ErrorResponse(c, apiErr)
	}

	newTracklist, apiErr := g.usecase.SetTrackList(ctx, authHeader, songID, tracklist)
	if apiErr != nil {
		return gateway.ErrorResponse(c, apiErr)
	}

	return c.JSON(http.StatusOK, newTracklist)
}

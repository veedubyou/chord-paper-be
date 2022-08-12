package songgateway

import (
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/veedubyou/chord-paper-be/src/server/internal/errors/api"
	"github.com/veedubyou/chord-paper-be/src/server/internal/errors/gateway"
	"github.com/veedubyou/chord-paper-be/src/server/internal/lib/request"
	"github.com/veedubyou/chord-paper-be/src/server/internal/song/entity"
	"github.com/veedubyou/chord-paper-be/src/server/internal/song/errors"
	"github.com/veedubyou/chord-paper-be/src/server/internal/song/usecase"
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

func (g Gateway) GetSong(c echo.Context, songID string) error {
	ctx := request.Context(c)

	song, apiErr := g.usecase.GetSong(ctx, songID)
	if apiErr != nil {
		return gateway.ErrorResponse(c, apiErr)
	}

	return c.JSON(http.StatusOK, song)
}

func (g Gateway) GetSongSummariesForUser(c echo.Context, ownerID string) error {
	ctx := request.Context(c)

	authHeader, apiErr := request.AuthHeader(c)
	if apiErr != nil {
		return gateway.ErrorResponse(c, apiErr)
	}

	songSummaries, apiErr := g.usecase.GetSongSummariesForUser(ctx, authHeader, ownerID)
	if apiErr != nil {
		return gateway.ErrorResponse(c, apiErr)
	}

	return c.JSON(http.StatusOK, songSummaries)
}

func (g Gateway) CreateSong(c echo.Context) error {
	ctx := request.Context(c)

	authHeader, apiErr := request.AuthHeader(c)
	if apiErr != nil {
		return gateway.ErrorResponse(c, apiErr)
	}

	song := songentity.Song{}
	err := c.Bind(&song)
	if err != nil {
		err = errors.Wrap(err, "Failed to bind request body to song object")
		apiErr := api.CommitError(err,
			songerrors.BadSongDataCode,
			"The song data received was malformed. Please contact the developer")
		return gateway.ErrorResponse(c, apiErr)
	}

	createdSong, apiErr := g.usecase.CreateSong(ctx, authHeader, song)
	if apiErr != nil {
		return gateway.ErrorResponse(c, apiErr)
	}

	return c.JSON(http.StatusOK, createdSong)
}

func (g Gateway) UpdateSong(c echo.Context, songID string) error {
	ctx := request.Context(c)

	authHeader, apiErr := request.AuthHeader(c)
	if apiErr != nil {
		return gateway.ErrorResponse(c, apiErr)
	}

	song := songentity.Song{}
	err := c.Bind(&song)
	if err != nil {
		err = errors.Wrap(err, "Failed to bind request body to song object")
		apiErr := api.CommitError(err,
			songerrors.BadSongDataCode,
			"The song data received was malformed. Please contact the developer")
		return gateway.ErrorResponse(c, apiErr)
	}

	updatedSong, apiErr := g.usecase.UpdateSong(ctx, authHeader, songID, song)
	if apiErr != nil {
		return gateway.ErrorResponse(c, apiErr)
	}

	return c.JSON(http.StatusOK, updatedSong)
}

func (g Gateway) DeleteSong(c echo.Context, songID string) error {
	ctx := request.Context(c)

	authHeader, apiErr := request.AuthHeader(c)
	if apiErr != nil {
		return gateway.ErrorResponse(c, apiErr)
	}

	apiErr = g.usecase.DeleteSong(ctx, authHeader, songID)
	if apiErr != nil {
		return gateway.ErrorResponse(c, apiErr)
	}

	return c.NoContent(http.StatusOK)
}

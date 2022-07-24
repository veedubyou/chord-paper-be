package songgateway

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/veedubyou/chord-paper-be/go-rewrite/src/errors/api"
	"github.com/veedubyou/chord-paper-be/go-rewrite/src/errors/gateway"
	"github.com/veedubyou/chord-paper-be/go-rewrite/src/lib/request"
	songentity "github.com/veedubyou/chord-paper-be/go-rewrite/src/song/entity"
	songerrors "github.com/veedubyou/chord-paper-be/go-rewrite/src/song/errors"
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
	ctx := request.Context(c)

	songID, err := uuid.Parse(songIDStr)
	if err != nil {
		err = errors.Wrap(err, "Failed to parse song ID")
		apiErr := api.CommitError(err,
			songerrors.InvalidIDCode,
			"The song ID provided is malformed")
		return gateway.ErrorResponse(c, apiErr)
	}

	song, apiErr := g.usecase.GetSong(ctx, songID)
	if apiErr != nil {
		return gateway.ErrorResponse(c, apiErr)
	}

	return c.JSON(http.StatusOK, song)
}

func (g Gateway) CreateSong(c echo.Context) error {
	ctx := request.Context(c)

	song := songentity.Song{}
	err := c.Bind(&song)
	if err != nil {
		err = errors.Wrap(err, "Failed to bind request body to song object")
		apiErr := api.CommitError(err,
			songerrors.BadSongDataCode,
			"The song data received was malformed. Please contact the developer")
		return gateway.ErrorResponse(c, apiErr)
	}

	authHeader := c.Request().Header.Get("authorization")
	createdSong, apiErr := g.usecase.CreateSong(ctx, authHeader, song)
	if apiErr != nil {
		return gateway.ErrorResponse(c, apiErr)
	}

	return c.JSON(http.StatusOK, createdSong)
}

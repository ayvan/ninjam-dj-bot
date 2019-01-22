package api

import (
	"github.com/Ayvan/ninjam-dj-bot/tracks"
	"github.com/labstack/echo"
	"net/http"
)

func Tracks(ctx echo.Context) error {
	t := tracks.GetTracks()

	return ctx.JSON(http.StatusOK, t)
}

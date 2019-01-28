package api

import (
	"github.com/Ayvan/ninjam-dj-bot/config"
	"github.com/Ayvan/ninjam-dj-bot/helpers"
	"github.com/Ayvan/ninjam-dj-bot/tracks"
	"github.com/Ayvan/ninjam-dj-bot/tracks_sync"
	"github.com/labstack/echo"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
)

type ErrorResp struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}

func Tracks(ctx echo.Context) error {
	t, err := tracks.Tracks()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, newError(http.StatusInternalServerError, err.Error()))
	}

	return ctx.JSON(http.StatusOK, t)
}

func Tags(ctx echo.Context) error {
	t, err := tracks.Tags()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, newError(http.StatusInternalServerError, err.Error()))
	}

	return ctx.JSON(http.StatusOK, t)
}

// PostTrack /tracks
func PostTrack(ctx echo.Context) error {
	file, err := ctx.FormFile("file")
	if err != nil {
		return err
	}

	if !helpers.IsMP3(file.Filename) {
		return ctx.JSON(http.StatusBadRequest, newError(http.StatusBadRequest, "bad file type, must be MP3 file"))
	}

	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	filePath := path.Join(config.Get().TracksDir, file.Filename)

	// файл с таким именем уже есть - нужно сохранить с другим именем
	for {
		_, err = os.OpenFile(filePath, os.O_RDONLY, 665)
		if err != nil {
			// ошибка - файла нет, всё ОК, можно создавать новый
			break
		}
		var newFileName string
		newFileName, err = helpers.NewFileName(file.Filename)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, newError(http.StatusBadRequest, "bad file name"))
		}

		filePath = path.Join(config.Get().TracksDir, newFileName)
	}

	// Destination
	dst, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer dst.Close()

	// Copy
	if _, err = io.Copy(dst, src); err != nil {
		return err
	}

	track, err := tracks_sync.ProcessMP3Track(filePath)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, newError(http.StatusInternalServerError, err.Error()))
	}

	return ctx.JSON(http.StatusCreated, track)
}

// PutTrack /tracks/:id
func PutTrack(ctx echo.Context) error {
	idParam := ctx.Param("id")

	id, err := strconv.Atoi(idParam)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, newError(http.StatusBadRequest, err.Error()))
	}

	req := tracks.Track{}
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, newError(http.StatusBadRequest, err.Error()))
	}

	track := &tracks.Track{}
	db := tracks.DB().First(track, "id", id)
	if db.RecordNotFound() {
		return ctx.JSON(http.StatusNotFound, newError(http.StatusNotFound))
	} else if db.Error != nil {
		return ctx.JSON(http.StatusInternalServerError, newError(http.StatusInternalServerError, err.Error()))
	}

	// данные модели ORM, путь к файлу, число проигрываний менять запрещено,
	// остальное - разрешено
	req.Model = track.Model
	req.FilePath = track.FilePath
	req.Played = track.Played

	db = tracks.DB().Save(req)
	if db.Error != nil {
		return ctx.JSON(http.StatusInternalServerError, newError(http.StatusInternalServerError, err.Error()))
	}

	return ctx.JSON(http.StatusOK, req)
}

func newError(code int, message ...string) ErrorResp {
	msg := ""
	if len(message) == 0 || message[0] == "" {
		msg = http.StatusText(code)
	}

	return ErrorResp{Error: msg, Code: code}
}

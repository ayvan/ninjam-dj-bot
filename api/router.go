package api

import (
	"github.com/ayvan/ninjam-dj-bot/dj"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"net/http"
)

// Echo is echo instance
var Echo *echo.Echo

func init() {
	Echo = echo.New()
	Echo.HideBanner = true
	Echo.HidePort = true
	Echo.Pre(middleware.RemoveTrailingSlash())
}

// Run app
// Run(""0.0.0.0:8080")
func Run(hostAndPort string, jamManager *dj.JamManager) {
	routes := Echo.Group("/v1")

	routes.Use(NoCacheHeaders)
	routes.POST("/login", Login)

	routes.Use(Auth())

	routes.GET("/tracks", Tracks)
	routes.GET("/tracks/:id", Track)
	routes.PUT("/tracks/:id", PutTrack)
	routes.POST("/tracks", PostTrack)

	routes.GET("/playlists", Playlists)
	routes.GET("/playlists/:id", Playlist)
	routes.PUT("/playlists/:id", PutPlaylist)
	routes.POST("/playlists", PostPlaylist)

	routes.GET("/tags", Tags)
	routes.GET("/tags/:id", Tag)
	routes.PUT("/tags/:id", PutTag)
	routes.POST("/tags", PostTag)

	routes.GET("/authors", Authors)
	routes.GET("/authors/:id", Author)
	routes.PUT("/authors/:id", PutAuthor)
	routes.POST("/authors", PostAuthor)

	queueController := QueueController{jm: jamManager}
	routes.GET("/queue/users", queueController.Users)
	routes.POST("/queue/:command", queueController.Command)
	routes.POST("/tts", queueController.TTS)

	routes.GET("/test", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "ok"})
	})

	if err := Echo.Start(hostAndPort); err != nil {
		panic(err)
	}
}

func NoCacheHeaders(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		res := c.Response()
		res.Header().Set("Cache-Control", "no-store, must-revalidate")
		res.Header().Set("Expires", "0")

		return next(c)
	}
}

func Auth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ok := authenticateUser(c)
			if !ok {
				return echo.ErrUnauthorized
			}

			return next(c)
		}
	}
}

func authenticateUser(ctx echo.Context) bool {
	ok, _ := authenticator.Validate(ctx.Request())

	return ok
}

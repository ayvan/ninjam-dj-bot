package api

import (
	"github.com/labstack/echo"
	"net/http"
)

// Echo is echo instance
var Echo *echo.Echo

func init() {
	Echo = echo.New()
	Echo.HideBanner = true
	Echo.HidePort = true
}

// Run app
// Run(""0.0.0.0:8080")
func Run(hostAndPort string) {
	routes := Echo.Group("/v1")

	routes.GET("/tracks", Tracks)
	routes.POST("/tracks/", PostTrack)
	routes.PUT("/tracks/:id", PutTrack)
	routes.GET("/test", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "ok"})
	})

	if err := Echo.Start(hostAndPort); err != nil {
		panic(err)
	}
}

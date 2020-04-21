package serviceHttp

import (
	"forum/service"
	"github.com/labstack/echo"
)

func RegisterHTTPEndpoints(router *echo.Echo, u service.UseCase) {
	h := NewHandler(u)

	router.POST("/service/clear", h.ClearDB)
	router.GET("/service/status", h.GetStatus)
}

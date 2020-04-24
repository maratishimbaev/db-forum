package serviceHttp

import (
	"forum/service"
	"github.com/labstack/echo"
)

func RegisterHTTPEndpoints(router *echo.Echo, u service.UseCase) {
	h := NewHandler(u)

	router.POST("api/service/clear", h.ClearDB)
	router.GET("api/service/status", h.GetStatus)
}

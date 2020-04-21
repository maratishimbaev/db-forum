package forumHttp

import (
	"forum/forum"
	"github.com/labstack/echo"
)

func RegisterHTTPEndpoints(router *echo.Echo, u forum.UseCase) {
	h := NewHandler(u)

	router.POST("/forum/create", h.CreateForum)
	router.GET("/forum/:slug/details", h.GetForum)
}

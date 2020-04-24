package forumHttp

import (
	"forum/forum"
	"github.com/labstack/echo"
)

func RegisterHTTPEndpoints(router *echo.Echo, u forum.UseCase) {
	h := NewHandler(u)

	router.POST("api//forum/create", h.CreateForum)
	router.GET("api//forum/:slug/details", h.GetForum)
}

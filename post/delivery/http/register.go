package postHttp

import (
	"forum/post"
	"github.com/labstack/echo"
)

func RegisterHTTPEndpoints(router *echo.Echo, u post.UseCase) {
	h := NewHandler(u)

	router.GET("api/post/:id/details", h.GetPostFull)
	router.POST("api/post/:id/details", h.ChangePost)
	router.POST("api/thread/:slug_or_id/create", h.CreatePosts)
	router.GET("api/thread/:slug_or_id/posts", h.GetThreadPosts)
}

package postHttp

import (
	"forum/post"
	"github.com/labstack/echo"
)

func RegisterHTTPEndpoints(router *echo.Echo, u post.UseCase) {
	h := NewHandler(u)

	router.GET("/post/:id/details", h.GetPostFull)
	router.POST("/post/:id/details", h.ChangePost)
	router.POST("/thread/:slug_or_id/create", h.CreatePosts)
	router.GET("thread/:slug_or_id/posts", h.GetThreadPosts)
}

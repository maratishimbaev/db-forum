package threadHttp

import (
	"forum/thread"
	"github.com/labstack/echo"
)

func RegisterHTTPEndpoints(router *echo.Echo, u thread.UseCase) {
	h := NewHandler(u)

	router.POST("/forum/:slug/create", h.CreateThread)
	router.GET("/forum/:slug/threads", h.GetThreads)
	router.GET("/thread/:slug_or_id/details", h.GetThread)
	router.POST("/thread/:slug_or_id/details", h.ChangeThread)
	router.POST("/thread/:slug_or_id/vote", h.VoteThread)
}

package threadHttp

import (
	"forum/thread"
	"github.com/labstack/echo"
)

func RegisterHTTPEndpoints(router *echo.Echo, u thread.UseCase) {
	h := NewHandler(u)

	router.POST("api/forum/:slug/create", h.CreateThread)
	router.GET("api/forum/:slug/threads", h.GetThreads)
	router.GET("api/thread/:slug_or_id/details", h.GetThread)
	router.POST("api/thread/:slug_or_id/details", h.ChangeThread)
	router.POST("api/thread/:slug_or_id/vote", h.VoteThread)
}

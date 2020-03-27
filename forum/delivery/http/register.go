package forumHttp

import (
	"forum/forum"
	"github.com/gin-gonic/gin"
)

func RegisterHTTPEndpoints(router *gin.Engine, u forum.UseCase) {
	h := NewHandler(u)

	router.POST("/forum/create", h.CreateForum)
	router.GET("/forum/:slug/details", h.GetForum)
}

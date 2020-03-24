package http

import (
	"forum/forum"
	"github.com/gin-gonic/gin"
)

func RegisterHTTPEndpoints(router *gin.Engine, u forum.UseCase) {
	h := NewHandler(u)

	router.POST("/forum/create", h.CreateForum)
	router.POST("/forum/:slug/detail")
}

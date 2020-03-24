package http

import (
	"forum/user"
	"github.com/gin-gonic/gin"
)

func RegisterHTTPEndpoints(router *gin.Engine, u user.UseCase) {
	h := NewHandler(u)

	router.POST("/user/:nickname/create", h.CreateUser)
	router.GET("/user/:nickname/profile", h.GetUser)
	router.POST("/user/:nickname/profile", h.ChangeUser)
}

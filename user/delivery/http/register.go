package userHttp

import (
	"forum/user"
	"github.com/labstack/echo"
)

func RegisterHTTPEndpoints(router *echo.Echo, u user.UseCase) {
	h := NewHandler(u)

	router.POST("/user/:nickname/create", h.CreateUser)
	router.GET("/user/:nickname/profile", h.GetUser)
	router.POST("/user/:nickname/profile", h.ChangeUser)
	router.GET("/forum/:slug/users", h.GetForumUsers)
}

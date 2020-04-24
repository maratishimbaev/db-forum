package userHttp

import (
	"forum/user"
	"github.com/labstack/echo"
)

func RegisterHTTPEndpoints(router *echo.Echo, u user.UseCase) {
	h := NewHandler(u)

	router.POST("api/user/:nickname/create", h.CreateUser)
	router.GET("api/user/:nickname/profile", h.GetUser)
	router.POST("api/user/:nickname/profile", h.ChangeUser)
	router.GET("api/forum/:slug/users", h.GetForumUsers)
}

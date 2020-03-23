package http

import (
	"forum/models"
	"forum/user"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Handler struct {
	useCase user.UseCase
}

func NewHandler(useCase user.UseCase) *Handler {
	return &Handler{useCase: useCase}
}

func (h *Handler) CreateUser(c *gin.Context) {
	newUser := models.User{
		Nickname: c.Param("nickname"),
	}

	if err := c.BindJSON(newUser); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
	}

	user, err := h.useCase.CreateUser(&newUser)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	c.JSON(http.StatusOK, user)
}

func (h *Handler) GetUser(c *gin.Context) {
	nickname := c.Param("nickname")

	user, err := h.useCase.GetUser(nickname)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	c.JSON(http.StatusOK, user)
}

func (h *Handler) ChangeUser(c *gin.Context) {
	newUser := models.User{
		Nickname: c.Param("nickname"),
	}

	if err := c.BindJSON(newUser); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
	}

	user, err := h.useCase.ChangeUser(&newUser)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	c.JSON(http.StatusOK, user)
}

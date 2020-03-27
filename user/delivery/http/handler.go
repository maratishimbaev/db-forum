package userHttp

import (
	"forum/models"
	"forum/user"
	"github.com/gin-gonic/gin"
	"log"
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

	if err := c.BindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: err.Error(),
		})
		return
	}

	log.Printf("newUser: %s", newUser)

	user, err := h.useCase.CreateUser(&newUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
		return
	}

	log.Printf("user: %s", user)

	c.JSON(http.StatusOK, user)
}

func (h *Handler) GetUser(c *gin.Context) {
	nickname := c.Param("nickname")

	user, err := h.useCase.GetUser(nickname)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *Handler) ChangeUser(c *gin.Context) {
	newUser := models.User{
		Nickname: c.Param("nickname"),
	}

	if err := c.BindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, models.Error{
			Message: err.Error(),
		})
		return
	}

	user, err := h.useCase.ChangeUser(&newUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

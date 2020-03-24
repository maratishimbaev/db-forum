package http

import (
	"forum/forum"
	"forum/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Handler struct {
	useCase forum.UseCase
}

func NewHandler(useCase forum.UseCase) *Handler {
	return &Handler{useCase: useCase}
}

func (h *Handler) CreateForum(c *gin.Context) {
	var newForum models.Forum

	if err := c.BindJSON(newForum); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
	}

	forum, err := h.useCase.CreateForum(&newForum)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	c.JSON(http.StatusOK, forum)
}

func (h *Handler) GetForum(c *gin.Context) {
	slug := c.Param("slug")

	forum, err := h.useCase.GetForum(slug)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	gin.Default()

	c.JSON(http.StatusOK, forum)
}

package forumHttp

import (
	"forum/forum"
	"forum/models"
	"github.com/labstack/echo"
	"net/http"
)

type Handler struct {
	useCase forum.UseCase
}

func NewHandler(useCase forum.UseCase) *Handler {
	return &Handler{useCase: useCase}
}

func (h *Handler) CreateForum(c echo.Context) (err error) {
	var newForum models.Forum

	if err := c.Bind(&newForum); err != nil {
		return c.JSON(http.StatusBadRequest, models.Error{
			Message: err.Error(),
		})
	}

	forum, err := h.useCase.CreateForum(&newForum)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, forum)
}

func (h *Handler) GetForum(c echo.Context) (err error) {
	slug := c.Param("slug")

	forum, err := h.useCase.GetForum(slug)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, forum)
}

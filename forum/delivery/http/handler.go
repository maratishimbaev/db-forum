package forumHttp

import (
	"errors"
	_forum "forum/forum"
	"forum/models"
	"github.com/labstack/echo"
	"net/http"
)

type Handler struct {
	useCase _forum.UseCase
}

func NewHandler(useCase _forum.UseCase) *Handler {
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
	switch true {
	case errors.Is(err, _forum.ErrUserNotFound):
		return c.JSON(http.StatusNotFound, models.Error{
			Message: err.Error(),
		})
	case errors.Is(err, _forum.ErrAlreadyExists):
		return c.JSON(http.StatusConflict, forum)
	case err == nil:
		return c.JSON(http.StatusCreated, forum)
	default:
		return c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
	}
}

func (h *Handler) GetForum(c echo.Context) (err error) {
	slug := c.Param("slug")

	forum, err := h.useCase.GetForum(slug)
	switch true {
	case errors.Is(err, _forum.ErrNotFound):
		return c.JSON(http.StatusNotFound, models.Error{
			Message: err.Error(),
		})
	case err == nil:
		return c.JSON(http.StatusOK, forum)
	default:
		return c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
	}
}

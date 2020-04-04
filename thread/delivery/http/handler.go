package threadHttp

import (
	"forum/models"
	_thread "forum/thread"
	"github.com/labstack/echo"
	"net/http"
	"strconv"
)

type Handler struct {
	useCase _thread.UseCase
}

func NewHandler(useCase _thread.UseCase) *Handler {
	return &Handler{useCase: useCase}
}

func (h *Handler) CreateThread(c echo.Context) (err error) {
	newThread := models.Thread{
		Forum: c.Param("slug"),
	}

	if err := c.Bind(&newThread); err != nil {
		return c.JSON(http.StatusBadRequest, models.Error{
			Message: err.Error(),
		})
	}

	thread, err := h.useCase.CreateThread(&newThread)
	switch err.(type) {
	case *_thread.UserOrForumNotFound:
		return c.JSON(http.StatusNotFound, models.Error{
			Message: err.Error(),
		})
	case *_thread.AlreadyExists:
		return c.JSON(http.StatusConflict, thread)
	case nil:
		return c.JSON(http.StatusCreated, thread)
	default:
		return c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
	}
}

func (h *Handler) GetThreads(c echo.Context) (err error) {
	limit, err := strconv.ParseUint(c.Request().URL.Query().Get("limit"), 10, 64)
	if err != nil {
		limit = 0
	}

	desc, err := strconv.ParseBool(c.Request().URL.Query().Get("desc"))
	if err != nil {
		desc = false
	}

	thread, err := h.useCase.GetThreads(
		c.Param("slug"),
		limit,
		c.Request().URL.Query().Get("since"),
		desc,
	)
	switch err.(type) {
	case *_thread.NotFound:
		return c.JSON(http.StatusNotFound, models.Error{
			Message: err.Error(),
		})
	case nil:
		return c.JSON(http.StatusOK, thread)
	default:
		return c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
	}
}

func (h *Handler) GetThread(c echo.Context) (err error) {
	thread, err := h.useCase.GetThread(
		c.Param("slug_or_id"),
	)
	switch err.(type) {
	case *_thread.NotFound:
		return c.JSON(http.StatusNotFound, models.Error{
			Message: err.Error(),
		})
	case nil:
		return c.JSON(http.StatusOK, thread)
	default:
		return c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
	}
}

func (h *Handler) ChangeThread(c echo.Context) (err error) {
	var newThread models.Thread

	if err = c.Bind(&newThread); err != nil {
		return c.JSON(http.StatusBadRequest, models.Error{
			Message: err.Error(),
		})
	}

	thread, err := h.useCase.ChangeThread(
		c.Param("slug_or_id"),
		&newThread,
	)
	switch err.(type) {
	case *_thread.NotFound:
		return c.JSON(http.StatusNotFound, models.Error{
			Message: err.Error(),
		})
	case nil:
		return c.JSON(http.StatusOK, thread)
	default:
		return c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
	}
}

func (h *Handler) VoteThread(c echo.Context) (err error) {
	var newVote models.Vote

	if err = c.Bind(&newVote); err != nil {
		return c.JSON(http.StatusBadRequest, models.Error{
			Message: err.Error(),
		})
	}

	thread, err := h.useCase.VoteThread(
		c.Param("slug_or_id"),
		newVote,
	)
	switch err.(type) {
	case *_thread.NotFound:
		return c.JSON(http.StatusNotFound, models.Error{
			Message: err.Error(),
		})
	case nil:
		return c.JSON(http.StatusOK, thread)
	default:
		return c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
	}
}

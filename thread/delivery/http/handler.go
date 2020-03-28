package threadHttp

import (
	"forum/models"
	"forum/thread"
	"github.com/labstack/echo"
	"net/http"
	"strconv"
)

type Handler struct {
	useCase thread.UseCase
}

func NewHandler(useCase thread.UseCase) *Handler {
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
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, thread)
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
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, thread)
}

func (h *Handler) GetThread(c echo.Context) (err error) {
	thread, err := h.useCase.GetThread(
		c.Param("slug_or_id"),
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, thread)
}

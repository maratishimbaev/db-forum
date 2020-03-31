package serviceHttp

import (
	"forum/models"
	"forum/service"
	"github.com/labstack/echo"
	"net/http"
)

type Handler struct {
	useCase service.UseCase
}

func NewHandler(useCase service.UseCase) *Handler {
	return &Handler{useCase: useCase}
}

func (h *Handler) ClearDB(c echo.Context) (err error) {
	err = h.useCase.ClearDB()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) GetStatus(c echo.Context) (err error) {
	status, err := h.useCase.GetStatus()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, status)
}

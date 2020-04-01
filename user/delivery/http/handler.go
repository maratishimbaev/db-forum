package userHttp

import (
	"forum/models"
	"forum/user"
	"github.com/labstack/echo"
	"net/http"
	"strconv"
)

type Handler struct {
	useCase user.UseCase
}

func NewHandler(useCase user.UseCase) *Handler {
	return &Handler{useCase: useCase}
}

func (h *Handler) CreateUser(c echo.Context) (err error) {
	newUser := models.User{
		Nickname: c.Param("nickname"),
	}

	if err := c.Bind(&newUser); err != nil {
		return c.JSON(http.StatusBadRequest, models.Error{
			Message: err.Error(),
		})
	}

	user, err := h.useCase.CreateUser(&newUser)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, user)
}

func (h *Handler) GetUser(c echo.Context) (err error) {
	nickname := c.Param("nickname")

	user, err := h.useCase.GetUser(nickname)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, user)
}

func (h *Handler) ChangeUser(c echo.Context) (err error) {
	newUser := models.User{
		Nickname: c.Param("nickname"),
	}

	if err := c.Bind(&newUser); err != nil {
		return c.JSON(http.StatusBadRequest, models.Error{
			Message: err.Error(),
		})
	}

	user, err := h.useCase.ChangeUser(&newUser)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, user)
}

func (h *Handler) GetForumUsers(c echo.Context) (err error) {
	limit, err := strconv.ParseUint(c.Request().URL.Query().Get("limit"), 10, 64)
	if err != nil {
		limit = 0
	}

	desc, err := strconv.ParseBool(c.Request().URL.Query().Get("desc"))
	if err != nil {
		desc = false
	}

	users, err := h.useCase.GetForumUsers(
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

	return c.JSON(http.StatusOK, users)
}

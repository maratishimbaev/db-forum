package userHttp

import (
	"errors"
	"forum/models"
	_user "forum/user"
	"github.com/labstack/echo"
	"net/http"
	"strconv"
)

type Handler struct {
	useCase _user.UseCase
}

func NewHandler(useCase _user.UseCase) *Handler {
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

	users, err := h.useCase.CreateUser(&newUser)
	switch true {
	case errors.Is(err, _user.ErrAlreadyExists):
		return c.JSON(http.StatusConflict, users)
	case err == nil:
		return c.JSON(http.StatusCreated, users[0])
	default:
		return c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
	}
}

func (h *Handler) GetUser(c echo.Context) (err error) {
	nickname := c.Param("nickname")

	user, err := h.useCase.GetUser(nickname)
	switch true {
	case errors.Is(err, _user.ErrNotFound):
		return c.JSON(http.StatusNotFound, models.Error{
			Message: err.Error(),
		})
	case err == nil:
		return c.JSON(http.StatusOK, user)
	default:
		return c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
	}
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
	switch true {
	case errors.Is(err, _user.ErrNotFound):
		return c.JSON(http.StatusNotFound, models.Error{
			Message: err.Error(),
		})
	case errors.Is(err, _user.ErrConflictData):
		return c.JSON(http.StatusConflict, models.Error{
			Message: err.Error(),
		})
	case err == nil:
		return c.JSON(http.StatusOK, user)
	default:
		return c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
	}
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
	switch true {
	case errors.Is(err, _user.ErrForumNotFound):
		return c.JSON(http.StatusNotFound, models.Error{
			Message: err.Error(),
		})
	case err == nil:
		return c.JSON(http.StatusOK, users)
	default:
		return c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
	}
}

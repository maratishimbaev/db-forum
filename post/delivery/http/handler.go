package postHttp

import (
	"forum/models"
	_post "forum/post"
	"github.com/labstack/echo"
	"net/http"
	"strconv"
)

type Handler struct {
	useCase _post.UseCase
}

func NewHandler(useCase _post.UseCase) *Handler {
	return &Handler{useCase: useCase}
}

func (h *Handler) GetPostFull(c echo.Context) (err error) {
	postID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.Error{
			Message: err.Error(),
		})
	}

	post, err := h.useCase.GetPostFull(
		postID,
		c.Request().URL.Query()["related"],
	)
	switch err.(type) {
	case *_post.NotFound:
		return c.JSON(http.StatusNotFound, models.Error{
			Message: err.Error(),
		})
	case nil:
		return c.JSON(http.StatusOK, post)
	default:
		return c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
	}
}

func (h *Handler) ChangePost(c echo.Context) (err error) {
	postID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.Error{
			Message: err.Error(),
		})
	}

	newPost := models.Post{
		ID: postID,
	}

	if err = c.Bind(&newPost); err != nil {
		return c.JSON(http.StatusBadRequest, models.Error{
			Message: err.Error(),
		})
	}

	post, err := h.useCase.ChangePost(&newPost)
	switch err.(type) {
	case *_post.NotFound:
		return c.JSON(http.StatusNotFound, models.Error{
			Message: err.Error(),
		})
	case nil:
		return c.JSON(http.StatusOK, post)
	default:
		return c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
	}
}

func (h *Handler) CreatePosts(c echo.Context) (err error) {
	var newPosts []models.Post

	if err = c.Bind(&newPosts); err != nil {
		return c.JSON(http.StatusBadRequest, models.Error{
			Message: err.Error(),
		})
	}

	posts, err := h.useCase.CreatePosts(
		c.Param("slug_or_id"),
		newPosts,
	)
	switch err.(type) {
	case *_post.ThreadNotFound:
		return c.JSON(http.StatusNotFound, models.Error{
			Message: err.Error(),
		})
	case *_post.ParentNotInThread:
		return c.JSON(http.StatusConflict, models.Error{
			Message: err.Error(),
		})
	case nil:
		return c.JSON(http.StatusCreated, posts)
	default:
		return c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
	}
}

func (h *Handler) GetThreadPosts(c echo.Context) (err error) {
	limit, err := strconv.ParseUint(c.Request().URL.Query().Get("limit"), 10, 64)
	if err != nil {
		limit = 0
	}

	since, err := strconv.ParseUint(c.Request().URL.Query().Get("since"), 10, 64)
	if err != nil {
		since = 0
	}

	desc, err := strconv.ParseBool(c.Request().URL.Query().Get("desc"))
	if err != nil {
		desc = false
	}

	posts, err := h.useCase.GetThreadPosts(
		c.Param("slug_or_id"),
		limit,
		since,
		c.Request().URL.Query().Get("sort"),
		desc,
	)
	switch err.(type) {
	case *_post.ThreadNotFound:
		return c.JSON(http.StatusNotFound, models.Error{
			Message: err.Error(),
		})
	case nil:
		return c.JSON(http.StatusOK, posts)
	default:
		return c.JSON(http.StatusInternalServerError, models.Error{
			Message: err.Error(),
		})
	}
}

package thread

import "fmt"

type UserOrForumNotFound struct {}

func NewUserOrForumNotFound() *UserOrForumNotFound {
	return &UserOrForumNotFound{}
}

func (e *UserOrForumNotFound) Error() string {
	return "User ot forum not found"
}

type AlreadyExists struct {
	Slug string
}

func NewAlreadyExists(slug string) *AlreadyExists {
	return &AlreadyExists{Slug: slug}
}

func (e *AlreadyExists) Error() string {
	return fmt.Sprintf("Thread with slug %s already exists", e.Slug)
}

type NotFound struct {
	SlugOrID string
}

func NewNotFound(slugOrID string) *NotFound {
	return &NotFound{SlugOrID: slugOrID}
}

func (e *NotFound) Error() string {
	return fmt.Sprintf("Can't find thread with slug or id %s", e.SlugOrID)
}

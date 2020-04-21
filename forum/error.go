package forum

import (
	"fmt"
)

type UserNotFound struct {
	Nickname string
}

func NewUserNotFound(nickname string) *UserNotFound {
	return &UserNotFound{Nickname: nickname}
}

func (e *UserNotFound) Error() string {
	return fmt.Sprintf("Can't find user with nickname %s", e.Nickname)
}

type AlreadyExits struct{}

func NewAlreadyExits() *AlreadyExits {
	return &AlreadyExits{}
}

func (e *AlreadyExits) Error() string {
	return ""
}

type NotFound struct {
	Slug string
}

func NewNotFound(slug string) *NotFound {
	return &NotFound{Slug: slug}
}

func (e *NotFound) Error() string {
	return fmt.Sprintf("Can't find forum with slug %s", e.Slug)
}

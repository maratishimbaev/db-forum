package user

import "fmt"

type AlreadyExists struct {}

func NewAlreadyExists() *AlreadyExists {
	return &AlreadyExists{}
}

func (e *AlreadyExists) Error() string {
	return "User already exists"
}

type NotFound struct {
	Nickname string
}

func NewNotFound(nickname string) *NotFound {
	return &NotFound{Nickname: nickname}
}

func (e *NotFound) Error() string {
	return fmt.Sprintf("Can't find user with nickname %s", e.Nickname)
}

type ConflictData struct {}

func NewConflictData() *ConflictData {
	return &ConflictData{}
}

func (e *ConflictData) Error() string {
	return "Conflict data"
}

type ForumNotFound struct {
	Slug string
}

func NewForumNotFound(slug string) *ForumNotFound {
	return &ForumNotFound{Slug: slug}
}

func (e *ForumNotFound) Error() string {
	return fmt.Sprintf("Can't find forum with slug %s", e.Slug)
}

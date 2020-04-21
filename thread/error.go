package thread

import (
	"errors"
)

var (
	UserOrForumNotFound = errors.New("user or forum not found")
	AlreadyExists       = errors.New("thread already exists")
	NotFound            = errors.New("can't find thread")
)

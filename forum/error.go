package forum

import (
	"errors"
)

var (
	ErrUserNotFound  = errors.New("can't find user")
	ErrAlreadyExists = errors.New("already exists")
	ErrNotFound      = errors.New("can't find forum")
)

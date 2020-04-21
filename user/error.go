package user

import (
	"errors"
)

var (
	ErrAlreadyExists = errors.New("user already exists")
	ErrNotFound      = errors.New("can't find user")
	ErrConflictData  = errors.New("conflict data")
	ErrForumNotFound = errors.New("can't find forum")
)

package post

import (
	"errors"
)

var (
	NotFound          = errors.New("can't find post")
	ThreadNotFound    = errors.New("can't find thread")
	ParentNotInThread = errors.New("parent not in thread")
)

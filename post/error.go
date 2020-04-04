package post

import "fmt"

type NotFound struct {
	ID uint64
}

func NewNotFound(id uint64) *NotFound {
	return &NotFound{ID: id}
}

func (e *NotFound) Error() string {
	return fmt.Sprintf("Can't find post with id %d", e.ID)
}

type ThreadNotFound struct {
	SlugOrID string
}

func NewThreadNotFound(slugOrID string) *ThreadNotFound {
	return &ThreadNotFound{SlugOrID: slugOrID}
}

func (e *ThreadNotFound) Error() string {
	return fmt.Sprintf("Can't find thread with slug or id %s", e.SlugOrID)
}

type ParentNotInThread struct {}

func NewParentNotInThread() *ParentNotInThread {
	return &ParentNotInThread{}
}

func (e *ParentNotInThread) Error() string {
	return "Parent not in thread"
}

package thread

import (
	"forum/models"
	"time"
)

type Repository interface {
	CreateThread(newThread *models.Thread) (thread *models.Thread, err error)
	GetThreads(slug string, limit uint64, since time.Time, desc bool) (threads []models.Thread, err error)
	GetThreadByID(id uint64) (thread *models.Thread, err error)
	GetThreadBySlug(slug string) (thread *models.Thread, err error)
	GetThread(slugOrID string) (thread *models.Thread, err error)
	ChangeThread(slugOrID string, newThread *models.Thread) (thread *models.Thread, err error)
	VoteThread(slugOrID string, vote models.Vote) (thread *models.Thread, err error)
}

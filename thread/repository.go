package thread

import "forum/models"

type Repository interface {
	CreateThread(newThread *models.Thread) (thread models.Thread, err error)
	GetThreads(slug string, limit uint64, since string, desc bool) (threads []models.Thread, err error)
	GetThread(slugOrID string) (thread models.Thread, err error)
	ChangeThread(slugOrID string, newThread *models.Thread) (thread models.Thread, err error)
	VoteThread(slugOrID string, vote models.Vote) (thread models.Thread, err error)
}

package thread

import "forum/models"

type Repository interface {
	CreateThread(newThread *models.Thread) (thread models.Thread, err error)
	GetThreads(slug string, limit uint64, since string, desc bool) (threads []models.Thread, err error)
	GetThread(slug string, id uint64) (thread models.Thread, err error)
}

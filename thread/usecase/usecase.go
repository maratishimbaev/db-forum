package threadUsecase

import (
	"forum/models"
	"forum/thread"
	"time"
)

type UseCase struct {
	repository thread.Repository
}

func NewUseCase(repository thread.Repository) *UseCase {
	return &UseCase{repository: repository}
}

func (u *UseCase) CreateThread(newThread *models.Thread) (thread models.Thread, err error) {
	return u.repository.CreateThread(newThread)
}

func (u *UseCase) GetThreads(slug string, limit uint64, since time.Time, desc bool) (threads []models.Thread, err error) {
	return u.repository.GetThreads(slug, limit, since, desc)
}

func (u *UseCase) GetThread(slugOrID string) (thread models.Thread, err error) {
	return u.repository.GetThread(slugOrID)
}

func (u *UseCase) ChangeThread(slugOrID string, newThread *models.Thread) (thread models.Thread, err error) {
	return u.repository.ChangeThread(slugOrID, newThread)
}

func (u *UseCase) VoteThread(slugOrID string, vote models.Vote) (thread models.Thread, err error) {
	return u.repository.VoteThread(slugOrID, vote)
}

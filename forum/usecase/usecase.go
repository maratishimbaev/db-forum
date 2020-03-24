package usecase

import (
	"forum/forum"
	"forum/models"
)

type UseCase struct {
	repository forum.Repository
}

func NewUseCase(repository forum.Repository) *UseCase {
	return &UseCase{repository: repository}
}

func (u *UseCase) CreateForum(newForum *models.Forum) (forum models.Forum, err error) {
	return u.repository.CreateForum(newForum)
}

func (u *UseCase) GetForum(slug string) (forum models.Forum, err error) {
	return u.repository.GetForum(slug)
}

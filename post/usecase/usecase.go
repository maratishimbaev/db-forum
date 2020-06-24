package postUseCase

import (
	"forum/models"
	"forum/post"
)

type UseCase struct {
	repository post.Repository
}

func NewUseCase(repository post.Repository) *UseCase {
	return &UseCase{repository: repository}
}

func (u *UseCase) GetPostFull(id uint64, related []string) (postFull *models.PostFull, err error) {
	return u.repository.GetPostFull(id, related)
}

func (u *UseCase) ChangePost(newPost *models.Post) (post *models.Post, err error) {
	return u.repository.ChangePost(newPost)
}

func (u *UseCase) CreatePosts(threadSlugOrID string, newPosts []models.Post) (posts []models.Post, err error) {
	return u.repository.CreatePosts(threadSlugOrID, newPosts)
}

func (u *UseCase) GetThreadPosts(threadSlugOrID string, limit uint64, since uint64, sort string, desc bool) (posts []models.Post, err error) {
	return u.repository.GetThreadPosts(threadSlugOrID, limit, since, sort, desc)
}

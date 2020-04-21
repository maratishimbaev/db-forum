package post

import "forum/models"

type UseCase interface {
	GetPostFull(id uint64, related []string) (postFull models.PostFull, err error)
	ChangePost(newPost *models.Post) (post models.Post, err error)
	CreatePosts(threadSlugOrID string, newPosts []models.Post) (posts []models.Post, err error)
	GetThreadPosts(threadSlugOrID string, limit uint64, since uint64, sort string, desc bool) (posts []models.Post, err error)
}

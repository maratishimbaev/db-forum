package forum

import "forum/models"

type Repository interface {
	CreateForum(newForum *models.Forum) (forum models.Forum, err error)
	GetForum(slug string) (forum models.Forum, err error)
	GetPostCount(forumSlug string) (postCount uint64, err error)
	GetThreadCount(forumSlug string) (threadCount uint64, err error)
}

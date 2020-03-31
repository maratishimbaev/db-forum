package forum

import "forum/models"

type Repository interface {
	GetPostCount(forumID uint64) (postCount uint64, err error)
	GetThreadCount(forumID uint64) (threadCount uint64, err error)
	CreateForum(newForum *models.Forum) (forum models.Forum, err error)
	GetForum(slug string) (forum models.Forum, err error)
}

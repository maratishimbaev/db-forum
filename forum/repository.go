package forum

import "forum/models"

type Repository interface {
	CreateForum(newForum *models.Forum) (forum models.Forum, err error)
	GetForum(slug string) (forum models.Forum, err error)
}

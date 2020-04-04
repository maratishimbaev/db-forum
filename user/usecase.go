package user

import "forum/models"

type UseCase interface {
	CreateUser(newUser *models.User) (users []models.User, err error)
	GetUser(nickname string) (user models.User, err error)
	ChangeUser(newUser *models.User) (user models.User, err error)
	GetForumUsers(forumSlug string, limit uint64, since string, desc bool) (users []models.User, err error)
}

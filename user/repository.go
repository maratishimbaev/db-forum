package user

import "forum/models"

type Repository interface {
	CreateUser(newUser *models.User) (user models.User, err error)
	GetUser(nickname string) (user models.User, err error)
	ChangeUser(newUser *models.User) (user models.User, err error)
	GetUserIDByNickname(nickname string) (id uint64, err error)
	GetUserNicknameByID(id uint64) (nickname string, err error)
}

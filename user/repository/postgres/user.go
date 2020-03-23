package postgres

import (
	"forum/models"
)

type Repository struct {}

func NewRepository() *Repository {
	return &Repository{}
}

type User struct {
	About string
	Email string
	FullName string
	Nickname string
}

func toPostgresUser(user models.User) *User {
	return &User{
		About:    user.About,
		Email:    user.Email,
		FullName: user.FullName,
		Nickname: user.Nickname,
	}
}

func toPostgresUserUpdate(user models.UserUpdate) *User {
	return &User{
		About:    user.About,
		Email:    user.Email,
		FullName: user.FullName,
	}
}

func toModel(user User) *models.User {
	return &models.User{
		About:    user.About,
		Email:    user.Email,
		FullName: user.FullName,
		Nickname: user.Nickname,
	}
}

func (r *Repository) CreateUser(newUser *models.User) (user models.User, err error) {
	return user, err
}

func (r *Repository) GetUser(nickname string) (user models.User, err error) {
	return user ,err
}

func (r *Repository) ChangeUser(newUser *models.User) (user models.User, err error) {
	return user, err
}

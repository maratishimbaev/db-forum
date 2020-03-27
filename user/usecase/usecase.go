package userUseCase

import (
	"forum/models"
	"forum/user"
)

type UseCase struct {
	repository user.Repository
}

func NewUseCase(repository user.Repository) *UseCase {
	return &UseCase{repository}
}

func (u *UseCase) CreateUser(newUser *models.User) (user models.User, err error) {
	return u.repository.CreateUser(newUser)
}

func (u *UseCase) GetUser(nickname string) (user models.User, err error) {
	return u.repository.GetUser(nickname)
}

func (u *UseCase) ChangeUser(newUser *models.User) (user models.User, err error) {
	return u.repository.ChangeUser(newUser)
}

package serviceUsecase

import (
	"forum/models"
	"forum/service"
)

type UseCase struct {
	repository service.Repository
}

func NewUseCase(repository service.Repository) *UseCase {
	return &UseCase{repository: repository}
}

func (u *UseCase) ClearDB() (err error) {
	return u.repository.ClearDB()
}

func (u *UseCase) GetStatus() (status *models.Status, err error) {
	return u.repository.GetStatus()
}

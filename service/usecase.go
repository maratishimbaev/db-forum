package service

import "forum/models"

type UseCase interface {
	ClearDB() (err error)
	GetStatus() (status models.Status, err error)
}

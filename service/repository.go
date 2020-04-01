package service

import "forum/models"

type Repository interface {
	ClearDB() (err error)
	GetStatus() (status models.Status, err error)
}

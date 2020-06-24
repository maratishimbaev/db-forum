package servicePostgres

import (
	"context"
	"database/sql"
	"forum/models"
)

type Repository struct {
	DB *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) ClearDB() (err error) {
	tx, err := r.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	clear := "DELETE FROM vote; DELETE FROM post; DELETE FROM thread;" +
		"DELETE FROM forum_user; DELETE FROM forum; DELETE FROM \"user\""
	_, err = r.DB.Exec(clear)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	_ = tx.Commit()
	return nil
}

func (r *Repository) GetStatus() (*models.Status, error) {
	var status models.Status

	countForum := "SELECT COUNT(*) FROM forum"
	err := r.DB.QueryRow(countForum).Scan(&status.ForumCount)
	if err != nil {
		return nil, err
	}

	countPost := "SELECT COUNT(*) FROM post"
	err = r.DB.QueryRow(countPost).Scan(&status.PostCount)
	if err != nil {
		return nil, err
	}

	countThread := "SELECT COUNT(*) FROM thread"
	err = r.DB.QueryRow(countThread).Scan(&status.ThreadCount)
	if err != nil {
		return nil, err
	}

	countUser := "SELECT COUNT(*) FROM \"user\""
	err = r.DB.QueryRow(countUser).Scan(&status.UserCount)
	if err != nil {
		return nil, err
	}

	return &status, nil
}

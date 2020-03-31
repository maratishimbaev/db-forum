package servicePostgres

import (
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
	beginTransaction := `BEGIN;`
	_, err = r.DB.Exec(beginTransaction)
	if err != nil {
		return err
	}

	rollbackTransaction := `ROLLBACK`

	clearVote := `DELETE FROM vote`
	_, err = r.DB.Exec(clearVote)
	if err != nil {
		r.DB.Exec(rollbackTransaction)
		return err
	}

	clearPost := `DELETE FROM post`
	_, err = r.DB.Exec(clearPost)
	if err != nil {
		r.DB.Exec(rollbackTransaction)
		return err
	}

	clearThread := `DELETE FROM thread`
	_, err = r.DB.Exec(clearThread)
	if err != nil {
		r.DB.Exec(rollbackTransaction)
		return err
	}

	clearForum := `DELETE FROM forum`
	_, err = r.DB.Exec(clearForum)
	if err != nil {
		r.DB.Exec(rollbackTransaction)
		return err
	}

	clearUser := `DELETE FROM "user"`
	_, err = r.DB.Exec(clearUser)
	if err != nil {
		r.DB.Exec(rollbackTransaction)
		return err
	}

	commitTransaction := `COMMIT`
	_, err = r.DB.Exec(commitTransaction)
	if err != nil {
		return err
	}

	return err
}

func (r *Repository) GetStatus() (status models.Status, err error) {
	countForum := `SELECT COUNT(*) FROM forum`
	err = r.DB.QueryRow(countForum).Scan(&status.ForumCount)
	if err != nil {
		return status, err
	}

	countPost := `SELECT COUNT(*) FROM post`
	err = r.DB.QueryRow(countPost).Scan(&status.PostCount)
	if err != nil {
		return status, err
	}

	countThread := `SELECT COUNT(*) FROM thread`
	err = r.DB.QueryRow(countThread).Scan(&status.ThreadCount)
	if err != nil {
		return status, err
	}

	countUser := `SELECT COUNT(*) FROM "user"`
	err = r.DB.QueryRow(countUser).Scan(&status.UserCount)
	if err != nil {
		return status, err
	}

	return status, err
}

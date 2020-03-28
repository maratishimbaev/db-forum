package forumPostgres

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

type Forum struct {
	ID uint64
	Slug string
	Title string
	User uint64
}

func (r *Repository) toPostgres(forum *models.Forum) *Forum {
	var userID uint64
	getUserID := `SELECT id FROM "user" WHERE nickname = $1`
	if err := r.DB.QueryRow(getUserID, forum.User).Scan(&userID); err != nil {
		userID = 0
	}

	return &Forum{
		Slug:  forum.Slug,
		Title: forum.Title,
		User:  userID,
	}
}

func (r *Repository) toModel(forum *Forum) *models.Forum {
	var userNickname string
	getUserNickname := `SELECT nickname FROM "user" WHERE id = $1`
	if err := r.DB.QueryRow(getUserNickname, forum.User).Scan(&userNickname); err != nil {
		userNickname = ""
	}

	return &models.Forum{
		Slug:    forum.Slug,
		Title:   forum.Title,
		User:    userNickname,
	}
}

func (r *Repository) CreateForum(newForum *models.Forum) (forum models.Forum, err error) {
	pgForum := r.toPostgres(newForum)

	createForum := `INSERT INTO forum (slug, title, "user")
					VALUES ($1, $2, $3)`
	_, err = r.DB.Exec(createForum, pgForum.Slug, pgForum.Title, pgForum.User)
	if err != nil {
		return forum, err
	}

	// TODO: posts and threads fields

	return *r.toModel(pgForum), err
}

func (r *Repository) GetForum(slug string) (forum models.Forum, err error) {
	pgForum := Forum{Slug: slug}

	getForum := `SELECT title, "user"
				 FROM forum WHERE slug = $1`
	err = r.DB.QueryRow(getForum, pgForum.Slug).Scan(&pgForum.Title, &pgForum.User)
	if err != nil {
		return forum, err
	}

	// TODO: posts and threads fields

	return *r.toModel(&pgForum), err
}

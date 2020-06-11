package forumPostgres

import (
	"database/sql"
	_forum "forum/forum"
	"forum/models"
)

type repository struct {
	db *sql.DB
}

func NewForumRepository(db *sql.DB) *repository {
	return &repository{db: db}
}

func (r *repository) CreateForum(newForum *models.Forum) (forum models.Forum, err error) {
	var userID uint64
	getUserID := `SELECT id FROM "user" WHERE LOWER(nickname) = LOWER($1)`
	if err := r.db.QueryRow(getUserID, newForum.User).Scan(&userID); err != nil {
		return forum, _forum.ErrUserNotFound
	}

	createForum := `
		INSERT INTO forum (slug, title, "user")
		VALUES ($1, $2, $3)`
	_, err = r.db.Exec(createForum, newForum.Slug, newForum.Title, userID)

	if err != nil {
		var hasUser bool

		checkUser := `SELECT COUNT(*) <> 0 FROM "user" WHERE LOWER(nickname) = LOWER($1)`
		err = r.db.QueryRow(checkUser, newForum.User).Scan(&hasUser)
		if err != nil {
			return forum, err
		}

		if !hasUser {
			return forum, _forum.ErrUserNotFound
		} else {
			forum, err = r.GetForum(newForum.Slug)
			if err != nil {
				return forum, err
			}

			return forum, _forum.ErrAlreadyExists
		}
	}

	forum, err = r.GetForum(newForum.Slug)
	if err != nil {
		return forum, err
	}

	return forum, err
}

func (r *repository) GetForum(slug string) (forum models.Forum, err error) {
	getForum := `SELECT f.title, u.nickname, f.slug, f.posts, f.threads
				 FROM forum f
				 LEFT JOIN "user" u ON f.user = u.id
				 WHERE LOWER(f.slug) = LOWER($1)`
	err = r.db.QueryRow(getForum, slug).Scan(&forum.Title, &forum.User, &forum.Slug, &forum.Posts, &forum.Threads)
	if err != nil {
		return forum, _forum.ErrNotFound
	}

	return forum, err
}

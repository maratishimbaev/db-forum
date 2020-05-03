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

func (r *repository) GetPostCount(forumSlug string) (postCount uint64, err error) {
	countPosts := `
		SELECT COUNT(*) FROM post p
		JOIN forum f ON p.forum = f.id
		WHERE LOWER(f.slug) = LOWER($1)`
	err = r.db.QueryRow(countPosts, forumSlug).Scan(&postCount)

	return postCount, err
}

func (r *repository) GetThreadCount(forumSlug string) (threadCount uint64, err error) {
	countThreads := `
		SELECT COUNT(*) FROM thread t
		JOIN forum f ON t.forum = f.id
		WHERE LOWER(f.slug) = LOWER($1)`
	err = r.db.QueryRow(countThreads, forumSlug).Scan(&threadCount)

	return threadCount, err
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
	getForum := `SELECT f.title, u.nickname, f.slug
				 FROM forum f
				 LEFT JOIN "user" u ON f.user = u.id
				 WHERE LOWER(f.slug) = LOWER($1)`
	err = r.db.QueryRow(getForum, slug).Scan(&forum.Title, &forum.User, &forum.Slug)
	if err != nil {
		return forum, _forum.ErrNotFound
	}

	forum.Posts, err = r.GetPostCount(forum.Slug)
	forum.Threads, err = r.GetThreadCount(forum.Slug)

	return forum, err
}

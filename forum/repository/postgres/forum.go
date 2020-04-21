package forumPostgres

import (
	"database/sql"
	_forum "forum/forum"
	"forum/models"
)

type Repository struct {
	DB *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}

type Forum struct {
	ID    uint64
	Slug  string
	Title string
	User  uint64
}

func (r *Repository) toPostgres(forum *models.Forum) *Forum {
	var userID uint64
	getUserID := `SELECT id FROM "user" WHERE LOWER(nickname) = LOWER($1)`
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
		Slug:  forum.Slug,
		Title: forum.Title,
		User:  userNickname,
	}
}

func (r *Repository) GetPostCount(forumSlug string) (postCount uint64, err error) {
	countPosts := `
		SELECT COUNT(*) FROM post p
		JOIN forum f ON p.forum = f.id
		WHERE LOWER(f.slug) = LOWER($1)`
	err = r.DB.QueryRow(countPosts, forumSlug).Scan(&postCount)

	return postCount, err
}

func (r *Repository) GetThreadCount(forumSlug string) (threadCount uint64, err error) {
	countThreads := `
		SELECT COUNT(*) FROM thread t
		JOIN forum f ON t.forum = f.id
		WHERE LOWER(f.slug) = LOWER($1)`
	err = r.DB.QueryRow(countThreads, forumSlug).Scan(&threadCount)

	return threadCount, err
}

func (r *Repository) CreateForum(newForum *models.Forum) (forum models.Forum, err error) {
	pgForum := r.toPostgres(newForum)

	createForum := `
		INSERT INTO forum (slug, title, "user")
		VALUES ($1, $2, $3)`
	_, err = r.DB.Exec(createForum, pgForum.Slug, pgForum.Title, pgForum.User)

	if err != nil {
		var userCount uint64

		getUserCount := `SELECT COUNT(*) FROM "user" WHERE LOWER(nickname) = LOWER($1)`
		err = r.DB.QueryRow(getUserCount, newForum.User).Scan(&userCount)
		if err != nil {
			return forum, err
		}

		if userCount == 0 {
			return forum, _forum.NewUserNotFound(newForum.User)
		} else {
			forum, err = r.GetForum(newForum.Slug)
			if err != nil {
				return forum, err
			}

			return forum, _forum.NewAlreadyExits()
		}
	}

	forum, err = r.GetForum(newForum.Slug)
	if err != nil {
		return forum, err
	}

	return forum, err
}

func (r *Repository) GetForum(slug string) (forum models.Forum, err error) {
	var pgForum Forum

	getForum := `SELECT title, "user", slug
				 FROM forum WHERE LOWER(slug) = LOWER($1)`
	err = r.DB.QueryRow(getForum, slug).Scan(&pgForum.Title, &pgForum.User, &pgForum.Slug)
	if err != nil {
		return forum, _forum.NewNotFound(slug)
	}

	forum = *r.toModel(&pgForum)

	forum.Posts, err = r.GetPostCount(pgForum.Slug)
	forum.Threads, err = r.GetThreadCount(pgForum.Slug)

	return forum, err
}

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

func (r *Repository) GetPostCount(forumID uint64) (postCount uint64, err error) {
	countPosts := `SELECT COUNT(*) FROM post WHERE forum = $1`
	err = r.DB.QueryRow(countPosts, forumID).Scan(&postCount)

	return postCount, err
}

func (r *Repository) GetThreadCount(forumID uint64) (threadCount uint64, err error) {
	countThreads := `SELECT COUNT(*) FROM thread WHERE forum = $1`
	err = r.DB.QueryRow(countThreads, forumID).Scan(&threadCount)

	return threadCount, err
}

func (r *Repository) CreateForum(newForum *models.Forum) (forum models.Forum, err error) {
	pgForum := r.toPostgres(newForum)

	createForum := `INSERT INTO forum (slug, title, "user")
					VALUES ($1, $2, $3) RETURNING id`
	err = r.DB.QueryRow(createForum, pgForum.Slug, pgForum.Title, pgForum.User).Scan(&pgForum.ID)
	if err != nil {
		return forum, err
	}

	forum = *r.toModel(pgForum)

	forum.Posts, err = r.GetPostCount(pgForum.ID)
	forum.Threads, err = r.GetThreadCount(pgForum.ID)

	return forum, err
}

func (r *Repository) GetForum(slug string) (forum models.Forum, err error) {
	pgForum := Forum{Slug: slug}

	getForum := `SELECT id, title, "user"
				 FROM forum WHERE slug = $1`
	err = r.DB.QueryRow(getForum, pgForum.Slug).Scan(&pgForum.ID, &pgForum.Title, &pgForum.User)
	if err != nil {
		return forum, err
	}

	forum = *r.toModel(&pgForum)

	forum.Posts, err = r.GetPostCount(pgForum.ID)
	forum.Threads, err = r.GetThreadCount(pgForum.ID)

	return forum, err
}

package threadPostgres

import (
	"database/sql"
	"forum/models"
	"time"
)

type Repository struct {
	DB *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}

type Thread struct {
	ID uint64
	Author uint64
	Created time.Time
	Forum uint64
	Message string
	Slug string
	Title string
}

func (r *Repository) toPostgres(thread *models.Thread) *Thread {
	var authorID uint64
	getAuthorID := `SELECT id FROM "user" WHERE nickname = $1`
	if err := r.DB.QueryRow(getAuthorID, thread.Author).Scan(&authorID); err != nil {
		authorID = 0
	}

	var forumID uint64
	getForumID := `SELECT id FROM forum WHERE slug = $1`
	if err := r.DB.QueryRow(getForumID, thread.Forum).Scan(&forumID); err != nil {
		forumID = 0
	}

	return &Thread{
		Author:  authorID,
		Created: thread.Created,
		Forum:   forumID,
		Message: thread.Message,
		Slug:    thread.Slug,
		Title:   thread.Title,
	}
}

func (r *Repository) toModel(thread *Thread) *models.Thread {
	var authorNickname string
	getAuthorNickname := `SELECT nickname FROM "user" WHERE id = $1`
	if err := r.DB.QueryRow(getAuthorNickname, thread.Author).Scan(&authorNickname); err != nil {
		authorNickname = ""
	}

	var forumSlug string
	getForumSlug := `SELECT slug FROM forum WHERE id = $1`
	if err := r.DB.QueryRow(getForumSlug, thread.Forum).Scan(&forumSlug); err != nil {
		forumSlug = ""
	}

	return &models.Thread{
		Author:  authorNickname,
		Created: thread.Created,
		Forum:   forumSlug,
		ID:      thread.ID,
		Message: thread.Message,
		Slug:    thread.Slug,
		Title:   thread.Title,
	}
}

func (r *Repository) CreateThread(newThread *models.Thread) (thread models.Thread, err error) {
	pgThread := r.toPostgres(newThread)

	createThread := `INSERT INTO thread (author, created, forum, message, slug, title)
					 VALUES ($1, $2, $3, $4, $5, $6)`
	_, err = r.DB.Exec(createThread, pgThread.Author, pgThread.Created, pgThread.Forum,
						  pgThread.Message, pgThread.Slug, pgThread.Title)

	// TODO: votes field

	return *newThread, err
}

func (r *Repository) GetThreads(slug string, limit uint64, since string, desc bool) (threads []models.Thread, err error) {
	var forumID uint64
	getForumID := `SELECT id FROM forum WHERE slug = $1`
	if err := r.DB.QueryRow(getForumID, slug).Scan(&forumID); err != nil {
		forumID = 0
	}

	getThreads := `SELECT id, author, created, message, slug, title
				   FROM thread WHERE forum = $1`
	rows, err := r.DB.Query(getThreads, forumID)
	if err != nil {
		return threads, err
	}

	// TODO: limit, since and desc params

	for rows.Next() {
		var pgThread Thread

		if err = rows.Scan(&pgThread.ID, &pgThread.Author, &pgThread.Created,
						&pgThread.Message, &pgThread.Slug, &pgThread.Title); err != nil {
			return threads, err
		}

		// TODO: votes field

		threads = append(threads, *r.toModel(&pgThread))
	}

	return threads, err
}

func (r *Repository) GetThread(slugOrID string) (thread models.Thread, err error) {
	var pgThread Thread

	getThreadByID := `SELECT id, author, created, message, slug, title
				   	  FROM thread WHERE id = $1`
	getThreadBySlug := `SELECT id, author, created, message, slug, title
				   		FROM thread WHERE slug = $1`

	if err = r.DB.QueryRow(getThreadByID, slugOrID).
				  Scan(&pgThread.ID, &pgThread.Author, &pgThread.Created,
				  	   &pgThread.Message, &pgThread.Slug, &pgThread.Title); err != nil {
		err = r.DB.QueryRow(getThreadBySlug, slugOrID).
				   Scan(&pgThread.ID, &pgThread.Author, &pgThread.Created,
						&pgThread.Message, &pgThread.Slug, &pgThread.Title)
	}

	return *r.toModel(&pgThread), err
}

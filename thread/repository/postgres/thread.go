package threadPostgres

import (
	"database/sql"
	"forum/models"
	"strconv"
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

type Vote struct {
	User uint64
	Voice int64
	Thread uint64
}

func (r *Repository) toPostgresThread(thread *models.Thread) *Thread {
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

func (r *Repository) toModelThread(thread *Thread) *models.Thread {
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

func (r *Repository) toPostgresVote(vote *models.Vote) *Vote {
	var userID uint64
	getUserID := `SELECT id FROM "user" WHERE nickname = $1`
	if err := r.DB.QueryRow(getUserID, vote.Nickname).Scan(&userID); err != nil {
		userID = 0
	}

	return &Vote{
		User:   userID,
		Voice:  vote.Voice,
	}
}

func (r *Repository) CreateThread(newThread *models.Thread) (thread models.Thread, err error) {
	pgThread := r.toPostgresThread(newThread)

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

		threads = append(threads, *r.toModelThread(&pgThread))
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

	return *r.toModelThread(&pgThread), err
}

func (r *Repository) ChangeThread(slugOrID string, newThread *models.Thread) (thread models.Thread, err error) {
	var threadID uint64

	var isThreadID bool
	checkThreadID := `SELECT COUNT(*) <> 0 FROM thread WHERE id = $1`
	err = r.DB.QueryRow(checkThreadID, slugOrID).Scan(&isThreadID)

	if isThreadID {
		if threadID, err = strconv.ParseUint(slugOrID, 10, 64); err != nil {
			return thread, err
		}
	} else {
		getThreadID := `SELECT id FROM thread WHERE slug = $1`
		if err = r.DB.QueryRow(getThreadID, slugOrID).Scan(&threadID); err != nil {
			return thread, err
		}
	}

	changeThread := `UPDATE thread
					 SET message = $1, title = $2
					 WHERE id = $3`
	_, err = r.DB.Exec(changeThread, newThread.Message, newThread.Title, threadID)

	var pgThread Thread

	getThread := `SELECT id, author, created, forum, message, slug, title
				  FROM thread WHERE id = $1`
	err = r.DB.QueryRow(getThread, threadID).
			   Scan(&pgThread.ID, &pgThread.Author, &pgThread.Created, &pgThread.Forum, &pgThread.Message, &pgThread.Slug, &pgThread.Title)

	return *r.toModelThread(&pgThread), err
}

func (r *Repository) VoteThread(slugOrID string, vote models.Vote) (thread models.Thread, err error) {
	var threadID uint64

	var isThreadID bool
	checkThreadID := `SELECT COUNT(*) <> 0 FROM thread WHERE id = $1`
	err = r.DB.QueryRow(checkThreadID, slugOrID).Scan(&isThreadID)

	if isThreadID {
		if threadID, err = strconv.ParseUint(slugOrID, 10, 64); err != nil {
			return thread, err
		}
	} else {
		getThreadID := `SELECT id FROM thread WHERE slug = $1`
		if err = r.DB.QueryRow(getThreadID, slugOrID).Scan(&threadID); err != nil {
			return thread, err
		}
	}

	pgVote := *r.toPostgresVote(&vote)

	createVote := `INSERT INTO vote ("user", voice, thread)
				   VALUES ($1, $2, $3)`
	_, err = r.DB.Exec(createVote, pgVote.User, pgVote.Voice, threadID)

	var pgThread Thread

	getThread := `SELECT id, author, created, forum, message, slug, title
				  FROM thread WHERE id = $1`
	err = r.DB.QueryRow(getThread, threadID).
		Scan(&pgThread.ID, &pgThread.Author, &pgThread.Created, &pgThread.Forum, &pgThread.Message, &pgThread.Slug, &pgThread.Title)

	return *r.toModelThread(&pgThread), err
}

package threadPostgres

import (
	"database/sql"
	"forum/models"
	_thread "forum/thread"
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
	ID      uint64
	Author  uint64
	Created time.Time
	Forum   uint64
	Message string
	Slug    string
	Title   string
}

type Vote struct {
	User   uint64
	Voice  int64
	Thread uint64
}

func (r *Repository) toPostgresThread(thread *models.Thread) (pgThread Thread, err error) {
	var authorID uint64
	getAuthorID := `SELECT id FROM "user" WHERE LOWER(nickname) = LOWER($1)`
	if err := r.DB.QueryRow(getAuthorID, thread.Author).Scan(&authorID); err != nil {
		return pgThread, _thread.UserOrForumNotFound
	}

	var forumID uint64
	getForumID := `SELECT id FROM forum WHERE LOWER(slug) = LOWER($1)`
	if err := r.DB.QueryRow(getForumID, thread.Forum).Scan(&forumID); err != nil {
		return pgThread, _thread.UserOrForumNotFound
	}

	return Thread{
		Author:  authorID,
		Created: thread.Created,
		Forum:   forumID,
		Message: thread.Message,
		Slug:    thread.Slug,
		Title:   thread.Title,
	}, err
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
		User:  userID,
		Voice: vote.Voice,
	}
}

func (r *Repository) CreateThread(newThread *models.Thread) (thread models.Thread, err error) {
	pgThread, err := r.toPostgresThread(newThread)
	if err != nil {
		return thread, err
	}

	createThread := `
		INSERT INTO thread (author, created, forum, message, slug, title)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	err = r.DB.QueryRow(createThread, pgThread.Author, pgThread.Created, pgThread.Forum, pgThread.Message, pgThread.Slug, pgThread.Title).
		Scan(&pgThread.ID)

	if err != nil {
		thread, err = r.GetThreadBySlug(pgThread.Slug)
		if err != nil {
			return thread, err
		}

		return thread, _thread.AlreadyExists
	}

	thread, err = r.GetThreadByID(pgThread.ID)
	if err != nil {
		return thread, err
	}

	return thread, err
}

func (r *Repository) GetThreads(slug string, limit uint64, since time.Time, desc bool) (threads []models.Thread, err error) {
	var forumID uint64
	getForumID := `SELECT id FROM forum WHERE LOWER(slug) = LOWER($1)`
	if err := r.DB.QueryRow(getForumID, slug).Scan(&forumID); err != nil {
		return threads, _thread.UserOrForumNotFound
	}

	var getThreads string

	if desc {
		getThreads = `
			SELECT id, author, created, message, slug, title
			FROM thread WHERE forum = $1 AND created <= $2 ORDER BY created DESC`

		if since == (time.Time{}) {
			since = time.Now().AddDate(1000, 0, 0)
		}
	} else {
		getThreads = `
			SELECT id, author, created, message, slug, title
			FROM thread WHERE forum = $1 AND created >= $2 ORDER BY created`

		if since == (time.Time{}) {
			since = time.Now().AddDate(-1000, 0, 0)
		}
	}

	var rows *sql.Rows

	if limit == 0 {
		rows, err = r.DB.Query(getThreads, forumID, since)
	} else {
		getThreads = getThreads + " LIMIT $3"

		rows, err = r.DB.Query(getThreads, forumID, since, limit)
	}

	if err != nil {
		return threads, err
	}

	var forumSlug string

	getForumSlug := `SELECT slug FROM forum WHERE LOWER(slug) = LOWER($1)`
	err = r.DB.QueryRow(getForumSlug, slug).Scan(&forumSlug)
	if err != nil {
		return threads, err
	}

	for rows.Next() {
		var pgThread Thread

		if err = rows.Scan(&pgThread.ID, &pgThread.Author, &pgThread.Created,
			&pgThread.Message, &pgThread.Slug, &pgThread.Title); err != nil {
			return threads, err
		}

		var voteCount uint64

		getVotes := `SELECT COUNT(*) FROM vote WHERE thread = $1`
		err = r.DB.QueryRow(getVotes, pgThread.ID).Scan(&voteCount)

		thread := *r.toModelThread(&pgThread)
		thread.Votes = voteCount
		thread.Forum = forumSlug

		threads = append(threads, thread)
	}

	if len(threads) == 0 {
		return []models.Thread{}, err
	}

	return threads, err
}

func (r *Repository) GetThreadByID(id uint64) (thread models.Thread, err error) {
	var pgThread Thread

	getThread := `
		SELECT id, author, created, message, slug, title, forum
		FROM thread WHERE id = $1`
	if err = r.DB.QueryRow(getThread, id).
		Scan(&pgThread.ID, &pgThread.Author, &pgThread.Created, &pgThread.Message, &pgThread.Slug, &pgThread.Title, &pgThread.Forum); err != nil {
		return thread, _thread.NotFound
	}

	thread = *r.toModelThread(&pgThread)

	getVotes := `SELECT SUM(voice) FROM vote WHERE thread = $1`
	_ = r.DB.QueryRow(getVotes, pgThread.ID).Scan(&thread.Votes)

	return thread, err
}

func (r *Repository) GetThreadBySlug(slug string) (thread models.Thread, err error) {
	var pgThread Thread

	getThread := `
		SELECT id, author, created, message, slug, title, forum
		FROM thread WHERE LOWER(slug) = LOWER($1)`
	if err = r.DB.QueryRow(getThread, slug).
		Scan(&pgThread.ID, &pgThread.Author, &pgThread.Created, &pgThread.Message, &pgThread.Slug, &pgThread.Title, &pgThread.Forum); err != nil {
		return thread, _thread.NotFound
	}

	thread = *r.toModelThread(&pgThread)

	getVotes := `SELECT COUNT(*) FROM vote WHERE thread = $1`
	_ = r.DB.QueryRow(getVotes, pgThread.ID).Scan(&thread.Votes)

	return thread, err
}

func (r *Repository) GetThread(slugOrID string) (thread models.Thread, err error) {
	if thread, err = r.GetThreadBySlug(slugOrID); err != nil {
		id, err := strconv.ParseUint(slugOrID, 10, 64)
		if err != nil {
			return thread, _thread.NotFound
		}

		if thread, err = r.GetThreadByID(id); err != nil {
			return thread, _thread.NotFound
		}

		return thread, err
	}

	return thread, err
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
		getThreadID := `SELECT id FROM thread WHERE LOWER(slug) = LOWER($1)`
		if err = r.DB.QueryRow(getThreadID, slugOrID).Scan(&threadID); err != nil {
			return thread, _thread.NotFound
		}
	}

	var oldThread Thread

	getOldThread := `SELECT message, title FROM thread WHERE id = $1`
	err = r.DB.QueryRow(getOldThread, threadID).Scan(&oldThread.Message, &oldThread.Title)
	if err != nil {
		return thread, err
	}

	if !(newThread.Message == "" && newThread.Title == "") {
		if newThread.Message == "" {
			newThread.Message = oldThread.Message
		}
		if newThread.Title == "" {
			newThread.Title = oldThread.Title
		}

		changeThread := `
			UPDATE thread
			SET message = $1, title = $2
			WHERE id = $3`
		_, err = r.DB.Exec(changeThread, newThread.Message, newThread.Title, threadID)
	}

	var pgThread Thread

	getThread := `
		SELECT id, author, created, forum, message, slug, title
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
		getThreadID := `SELECT id FROM thread WHERE LOWER(slug) = LOWER($1)`
		if err = r.DB.QueryRow(getThreadID, slugOrID).Scan(&threadID); err != nil {
			return thread, _thread.NotFound
		}
	}

	pgVote := *r.toPostgresVote(&vote)

	checkUser := `SELECT COUNT(*) <> 0 FROM "user" WHERE LOWER(nickname) = LOWER($1)`

	var hasUser bool

	err = r.DB.QueryRow(checkUser, vote.Nickname).Scan(&hasUser)
	if err != nil || !hasUser {
		return thread, _thread.NotFound
	}

	createVote := `INSERT INTO vote ("user", voice, thread) VALUES ($1, $2, $3)`
	if _, err = r.DB.Exec(createVote, pgVote.User, pgVote.Voice, threadID); err != nil {
		changeVote := `UPDATE vote SET voice = $1 WHERE "user" = $2 AND thread = $3`
		if _, err = r.DB.Exec(changeVote, pgVote.Voice, pgVote.User, threadID); err != nil {
			return thread, err
		}
	}

	return r.GetThreadByID(threadID)
}

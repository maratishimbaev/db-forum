package threadPostgres

import (
	"database/sql"
	"forum/models"
	_thread "forum/thread"
	"strconv"
	"time"
)

type repository struct {
	db *sql.DB
}

func NewThreadRepository(db *sql.DB) *repository {
	return &repository{db: db}
}

func (r *repository) CreateThread(newThread *models.Thread) (*models.Thread, error) {
	var authorNickname string
	getAuthorNickname := `SELECT nickname FROM "user" WHERE LOWER(nickname) = LOWER($1)`
	err := r.db.QueryRow(getAuthorNickname, newThread.Author).Scan(&authorNickname)
	if err != nil {
		return nil, _thread.UserOrForumNotFound
	}

	var forumSlug string
	getForumSlug := `SELECT slug FROM forum WHERE LOWER(slug) = LOWER($1)`
	err = r.db.QueryRow(getForumSlug, newThread.Forum).Scan(&forumSlug)
	if err != nil {
		return nil, _thread.UserOrForumNotFound
	}

	if newThread.Slug != "" {
		thread, err := r.GetThreadBySlug(newThread.Slug)
		if err == nil {
			return thread, _thread.AlreadyExists
		}
	}

	createThread := `
		INSERT INTO thread (author, created, forum, message, slug, title)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	err = r.db.QueryRow(createThread, authorNickname, newThread.Created, forumSlug, newThread.Message, newThread.Slug, newThread.Title).
		Scan(&newThread.ID)

	thread, err := r.GetThreadByID(newThread.ID)
	if err != nil {
		return nil, err
	}

	return thread, err
}

func (r *repository) GetThreads(slug string, limit uint64, since time.Time, desc bool) (threads []models.Thread, err error) {
	var forumExists bool
	checkForum := `SELECT EXISTS(SELECT 1 FROM forum WHERE LOWER(slug) = LOWER($1))`
	err = r.db.QueryRow(checkForum, slug).Scan(&forumExists)
	if err != nil || !forumExists {
		return nil, _thread.UserOrForumNotFound
	}

	getThreads := `
		SELECT id, author, created, message, slug, title, forum, votes
		FROM thread
		WHERE forum = $1`

	if desc {
		if since != (time.Time{}) {
			getThreads += ` AND created <= $2`
		}
		getThreads += ` ORDER BY created DESC`
	} else {
		if since != (time.Time{}) {
			getThreads += ` AND created >= $2`
		}
		getThreads += ` ORDER BY created`
	}

	var rows *sql.Rows

	if limit == 0 {
		if since != (time.Time{}) {
			rows, err = r.db.Query(getThreads, slug, since)
		} else {
			rows, err = r.db.Query(getThreads, slug)
		}
	} else {
		if since != (time.Time{}) {
			getThreads += `	LIMIT $3`
			rows, err = r.db.Query(getThreads, slug, since, limit)
		} else {
			getThreads += `	LIMIT $2`
			rows, err = r.db.Query(getThreads, slug, limit)
		}
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var thread models.Thread

		if err = rows.Scan(&thread.ID, &thread.Author, &thread.Created, &thread.Message, &thread.Slug, &thread.Title,
			&thread.Forum, &thread.Votes); err != nil {
			return nil, err
		}

		threads = append(threads, thread)
	}

	if len(threads) == 0 {
		return []models.Thread{}, nil
	}

	return threads, nil
}

func (r *repository) GetThreadByID(id uint64) (*models.Thread, error) {
	var thread models.Thread

	getThread := `
		SELECT id, author, created, message, slug, title, forum, votes
		FROM thread
		WHERE id = $1`
	if err := r.db.QueryRow(getThread, id).
		Scan(&thread.ID, &thread.Author, &thread.Created, &thread.Message, &thread.Slug, &thread.Title, &thread.Forum,
			&thread.Votes); err != nil {
		return nil, _thread.NotFound
	}

	return &thread, nil
}

func (r *repository) GetThreadBySlug(slug string) (*models.Thread, error) {
	var thread models.Thread

	getThread := `
		SELECT id, author, created, message, slug, title, forum, votes
		FROM thread
		WHERE slug = $1`
	if err := r.db.QueryRow(getThread, slug).
		Scan(&thread.ID, &thread.Author, &thread.Created, &thread.Message, &thread.Slug, &thread.Title, &thread.Forum,
			&thread.Votes); err != nil {
		return nil, _thread.NotFound
	}

	return &thread, nil
}

func (r *repository) GetThread(slugOrID string) (*models.Thread, error) {
	thread, err := r.GetThreadBySlug(slugOrID)
	if err != nil {
		id, err := strconv.ParseUint(slugOrID, 10, 64)
		if err != nil {
			return nil, _thread.NotFound
		}

		if thread, err = r.GetThreadByID(id); err != nil {
			return nil, _thread.NotFound
		}

		return thread, err
	}

	return thread, nil
}

func (r *repository) ChangeThread(slugOrID string, newThread *models.Thread) (*models.Thread, error) {
	var threadID uint64

	var isThreadID bool
	checkThreadID := `SELECT EXISTS(SELECT 1 FROM thread WHERE id = $1)`
	err := r.db.QueryRow(checkThreadID, slugOrID).Scan(&isThreadID)

	if isThreadID {
		if threadID, err = strconv.ParseUint(slugOrID, 10, 64); err != nil {
			return nil, err
		}
	} else {
		getThreadID := `SELECT id FROM thread WHERE LOWER(slug) = LOWER($1)`
		if err = r.db.QueryRow(getThreadID, slugOrID).Scan(&threadID); err != nil {
			return nil, _thread.NotFound
		}
	}

	var oldMessage, oldTitle string

	getOldThread := `SELECT message, title FROM thread WHERE id = $1`
	err = r.db.QueryRow(getOldThread, threadID).Scan(&oldMessage, &oldTitle)
	if err != nil {
		return nil, err
	}

	if !(newThread.Message == "" && newThread.Title == "") {
		if newThread.Message == "" {
			newThread.Message = oldMessage
		}
		if newThread.Title == "" {
			newThread.Title = oldTitle
		}

		changeThread := `
			UPDATE thread
			SET message = $1, title = $2
			WHERE id = $3`
		_, err = r.db.Exec(changeThread, newThread.Message, newThread.Title, threadID)
	}

	var thread models.Thread

	getThread := `
		SELECT t.id, t.author, t.created, t.forum, t.message, t.slug, t.title
		FROM thread t
		WHERE t.id = $1`
	err = r.db.QueryRow(getThread, threadID).
		Scan(&thread.ID, &thread.Author, &thread.Created, &thread.Forum, &thread.Message, &thread.Slug, &thread.Title)

	return &thread, err
}

func (r *repository) VoteThread(slugOrID string, vote models.Vote) (*models.Thread, error) {
	var threadID uint64
	var isID bool

	threadID, err := strconv.ParseUint(slugOrID, 10, 64)
	if err != nil {
		threadID = 0
	}

	threadExists := `SELECT EXISTS(SELECT 1 FROM thread WHERE id = $1)`
	if err = r.db.QueryRow(threadExists, threadID).Scan(&isID); err != nil {
		return nil, err
	}
	if !isID {
		getThreadID := `SELECT id FROM thread WHERE LOWER(slug) = LOWER($1)`
		if err = r.db.QueryRow(getThreadID, slugOrID).Scan(&threadID); err != nil {
			return nil, _thread.NotFound
		}
	}

	createOrUpdateVote := `
		INSERT INTO vote (voice, "user", thread)
		SELECT $1, id, $3 FROM "user" WHERE LOWER(nickname) = LOWER($2)
		ON CONFLICT ON CONSTRAINT unique_user_and_thread DO
		UPDATE SET voice = $1`
	res, err := r.db.Exec(createOrUpdateVote, vote.Voice, vote.Nickname, threadID)
	if err != nil {
		return nil, _thread.NotFound
	}
	count, err := res.RowsAffected()
	if err != nil || count == 0 {
		return nil, _thread.NotFound
	}

	var thread models.Thread

	getThread := `
		SELECT t.id, t.author, t.created, t.message, t.slug, t.title, t.forum, t.votes
		FROM thread t
		WHERE t.id = $1`
	if err = r.db.QueryRow(getThread, threadID).
		Scan(&thread.ID, &thread.Author, &thread.Created, &thread.Message, &thread.Slug, &thread.Title, &thread.Forum,
			&thread.Votes); err != nil {
		return nil, _thread.NotFound
	}

	return &thread, nil
}

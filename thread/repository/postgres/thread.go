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

func (r *repository) CreateThread(newThread *models.Thread) (thread models.Thread, err error) {
	var authorID uint64
	getAuthorID := `SELECT id FROM "user" WHERE LOWER(nickname) = LOWER($1)`
	err = r.db.QueryRow(getAuthorID, newThread.Author).Scan(&authorID)
	if err != nil {
		return thread, _thread.UserOrForumNotFound
	}

	var forumID uint64
	getForumID := `SELECT id FROM forum WHERE LOWER(slug) = LOWER($1)`
	err = r.db.QueryRow(getForumID, newThread.Forum).Scan(&forumID)
	if err != nil {
		return thread, _thread.UserOrForumNotFound
	}

	createThread := `
		INSERT INTO thread (author, created, forum, message, slug, title)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	err = r.db.QueryRow(createThread, authorID, newThread.Created, forumID, newThread.Message, newThread.Slug, newThread.Title).
		Scan(&newThread.ID)

	if err != nil {
		thread, err = r.GetThreadBySlug(newThread.Slug)
		if err != nil {
			return thread, err
		}

		return thread, _thread.AlreadyExists
	}

	thread, err = r.GetThreadByID(newThread.ID)
	if err != nil {
		return thread, err
	}

	return thread, err
}

func (r *repository) GetThreads(slug string, limit uint64, since time.Time, desc bool) (threads []models.Thread, err error) {
	var forumID uint64
	getForumID := `SELECT id FROM forum WHERE LOWER(slug) = LOWER($1)`
	if err := r.db.QueryRow(getForumID, slug).Scan(&forumID); err != nil {
		return threads, _thread.UserOrForumNotFound
	}

	var getThreads string

	if desc {
		getThreads = `
			SELECT t.id, u.nickname, t.created, t.message, t.slug, t.title, f.slug, COALESCE(SUM(v.voice), 0)
			FROM thread t
			JOIN "user" u ON t.author = u.id
			JOIN forum f ON t.forum = f.id
			LEFT JOIN vote v ON t.id = v.thread
			WHERE t.forum = $1 AND t.created <= $2
			GROUP BY t.id, u.nickname, f.slug
			ORDER BY t.created DESC`

		if since == (time.Time{}) {
			since = time.Now().AddDate(1000, 0, 0)
		}
	} else {
		getThreads = `
			SELECT t.id, u.nickname, t.created, t.message, t.slug, t.title, f.slug, COALESCE(SUM(v.voice), 0)
			FROM thread t
			JOIN "user" u ON t.author = u.id
			JOIN forum f ON t.forum = f.id
			LEFT JOIN vote v ON t.id = v.thread
			WHERE t.forum = $1 AND t.created >= $2
			GROUP BY t.id, u.nickname, f.slug
			ORDER BY t.created`

		if since == (time.Time{}) {
			since = time.Now().AddDate(-1000, 0, 0)
		}
	}

	var rows *sql.Rows

	if limit == 0 {
		rows, err = r.db.Query(getThreads, forumID, since)
	} else {
		getThreads = getThreads + " LIMIT $3"

		rows, err = r.db.Query(getThreads, forumID, since, limit)
	}
	if err != nil {
		return threads, err
	}
	defer rows.Close()

	for rows.Next() {
		var thread models.Thread

		if err = rows.Scan(&thread.ID, &thread.Author, &thread.Created, &thread.Message, &thread.Slug, &thread.Title,
			&thread.Forum, &thread.Votes); err != nil {
			return threads, err
		}

		threads = append(threads, thread)
	}

	if len(threads) == 0 {
		return []models.Thread{}, err
	}

	return threads, err
}

func (r *repository) GetThreadByID(id uint64) (thread models.Thread, err error) {
	getThread := `
		SELECT t.id, u.nickname, t.created, t.message, t.slug, t.title, f.slug, t.votes
		FROM thread t
		JOIN "user" u ON t.author = u.id
		JOIN forum f ON t.forum = f.id
		WHERE t.id = $1`
	if err = r.db.QueryRow(getThread, id).
		Scan(&thread.ID, &thread.Author, &thread.Created, &thread.Message, &thread.Slug, &thread.Title, &thread.Forum,
			&thread.Votes); err != nil {
		return thread, _thread.NotFound
	}

	return thread, err
}

func (r *repository) GetThreadBySlug(slug string) (thread models.Thread, err error) {
	getThread := `
		SELECT t.id, u.nickname, t.created, t.message, t.slug, t.title, f.slug, t.votes
		FROM thread t
		JOIN "user" u ON t.author = u.id
		JOIN forum f ON t.forum = f.id
		WHERE LOWER(t.slug) = LOWER($1)`
	if err = r.db.QueryRow(getThread, slug).
		Scan(&thread.ID, &thread.Author, &thread.Created, &thread.Message, &thread.Slug, &thread.Title, &thread.Forum,
			&thread.Votes); err != nil {
		return thread, _thread.NotFound
	}

	return thread, err
}

func (r *repository) GetThread(slugOrID string) (thread models.Thread, err error) {
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

func (r *repository) ChangeThread(slugOrID string, newThread *models.Thread) (thread models.Thread, err error) {
	var threadID uint64

	var isThreadID bool
	checkThreadID := `SELECT COUNT(*) <> 0 FROM thread WHERE id = $1`
	err = r.db.QueryRow(checkThreadID, slugOrID).Scan(&isThreadID)

	if isThreadID {
		if threadID, err = strconv.ParseUint(slugOrID, 10, 64); err != nil {
			return thread, err
		}
	} else {
		getThreadID := `SELECT id FROM thread WHERE LOWER(slug) = LOWER($1)`
		if err = r.db.QueryRow(getThreadID, slugOrID).Scan(&threadID); err != nil {
			return thread, _thread.NotFound
		}
	}

	var oldMessage, oldTitle string

	getOldThread := `SELECT message, title FROM thread WHERE id = $1`
	err = r.db.QueryRow(getOldThread, threadID).Scan(&oldMessage, &oldTitle)
	if err != nil {
		return thread, err
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

	getThread := `
		SELECT t.id, u.nickname, t.created, f.slug, t.message, t.slug, t.title
		FROM thread t
		JOIN "user" u ON t.author = u.id
		JOIN forum f ON t.forum = f.id
		WHERE t.id = $1`
	err = r.db.QueryRow(getThread, threadID).
		Scan(&thread.ID, &thread.Author, &thread.Created, &thread.Forum, &thread.Message, &thread.Slug, &thread.Title)

	return thread, err
}

func (r *repository) VoteThread(slugOrID string, vote models.Vote) (thread models.Thread, err error) {
	var threadID uint64
	var isID bool

	threadID, err = strconv.ParseUint(slugOrID, 10, 64)
	if err != nil {
		threadID = 0
	}

	threadExists := `SELECT EXISTS(SELECT 1 FROM thread WHERE id = $1)`
	if err = r.db.QueryRow(threadExists, threadID).Scan(&isID); err != nil {
		return thread, err
	}
	if !isID {
		getThreadID := `SELECT id FROM thread WHERE LOWER(slug) = LOWER($1)`
		if err = r.db.QueryRow(getThreadID, slugOrID).Scan(&threadID); err != nil {
			return thread, _thread.NotFound
		}
	}

	createOrUpdateVote := `
		INSERT INTO vote (voice, "user", thread)
		SELECT $1, id, $3 FROM "user" WHERE LOWER(nickname) = LOWER($2)
		ON CONFLICT ON CONSTRAINT unique_user_and_thread DO
		UPDATE SET voice = $1`
	res, err := r.db.Exec(createOrUpdateVote, vote.Voice, vote.Nickname, threadID)
	if err != nil {
		return thread, _thread.NotFound
	}
	count, err := res.RowsAffected()
	if err != nil || count == 0 {
		return thread, _thread.NotFound
	}

	getThread := `
		SELECT t.id, u.nickname, t.created, t.message, t.slug, t.title, f.slug, t.votes
		FROM thread t
		JOIN "user" u ON t.author = u.id
		JOIN forum f ON t.forum = f.id
		WHERE t.id = $1`
	if err = r.db.QueryRow(getThread, threadID).
		Scan(&thread.ID, &thread.Author, &thread.Created, &thread.Message, &thread.Slug, &thread.Title, &thread.Forum,
			&thread.Votes); err != nil {
		return thread, _thread.NotFound
	}

	return thread, err
}

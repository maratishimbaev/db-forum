package postPostgres

import (
	"database/sql"
	"errors"
	"forum/models"
	_post "forum/post"
	"forum/utils"
	"strconv"
	"strings"
	"time"
)

type repository struct {
	db *sql.DB
}

func NewPostRepository(db *sql.DB) *repository {
	return &repository{
		db: db,
	}
}

func contains(slice []string, searchable string) bool {
	for _, el := range slice {
		if el == searchable {
			return true
		}
	}
	return false
}

func (r *repository) GetPostFull(id uint64, related []string) (postFull models.PostFull, err error) {
	postFull.Post.ID = id

	getPost := `
		SELECT author, created, forum, is_edited, message, parent, thread
		FROM post
		WHERE id = $1`
	if err = r.db.QueryRow(getPost, id).
		Scan(&postFull.Post.Author, &postFull.Post.Created, &postFull.Post.Forum, &postFull.Post.IsEdited,
			&postFull.Post.Message, &postFull.Post.Parent, &postFull.Post.Thread); err != nil {
		return postFull, _post.NotFound
	}

	authorContains := contains(related, "user")
	if authorContains {
		postFull.Author = &models.User{}

		getAuthor := `
			SELECT about, email, fullname, nickname
			FROM "user" WHERE nickname = $1`
		if err = r.db.QueryRow(getAuthor, postFull.Post.Author).
			Scan(&postFull.Author.About, &postFull.Author.Email, &postFull.Author.FullName,
				&postFull.Author.Nickname); err != nil {
			return postFull, err
		}
	}

	forumContains := contains(related, "forum")
	if forumContains {
		postFull.Forum = &models.Forum{}

		getForum := `
			SELECT slug, title, "user", posts, threads
			FROM forum
			WHERE slug = $1`
		if err = r.db.QueryRow(getForum, postFull.Post.Forum).
			Scan(&postFull.Forum.Slug, &postFull.Forum.Title, &postFull.Forum.User, &postFull.Forum.Posts,
				&postFull.Forum.Threads); err != nil {
			return postFull, err
		}
	}

	threadContains := contains(related, "thread")
	if threadContains {
		postFull.Thread = &models.Thread{}

		getThread := `
			SELECT id, author, created, forum, message, slug, title, votes
			FROM thread
			WHERE id = $1`
		if err = r.db.QueryRow(getThread, postFull.Post.Thread).
			Scan(&postFull.Thread.ID, &postFull.Thread.Author, &postFull.Thread.Created, &postFull.Thread.Forum,
				&postFull.Thread.Message, &postFull.Thread.Slug, &postFull.Thread.Title, &postFull.Thread.Votes); err != nil {
			return postFull, err
		}
	}

	return postFull, err
}

func (r *repository) ChangePost(newPost *models.Post) (post models.Post, err error) {
	var hasPost bool

	checkPost := `SELECT EXISTS(SELECT 1 FROM post WHERE id = $1)`
	if err = r.db.QueryRow(checkPost, newPost.ID).Scan(&hasPost); err != nil || !hasPost {
		return post, _post.NotFound
	}

	getPostMessage := `SELECT message FROM post WHERE id = $1`

	var oldPostMessage string

	err = r.db.QueryRow(getPostMessage, newPost.ID).Scan(&oldPostMessage)
	if err != nil {
		return post, err
	}

	if oldPostMessage != newPost.Message && newPost.Message != "" {
		changePost := `UPDATE post SET message = $1, is_edited = true WHERE id = $2`
		if _, err = r.db.Exec(changePost, newPost.Message, newPost.ID); err != nil {
			return post, _post.NotFound
		}
	}

	getPost := `
		SELECT p.id, p.author, p.created, p.forum, p.is_edited, p.message, p.parent, p.thread
		FROM post p
		WHERE p.id = $1`
	err = r.db.QueryRow(getPost, newPost.ID).
		Scan(&post.ID, &post.Author, &post.Created, &post.Forum, &post.IsEdited, &post.Message, &post.Parent, &post.Thread)

	return post, err
}

func (r *repository) GetThreadID(threadSlugOrID string) (threadID uint64, err error) {
	getThreadIDByID := `SELECT id FROM thread WHERE id = $1`
	err = r.db.QueryRow(getThreadIDByID, threadSlugOrID).Scan(&threadID)
	if err != nil {
		getThreadIDBySlug := `SELECT id from thread WHERE LOWER(slug) = LOWER($1)`
		err = r.db.QueryRow(getThreadIDBySlug, threadSlugOrID).Scan(&threadID)
		if err != nil {
			return threadID, _post.ThreadNotFound
		}
	}

	return threadID, err
}

func (r *repository) CreatePosts(threadSlugOrID string, newPosts []models.Post) (posts []models.Post, err error) {
	tx, err := r.db.Begin()
	if err != nil {
		return posts, err
	}

	var forumSlug string

	threadId, err := strconv.ParseUint(threadSlugOrID, 10, 64)
	if err != nil {
		threadId = 0
	}

	getThread := `SELECT id, forum FROM thread WHERE id = $1 OR (slug <> '' AND slug = $2)`
	err = tx.QueryRow(getThread, threadId, threadSlugOrID).Scan(&threadId, &forumSlug)
	if err != nil {
		_ = tx.Rollback()
		return posts, _post.ThreadNotFound
	}

	if len(newPosts) == 0 {
		_ = tx.Commit()
		return []models.Post{}, err
	}

	now := time.Now()

	createPost := `INSERT INTO post (id, author, created, forum, is_edited, message, parent, thread, path) VALUES`
	var vals []interface{}

	for _, newPost := range newPosts {
		newPost.Thread = threadId

		if newPost.Parent != 0 {
			var parentThread uint64

			err = tx.QueryRow(`SELECT thread FROM post WHERE id = $1`, newPost.Parent).Scan(&parentThread)
			if err != nil || parentThread != newPost.Thread {
				_ = tx.Rollback()
				return posts, _post.ParentNotInThread
			}
			createPost += `
				(nextval('post_id_seq'::regclass),
				?, ?, ?, ?, ?, ?, ?,
				(SELECT path FROM post WHERE id = ?) || currval(pg_get_serial_sequence('post', 'id'))::integer),`
			vals = append(vals, newPost.Author, now, forumSlug, false, newPost.Message, newPost.Parent, newPost.Thread, newPost.Parent)
		} else {
			createPost += `
				(nextval('post_id_seq'::regclass),
				?, ?, ?, ?, ?, ?, ?,
				ARRAY[currval(pg_get_serial_sequence('post', 'id'))::integer]),`
			vals = append(vals, newPost.Author, now, forumSlug, false, newPost.Message, newPost.Parent, newPost.Thread)
		}
	}

	createPost = createPost[0 : len(createPost)-1]
	createPost += ` RETURNING id, author, created, forum, is_edited, message, parent, thread`
	createPost = utils.ReplaceSQL(createPost, "?", 1)

	statement, err := tx.Prepare(createPost)
	if err != nil {
		_ = tx.Rollback()
		return posts, errors.New("statement error: " + err.Error())
	}
	rows, err := statement.Query(vals...)
	if err != nil {
		_ = tx.Rollback()
		return posts, _post.NotFound
	}

	for rows.Next() {
		var post models.Post

		err = rows.Scan(&post.ID, &post.Author, &post.Created, &post.Forum, &post.IsEdited, &post.Message, &post.Parent, &post.Thread)
		if err != nil {
			_ = tx.Rollback()
			return posts, err
		}

		posts = append(posts, post)
	}

	_, err = tx.Exec(`UPDATE forum SET posts = posts + $1 WHERE slug = $2`, len(posts), forumSlug)
	if err != nil {
		_ = tx.Rollback()
		return posts, err
	}

	_ = tx.Commit()
	return posts, err
}

func (r *repository) GetFlatSortPosts(threadID, limit, since uint64, desc bool) (posts []models.Post, err error) {
	getPosts := `
		SELECT id, author, created, forum, is_edited, message, parent 
		FROM post
		WHERE thread = $1`

	if !desc {
		if since != 0 {
			getPosts += ` AND id > ?`
		}
		getPosts += ` ORDER BY id`
	} else {
		if since != 0 {
			getPosts += ` AND id < ?`
		}
		getPosts += ` ORDER BY id DESC`
	}

	if limit != 0 {
		getPosts += ` LIMIT ?`
	}

	getPosts = utils.ReplaceSQL(getPosts, "?", 2)

	var parentRows *sql.Rows

	switch true {
	case since != 0 && limit != 0:
		parentRows, err = r.db.Query(getPosts, threadID, since, limit)
		break
	case since != 0:
		parentRows, err = r.db.Query(getPosts, threadID, since)
		break
	case limit != 0:
		parentRows, err = r.db.Query(getPosts, threadID, limit)
		break
	default:
		parentRows, err = r.db.Query(getPosts, threadID)
	}

	if err != nil {
		return posts, err
	}
	defer parentRows.Close()

	for parentRows.Next() {
		post := models.Post{Thread: threadID}

		err = parentRows.Scan(&post.ID, &post.Author, &post.Created, &post.Forum, &post.IsEdited, &post.Message, &post.Parent)
		if err != nil {
			return posts, err
		}

		posts = append(posts, post)
	}

	if len(posts) == 0 {
		return []models.Post{}, err
	}

	return posts, err
}

func (r *repository) GetTreeSortPosts(threadID, limit, since uint64, desc bool) (posts []models.Post, err error) {
	getPosts := `
		SELECT p.id, p.author, p.created, p.forum, p.is_edited, p.message, p.parent
		FROM post p
		WHERE p.thread = $1`

	if !desc {
		if since != 0 {
			getPosts += ` AND path > (SELECT path FROM post WHERE id = $2)`
		}
		getPosts += ` ORDER BY path`
	} else {
		if since != 0 {
			getPosts += ` AND path < (SELECT path FROM post WHERE id = $2)`
		}
		getPosts += ` ORDER BY path DESC`
	}

	if limit != 0 {
		getPosts += ` LIMIT ?`
	}

	var startsWith uint64 = 2
	if since != 0 {
		startsWith = 3
	}
	getPosts = utils.ReplaceSQL(getPosts, "?", startsWith)

	var rows *sql.Rows
	switch true {
	case since != 0 && limit != 0:
		rows, err = r.db.Query(getPosts, threadID, since, limit)
		break
	case since != 0:
		rows, err = r.db.Query(getPosts, threadID, since)
		break
	case limit != 0:
		rows, err = r.db.Query(getPosts, threadID, limit)
		break
	default:
		rows, err = r.db.Query(getPosts, threadID)
	}
	if err != nil {
		return posts, err
	}
	defer rows.Close()

	for rows.Next() {
		post := models.Post{Thread: threadID}

		err = rows.Scan(&post.ID, &post.Author, &post.Created, &post.Forum, &post.IsEdited, &post.Message, &post.Parent)
		if err != nil {
			return posts, err
		}

		posts = append(posts, post)
	}

	if len(posts) == 0 {
		return []models.Post{}, err
	}

	return posts, err
}

func (r *repository) GetParentTreeSortPosts(threadID, limit, since uint64, desc bool) (posts []models.Post, err error) {
	getPosts := `
		SELECT id, author, created, forum, is_edited, message, parent
		FROM post
		WHERE thread = $1 AND path && (
			SELECT ARRAY (
				SELECT id
				FROM post
				WHERE thread = $1 AND parent = 0 since
				ORDER BY id desc
				limit
			)
		)
		ORDER BY path[1] desc, path`

	if since != 0 {
		if !desc {
			getPosts = strings.Replace(getPosts, "since", "AND path > (SELECT path[1:1] FROM post WHERE id = $2)", 1)
		} else {
			getPosts = strings.Replace(getPosts, "since", "AND path < (SELECT path[1:1] FROM post WHERE id = $2)", 1)
		}
	} else {
		getPosts = strings.Replace(getPosts, "since", "", 1)
	}

	if !desc {
		getPosts = strings.Replace(getPosts, "desc", "", 2)
	}

	if limit != 0 {
		getPosts = strings.Replace(getPosts, "limit", "LIMIT " + strconv.Itoa(int(limit)), 1)
	} else {
		getPosts = strings.Replace(getPosts, "limit", "", 1)
	}

	var rows *sql.Rows
	if since != 0 {
		rows, err = r.db.Query(getPosts, threadID, since)
	} else {
		rows, err = r.db.Query(getPosts, threadID)
	}
	if err != nil {
		return posts, err
	}
	defer rows.Close()

	for rows.Next() {
		post := models.Post{Thread: threadID}

		err = rows.Scan(&post.ID, &post.Author, &post.Created, &post.Forum, &post.IsEdited, &post.Message, &post.Parent)
		if err != nil {
			return posts, err
		}

		posts = append(posts, post)
	}

	if len(posts) == 0 {
		return []models.Post{}, err
	}

	return posts, err
}

func (r *repository) GetThreadPosts(threadSlugOrID string, limit, since uint64, sort string, desc bool) (posts []models.Post, err error) {
	threadID, err := r.GetThreadID(threadSlugOrID)
	if err != nil {
		return posts, err
	}

	if sort == "tree" {
		return r.GetTreeSortPosts(threadID, limit, since, desc)
	} else if sort == "parent_tree" {
		return r.GetParentTreeSortPosts(threadID, limit, since, desc)
	}

	return r.GetFlatSortPosts(threadID, limit, since, desc)
}

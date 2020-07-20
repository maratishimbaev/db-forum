package postPostgres

import (
	"database/sql"
	"fmt"
	"forum/models"
	_post "forum/post"
	"forum/utils"
	"strconv"
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

func (r *repository) GetPostFull(id uint64, related []string) (*models.PostFull, error) {
	postFull := models.PostFull{Post: models.Post{ID: id}}

	getPost := "SELECT author, created, forum, is_edited, message, parent, thread FROM post WHERE id = $1"
	if err := r.db.QueryRow(getPost, id).
		Scan(&postFull.Post.Author, &postFull.Post.Created, &postFull.Post.Forum, &postFull.Post.IsEdited,
			&postFull.Post.Message, &postFull.Post.Parent, &postFull.Post.Thread); err != nil {
		return nil, _post.NotFound
	}

	authorContains := contains(related, "user")
	if authorContains {
		postFull.Author = &models.User{}

		getAuthor := "SELECT about, email, fullname, nickname FROM \"user\" WHERE nickname = $1"
		if err := r.db.QueryRow(getAuthor, postFull.Post.Author).
			Scan(&postFull.Author.About, &postFull.Author.Email, &postFull.Author.FullName,
				&postFull.Author.Nickname); err != nil {
			return nil, err
		}
	}

	forumContains := contains(related, "forum")
	if forumContains {
		postFull.Forum = &models.Forum{}

		getForum := "SELECT slug, title, \"user\", posts, threads FROM forum WHERE slug = $1"
		if err := r.db.QueryRow(getForum, postFull.Post.Forum).
			Scan(&postFull.Forum.Slug, &postFull.Forum.Title, &postFull.Forum.User, &postFull.Forum.Posts,
				&postFull.Forum.Threads); err != nil {
			return nil, err
		}
	}

	threadContains := contains(related, "thread")
	if threadContains {
		postFull.Thread = &models.Thread{}

		getThread := "SELECT id, author, created, forum, message, slug, title, votes FROM thread WHERE id = $1"
		if err := r.db.QueryRow(getThread, postFull.Post.Thread).
			Scan(&postFull.Thread.ID, &postFull.Thread.Author, &postFull.Thread.Created, &postFull.Thread.Forum,
				&postFull.Thread.Message, &postFull.Thread.Slug, &postFull.Thread.Title, &postFull.Thread.Votes); err != nil {
			return nil, err
		}
	}

	return &postFull, nil
}

func (r *repository) ChangePost(newPost *models.Post) (*models.Post, error) {
	var hasPost bool

	checkPost := "SELECT EXISTS(SELECT 1 FROM post WHERE id = $1)"
	if err := r.db.QueryRow(checkPost, newPost.ID).Scan(&hasPost); err != nil || !hasPost {
		return nil, _post.NotFound
	}

	getPostMessage := "SELECT message FROM post WHERE id = $1"

	var oldPostMessage string

	err := r.db.QueryRow(getPostMessage, newPost.ID).Scan(&oldPostMessage)
	if err != nil {
		return nil, err
	}

	if oldPostMessage != newPost.Message && newPost.Message != "" {
		changePost := "UPDATE post SET message = $1, is_edited = true WHERE id = $2"
		if _, err = r.db.Exec(changePost, newPost.Message, newPost.ID); err != nil {
			return nil, _post.NotFound
		}
	}

	var post models.Post

	getPost := "SELECT id, author, created, forum, is_edited, message, parent, thread FROM post WHERE id = $1"
	err = r.db.QueryRow(getPost, newPost.ID).
		Scan(&post.ID, &post.Author, &post.Created, &post.Forum, &post.IsEdited, &post.Message, &post.Parent, &post.Thread)

	return &post, nil
}

func (r *repository) GetThreadID(threadSlugOrID string) (threadID uint64, err error) {
	getThreadIDByID := "SELECT id FROM thread WHERE id = $1"
	err = r.db.QueryRow(getThreadIDByID, threadSlugOrID).Scan(&threadID)
	if err != nil {
		getThreadIDBySlug := "SELECT id from thread WHERE LOWER(slug) = LOWER($1)"
		err = r.db.QueryRow(getThreadIDBySlug, threadSlugOrID).Scan(&threadID)
		if err != nil {
			return 0, _post.ThreadNotFound
		}
	}

	return threadID, nil
}

func (r *repository) CreatePosts(threadSlugOrID string, newPosts []models.Post) (posts []models.Post, err error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}

	var forumSlug string

	threadId, err := strconv.ParseUint(threadSlugOrID, 10, 64)
	if err != nil {
		threadId = 0
	}

	getThread := "SELECT id, forum FROM thread WHERE id = $1 OR (slug <> '' AND slug = $2)"
	err = tx.QueryRow(getThread, threadId, threadSlugOrID).Scan(&threadId, &forumSlug)
	if err != nil {
		_ = tx.Rollback()
		return nil, _post.ThreadNotFound
	}

	if len(newPosts) == 0 {
		_ = tx.Commit()
		return []models.Post{}, nil
	}

	now := time.Now()

	createPost := "INSERT INTO post (id, author, created, forum, is_edited, message, parent, thread, path) VALUES"
	var vals []interface{}

	for _, newPost := range newPosts {
		newPost.Thread = threadId

		if newPost.Parent != 0 {
			var parentThread uint64

			err = tx.QueryRow(`SELECT thread FROM post WHERE id = $1`, newPost.Parent).Scan(&parentThread)
			if err != nil || parentThread != newPost.Thread {
				_ = tx.Rollback()
				return nil, _post.ParentNotInThread
			}
			createPost += "(nextval('post_id_seq'::regclass), ?, ?, ?, ?, ?, ?, ?," +
				"(SELECT path FROM post WHERE id = ?) || currval(pg_get_serial_sequence('post', 'id'))::integer),"
			vals = append(vals, newPost.Author, now, forumSlug, false, newPost.Message, newPost.Parent, newPost.Thread, newPost.Parent)
		} else {
			createPost += "(nextval('post_id_seq'::regclass), ?, ?, ?, ?, ?, ?, ?," +
				"ARRAY[currval(pg_get_serial_sequence('post', 'id'))::integer]),"
			vals = append(vals, newPost.Author, now, forumSlug, false, newPost.Message, newPost.Parent, newPost.Thread)
		}
	}

	createPost = createPost[0 : len(createPost)-1]
	createPost += " RETURNING id, author, created, forum, is_edited, message, parent, thread"
	createPost = utils.ReplaceSQL(createPost, "?", 1)

	statement, err := tx.Prepare(createPost)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	rows, err := statement.Query(vals...)
	if err != nil {
		_ = tx.Rollback()
		return nil, _post.NotFound
	}

	for rows.Next() {
		var post models.Post

		err = rows.Scan(&post.ID, &post.Author, &post.Created, &post.Forum, &post.IsEdited, &post.Message, &post.Parent, &post.Thread)
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}

		posts = append(posts, post)
	}

	_, err = tx.Exec("UPDATE forum SET posts = posts + $1 WHERE slug = $2", len(posts), forumSlug)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	_ = tx.Commit()
	return posts, nil
}

func (r *repository) GetFlatSortPosts(threadID, limit, since uint64, desc bool) (posts []models.Post, err error) {
	getPosts := "SELECT id, author, created, forum, is_edited, message, parent FROM post WHERE thread = $1"

	if since != 0 {
		var strSign string
		if strSign = ">"; desc {
			strSign = "<"
		}
		getPosts += fmt.Sprintf(" AND id %s %d", strSign, since)
	}

	var strDesc string
	if desc {
		strDesc = "DESC"
	}
	getPosts += fmt.Sprintf(" ORDER BY id %s", strDesc)

	if limit != 0 {
		getPosts += fmt.Sprintf(" LIMIT %d", limit)
	}

	parentRows, err := r.db.Query(getPosts, threadID)
	if err != nil {
		return nil, err
	}
	defer parentRows.Close()

	for parentRows.Next() {
		post := models.Post{Thread: threadID}

		err = parentRows.Scan(&post.ID, &post.Author, &post.Created, &post.Forum, &post.IsEdited, &post.Message, &post.Parent)
		if err != nil {
			return nil, err
		}

		posts = append(posts, post)
	}

	if len(posts) == 0 {
		return []models.Post{}, nil
	}

	return posts, nil
}

func (r *repository) GetTreeSortPosts(threadID, limit, since uint64, desc bool) (posts []models.Post, err error) {
	getPosts := "SELECT id, author, created, forum, is_edited, message, parent FROM post WHERE thread = $1"

	if since != 0 {
		var strSign string
		if strSign = ">"; desc {
			strSign = "<"
		}
		getPosts += fmt.Sprintf(" AND path %s (SELECT path FROM post WHERE id = %d)", strSign, since)
	}

	var strDesc string
	if desc {
		strDesc = "DESC"
	}
	getPosts += fmt.Sprintf(" ORDER BY path %s", strDesc)

	if limit != 0 {
		getPosts += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := r.db.Query(getPosts, threadID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		post := models.Post{Thread: threadID}

		err = rows.Scan(&post.ID, &post.Author, &post.Created, &post.Forum, &post.IsEdited, &post.Message, &post.Parent)
		if err != nil {
			return nil, err
		}

		posts = append(posts, post)
	}

	if len(posts) == 0 {
		return []models.Post{}, nil
	}

	return posts, nil
}

func (r *repository) GetParentTreeSortPosts(threadID, limit, since uint64, desc bool) (posts []models.Post, err error) {
	var strSince string
	if since != 0 {
		var strSign string
		if strSign = ">"; desc {
			strSign = "<"
		}
		strSince = fmt.Sprintf("AND path %s (SELECT path[1:1] FROM post WHERE id = %d)", strSign, since)
	}

	var strLimit string
	if limit != 0 {
		strLimit = fmt.Sprintf("LIMIT %d", limit)
	}

	var strDesc string
	if desc {
		strDesc = "DESC"
	}

	getPosts := fmt.Sprintf(
		"SELECT id, author, created, forum, is_edited, message, parent " +
			"FROM post WHERE thread = $1 AND path && " +
			"(SELECT ARRAY(SELECT id FROM post WHERE thread = $1 AND parent = 0 %s ORDER BY id %s %s)) " +
			"ORDER BY path[1] %s, path",
			strSince, strDesc, strLimit, strDesc,
	)

	rows, err := r.db.Query(getPosts, threadID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		post := models.Post{Thread: threadID}

		err = rows.Scan(&post.ID, &post.Author, &post.Created, &post.Forum, &post.IsEdited, &post.Message, &post.Parent)
		if err != nil {
			return nil, err
		}

		posts = append(posts, post)
	}

	if len(posts) == 0 {
		return []models.Post{}, nil
	}

	return posts, nil
}

func (r *repository) GetThreadPosts(threadSlugOrID string, limit, since uint64, sort string, desc bool) (posts []models.Post, err error) {
	threadID, err := r.GetThreadID(threadSlugOrID)
	if err != nil {
		return nil, err
	}

	if sort == "tree" {
		return r.GetTreeSortPosts(threadID, limit, since, desc)
	} else if sort == "parent_tree" {
		return r.GetParentTreeSortPosts(threadID, limit, since, desc)
	}

	return r.GetFlatSortPosts(threadID, limit, since, desc)
}

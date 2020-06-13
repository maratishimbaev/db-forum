package postPostgres

import (
	"database/sql"
	"errors"
	"fmt"
	"forum/models"
	_post "forum/post"
	"forum/utils"
	"github.com/lib/pq"
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

func (r *repository) GetPostFull(id uint64, related []string) (postFull models.PostFull, err error) {
	postFull.Post.ID = id

	getPost := `
		SELECT u.nickname, p.created, f.slug, p.is_edited, p.message, p.parent, p.thread
		FROM post p
		JOIN "user" u ON p.author = u.id
		JOIN forum f ON p.forum = f.id
		WHERE p.id = $1`
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
			FROM "user" WHERE LOWER(nickname) = LOWER($1)`
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
			SELECT f.slug, f.title, u.nickname, f.posts, f.threads
			FROM forum f
			JOIN "user" u ON f.user = u.id
			WHERE LOWER(f.slug) = LOWER($1)`
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
			SELECT t.id, u.nickname, t.created, f.slug, t.message, t.slug, t.title, t.votes
			FROM thread t
			JOIN "user" u ON t.author = u.id
			JOIN forum f ON t.forum = f.id
			WHERE t.id = $1`
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
		SELECT p.id, u.nickname, p.created, f.slug, p.is_edited, p.message, p.parent, p.thread
		FROM post p
		JOIN "user" u ON p.author = u.id
		JOIN forum f ON p.forum = f.id
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
	var threadID uint64
	var isID bool

	threadID, err = strconv.ParseUint(threadSlugOrID, 10, 64)
	if err != nil {
		threadID = 0
	}

	threadExists := `SELECT EXISTS(SELECT 1 FROM thread WHERE id = $1)`
	if err = r.db.QueryRow(threadExists, threadID).Scan(&isID); err != nil {
		return posts, err
	}
	if !isID {
		getThreadID := `SELECT id FROM thread WHERE LOWER(slug) = LOWER($1)`
		if err = r.db.QueryRow(getThreadID, threadSlugOrID).Scan(&threadID); err != nil {
			return posts, _post.ThreadNotFound
		}
	}

	if len(newPosts) == 0 {
		return []models.Post{}, err
	}

	now := time.Now()

	createPost := `INSERT INTO post (author, created, forum, is_edited, message, parent, thread) VALUES`
	var vals []interface{}

	for _, newPost := range newPosts {
		newPost.Thread = threadID

		if newPost.Parent != 0 {
			var parentThread uint64

			err = r.db.QueryRow(`SELECT thread FROM post WHERE id = $1`, newPost.Parent).Scan(&parentThread)
			if err != nil || parentThread != newPost.Thread {
				return posts, _post.ParentNotInThread
			}
		}

		var authorID uint64
		err = r.db.QueryRow(`SELECT id FROM "user" WHERE LOWER(nickname) = LOWER($1)`, newPost.Author).Scan(&authorID)
		if err != nil {
			return posts, _post.NotFound
		}

		var forumID uint64
		err = r.db.QueryRow(`SELECT forum FROM thread WHERE id = $1`, threadID).Scan(&forumID)
		if err != nil {
			return posts, errors.New("can't find forum, error: " + err.Error())
		}

		createPost += " (?, ?, ?, ?, ?, ?, ?),"
		vals = append(vals, authorID, now, forumID, false, newPost.Message, newPost.Parent, newPost.Thread)
	}

	createPost = createPost[0 : len(createPost)-1]
	createPost += ` RETURNING id`
	createPost = utils.ReplaceSQL(createPost, "?", 1)

	statement, err := r.db.Prepare(createPost)
	if err != nil {
		return posts, errors.New("statement error: " + err.Error())
	}
	idRows, err := statement.Query(vals...)
	if err != nil {
		return posts, errors.New("idRows error: " + err.Error())
	}

	var ids []uint64
	for idRows.Next() {
		var id uint64
		err = idRows.Scan(&id)
		if err != nil {
			return posts, err
		}
		ids = append(ids, id)
	}

	getPosts := `
		SELECT p.id, u.nickname, p.created, f.slug, p.is_edited, p.message, p.parent, thread
		FROM post p
		JOIN "user" u ON p.author = u.id
		JOIN forum f ON p.forum = f.id
		WHERE p.id = ANY($1)
		ORDER BY id`
	rows, err := r.db.Query(getPosts, pq.Array(ids))
	if err != nil {
		fmt.Printf("getPosts: %s\n", getPosts)
		return posts, errors.New("rows error: " + err.Error())
	}

	for rows.Next() {
		var post models.Post

		err = rows.Scan(&post.ID, &post.Author, &post.Created, &post.Forum, &post.IsEdited, &post.Message, &post.Parent, &post.Thread)
		if err != nil {
			return posts, err
		}

		posts = append(posts, post)
	}

	return posts, err
}

func (r *repository) GetFlatSortPosts(threadID, limit, since uint64, desc bool) (posts []models.Post, err error) {
	getPosts := `
		SELECT p.id, u.nickname, p.created, f.slug, p.is_edited, p.message, p.parent 
		FROM post p 
		JOIN "user" u ON p.author = u.id
		JOIN forum f ON p.forum = f.id
		WHERE p.thread = $1`

	if !desc {
		if since != 0 {
			getPosts += ` AND p.id > ?`
		}
		getPosts += ` ORDER BY created, id`
	} else {
		if since != 0 {
			getPosts += ` AND p.id < ?`
		}
		getPosts += ` ORDER BY created DESC, id DESC`
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
	var startsWith uint64

	if since != 0 && !desc {
		getParents := `
		WITH RECURSIVE parent (id, parent) AS (
			SELECT id, parent
			FROM post
			WHERE id = $1
		UNION ALL
			SELECT post.id, post.parent
			FROM post
			JOIN parent ON post.id = parent.parent
		)
		SELECT id FROM parent
		WHERE parent = 0`
		err = r.db.QueryRow(getParents, since).Scan(&startsWith)
	}

	getParentPosts := `
		SELECT id
		FROM post
		WHERE thread = $1 AND parent = 0 AND id >= $2
		ORDER BY created, id`

	parentRows, err := r.db.Query(getParentPosts, threadID, startsWith)
	if err != nil {
		return posts, err
	}
	defer parentRows.Close()

	getChildPosts := `
		WITH RECURSIVE children (id, author, created, forum, is_edited, message, parent, path) AS (
			SELECT p.id, u.nickname, p.created, f.slug, p.is_edited, p.message, p.parent, lpad(p.id::text, 7, '0')
			FROM post p
			JOIN "user" u ON p.author = u.id
			JOIN forum f ON p.forum = f.id
			WHERE p.id = $1
		UNION ALL
			SELECT p.id, u.nickname, p.created, f.slug, p.is_edited, p.message, p.parent, c.path || '.' || lpad(p.id::text, 7, '0')
			FROM post p
			JOIN children c ON p.parent = c.id
			JOIN "user" u ON p.author = u.id
			JOIN forum f ON p.forum = f.id
		)
		SELECT id, author, created, forum, is_edited, message, parent FROM children
		ORDER BY path`

	var isSince bool
	var count uint64

MainFor:
	for parentRows.Next() && (limit == 0 || count < limit) {
		var parentID uint64

		err = parentRows.Scan(&parentID)
		if err != nil {
			return posts, err
		}

		childRows, err := r.db.Query(getChildPosts, parentID)
		if err != nil {
			return posts, err
		}

		for childRows.Next() && (limit == 0 || count < limit) {
			post := models.Post{Thread: threadID}

			err = childRows.Scan(&post.ID, &post.Author, &post.Created, &post.Forum, &post.IsEdited, &post.Message, &post.Parent)
			if err != nil {
				return posts, err
			}

			if !desc {
				if since == 0 || isSince {
					posts = append(posts, post)
					count++
				}
				if post.ID == since {
					isSince = true
				}
			} else {
				if post.ID == since {
					break MainFor
				}

				posts = append(posts, post)
			}
		}
	}

	if len(posts) == 0 {
		return []models.Post{}, err
	}

	if desc {
		if len(posts) > int(limit) {
			posts = posts[len(posts)-int(limit):]
		}

		for left, right := 0, len(posts)-1; left < right; left, right = left+1, right-1 {
			posts[left], posts[right] = posts[right], posts[left]
		}
	}

	return posts, err
}

func (r *repository) GetParentTreeSortPosts(threadID, limit, since uint64, desc bool) (posts []models.Post, err error) {
	var startsWith uint64

	if since != 0 {
		getParents := `
		WITH RECURSIVE parent (id, parent) AS (
			SELECT id, parent
			FROM post
			WHERE id = $1
		UNION ALL
			SELECT post.id, post.parent
			FROM post
			JOIN parent ON post.id = parent.parent
		)
		SELECT id FROM parent
		WHERE parent = 0`
		err = r.db.QueryRow(getParents, since).Scan(&startsWith)
	}

	getParentPosts := `
		SELECT id
		FROM post
		WHERE thread = $1 AND parent = 0`

	if desc {
		if since != 0 {
			getParentPosts += ` AND id < $3`
		}
		getParentPosts += ` ORDER BY created DESC, id DESC`
	} else {
		if since != 0 {
			getParentPosts += ` AND id > $3`
		}
		getParentPosts += ` ORDER BY created, id`
	}

	getParentPosts += ` LIMIT $2`

	var parentRows *sql.Rows
	if since != 0 {
		parentRows, err = r.db.Query(getParentPosts, threadID, limit, startsWith)
	} else {
		parentRows, err = r.db.Query(getParentPosts, threadID, limit)
	}
	if err != nil {
		return posts, err
	}
	defer parentRows.Close()

	getChildPosts := `
		WITH RECURSIVE children (id, author, created, forum, is_edited, message, parent, path) AS (
			SELECT p.id, u.nickname, p.created, f.slug, p.is_edited, p.message, p.parent, lpad(p.id::text, 7, '0')
			FROM post p
			JOIN "user" u ON p.author = u.id
			JOIN forum f ON p.forum = f.id
			WHERE p.id = $1
		UNION ALL
			SELECT p.id, u.nickname, p.created, f.slug, p.is_edited, p.message, p.parent, c.path || '.' || lpad(p.id::text, 7, '0')
			FROM post p
			JOIN children c ON p.parent = c.id
			JOIN "user" u ON p.author = u.id
			JOIN forum f ON p.forum = f.id
		)
		SELECT id, author, created, forum, is_edited, message, parent FROM children
		ORDER BY path`

	for parentRows.Next() {
		var parentID uint64

		err = parentRows.Scan(&parentID)
		if err != nil {
			return posts, err
		}

		childRows, err := r.db.Query(getChildPosts, parentID)
		if err != nil {
			return posts, err
		}

		for childRows.Next() {
			post := models.Post{Thread: threadID}

			err = childRows.Scan(&post.ID, &post.Author, &post.Created, &post.Forum, &post.IsEdited, &post.Message, &post.Parent)
			if err != nil {
				return posts, err
			}

			posts = append(posts, post)
		}
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

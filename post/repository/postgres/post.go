package postPostgres

import (
	"database/sql"
	"errors"
	"forum/models"
	_post "forum/post"
	"log"
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
			Scan(&postFull.Author.About, &postFull.Author.Email, &postFull.Author.FullName, &postFull.Author.Nickname); err != nil {
			return postFull, err
		}
	}

	forumContains := contains(related, "forum")
	if forumContains {
		postFull.Forum = &models.Forum{}

		getForum := `
			SELECT f.slug, f.title, u.nickname
			FROM forum f
			JOIN "user" u ON f.user = u.id
			WHERE LOWER(f.slug) = LOWER($1)`
		if err = r.db.QueryRow(getForum, postFull.Post.Forum).
			Scan(&postFull.Forum.Slug, &postFull.Forum.Title, &postFull.Forum.User); err != nil {
			return postFull, err
		}

		getPostCount := `
			SELECT COUNT(*)
			FROM post p
			JOIN forum f ON p.forum = f.id
			WHERE LOWER(f.slug) = LOWER($1)`
		err = r.db.QueryRow(getPostCount, postFull.Forum.Slug).Scan(&postFull.Forum.Posts)
		if err != nil {
			return postFull, errors.New("can't get post count, error: " + err.Error())
		}

		getThreadCount := `
			SELECT COUNT(*)
			FROM thread t
			JOIN forum f ON t.forum = f.id
			WHERE LOWER(f.slug) = LOWER($1)`
		err = r.db.QueryRow(getThreadCount, postFull.Forum.Slug).Scan(&postFull.Forum.Threads)
		if err != nil {
			return postFull, errors.New("can't get thread count, error: " + err.Error())
		}
	}

	threadContains := contains(related, "thread")
	if threadContains {
		postFull.Thread = &models.Thread{}

		getThread := `
			SELECT t.id, u.nickname, t.created, f.slug, t.message, t.slug, t.title
			FROM thread t
			JOIN "user" u ON t.author = u.id
			JOIN forum f ON t.forum = f.id
			WHERE t.id = $1`
		if err = r.db.QueryRow(getThread, postFull.Post.Thread).
			Scan(&postFull.Thread.ID, &postFull.Thread.Author, &postFull.Thread.Created, &postFull.Thread.Forum,
				&postFull.Thread.Message, &postFull.Thread.Slug, &postFull.Thread.Title); err != nil {
			return postFull, err
		}
	}

	return postFull, err
}

func (r *repository) ChangePost(newPost *models.Post) (post models.Post, err error) {
	var hasPost bool

	checkPost := `SELECT COUNT(*) <> 0 FROM post WHERE id = $1`
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
		LEFT JOIN "user" u ON p.author = u.id
		LEFT JOIN forum f ON p.forum = f.id
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
	threadID, err := r.GetThreadID(threadSlugOrID)
	if err != nil {
		return posts, err
	}

	checkThread := `SELECT COUNT(*) <> 0 FROM thread WHERE id = $1`

	var hasThread bool

	err = r.db.QueryRow(checkThread, threadID).Scan(&hasThread)
	if err != nil || !hasThread {
		return posts, _post.ThreadNotFound
	}

	if len(newPosts) == 0 {
		return []models.Post{}, err
	}

	getThread := `SELECT thread FROM post WHERE id = $1`
	createPost := `
		INSERT INTO post (author, created, forum, is_edited, message, parent, thread)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
	getPost := `
		SELECT p.id, u.nickname, p.created, f.slug, p.is_edited, p.message, p.parent, p.thread
		FROM post p
		LEFT JOIN "user" u ON p.author = u.id
		LEFT JOIN forum f ON p.forum = f.id
		WHERE p.id = $1`

	now := time.Now()

	checkAuthor := `SELECT COUNT(*) <> 0 FROM "user" WHERE LOWER(nickname) = LOWER($1)`

	for _, newPost := range newPosts {
		newPost.Thread = threadID

		if newPost.Parent != 0 {
			var parentThread uint64

			err = r.db.QueryRow(getThread, newPost.Parent).Scan(&parentThread)
			if err != nil || parentThread != newPost.Thread {
				return posts, _post.ParentNotInThread
			}
		}

		var postID uint64
		var hasAuthor bool

		err = r.db.QueryRow(checkAuthor, newPost.Author).Scan(&hasAuthor)
		if err != nil || !hasAuthor {
			return posts, _post.NotFound
		}

		var authorID uint64
		getAuthorID := `SELECT id FROM "user" WHERE LOWER(nickname) = LOWER($1)`
		err = r.db.QueryRow(getAuthorID, newPost.Author).Scan(&authorID)
		if err != nil {
			return posts, errors.New("can't find author, error: " + err.Error())
		}

		var forumID uint64
		getForumID := `SELECT forum FROM thread WHERE id = $1`
		err = r.db.QueryRow(getForumID, threadID).Scan(&forumID)
		if err != nil {
			return posts, errors.New("can't find forum, error: " + err.Error())
		}

		if err = r.db.QueryRow(createPost, authorID, now, forumID, false, newPost.Message, newPost.Parent, newPost.Thread).
			Scan(&postID); err != nil {
			return posts, errors.New("can't create post, error: " + err.Error())
		}

		var post models.Post

		err = r.db.QueryRow(getPost, postID).
			Scan(&post.ID, &post.Author, &post.Created, &post.Forum, &post.IsEdited, &post.Message, &post.Parent, &post.Thread)

		posts = append(posts, post)
	}

	return posts, err
}

func (r *repository) GetFlatSortPosts(threadID, limit, since uint64, desc bool) (posts []models.Post, err error) {
	getPosts := `
		SELECT p.id, u.nickname, p.created, f.slug, p.is_edited, p.message, p.parent 
		FROM post p 
		LEFT JOIN "user" u ON p.author = u.id
		LEFT JOIN forum f ON p.forum = f.id
		WHERE p.thread = $1`

	if desc {
		getPosts = getPosts + " ORDER BY created DESC, id DESC"
	} else {
		getPosts = getPosts + " ORDER BY created, id"
	}

	var parentRows *sql.Rows

	parentRows, err = r.db.Query(getPosts, threadID)
	if err != nil {
		return posts, err
	}
	defer parentRows.Close()

	var postCount uint64
	var isSince bool

	for parentRows.Next() && (limit == 0 || postCount < limit) {
		post := models.Post{Thread: threadID}

		err = parentRows.Scan(&post.ID, &post.Author, &post.Created, &post.Forum, &post.IsEdited, &post.Message, &post.Parent)
		if err != nil {
			return posts, err
		}

		if since == 0 || isSince {
			posts = append(posts, post)
			postCount++
		}
		if post.ID == since {
			isSince = true
		}
	}

	if len(posts) == 0 {
		return []models.Post{}, err
	}

	return posts, err
}

func (r *repository) GetTreeSortPosts(threadID, limit, since uint64, desc bool) (posts []models.Post, err error) {
	getParentPosts := "SELECT id FROM post WHERE thread = $1 AND parent = 0 ORDER BY created, id"

	parentRows, err := r.db.Query(getParentPosts, threadID)
	if err != nil {
		return posts, err
	}
	defer parentRows.Close()

	getChildPosts := `
		WITH RECURSIVE children (id, author, created, forum, is_edited, message, parent, path) AS (
			SELECT p.id, u.nickname, p.created, f.slug, p.is_edited, p.message, p.parent, lpad(p.id::text, 5, '0')
			FROM post p
			LEFT JOIN "user" u ON p.author = u.id
			LEFT JOIN forum f ON p.forum = f.id
			WHERE p.id = $1
		UNION ALL
			SELECT p.id, u.nickname, p.created, f.slug, p.is_edited, p.message, p.parent, c.path || '.' || lpad(p.id::text, 5, '0')
			FROM post p
			INNER JOIN children c ON p.parent = c.id
			LEFT JOIN "user" u ON p.author = u.id
			LEFT JOIN forum f ON p.forum = f.id
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

	if desc {
		for left, right := 0, len(posts)-1; left < right; left, right = left+1, right-1 {
			posts[left], posts[right] = posts[right], posts[left]
		}
	}

	var sincePosts []models.Post
	var isSince bool

	for _, post := range posts {
		if since == 0 || isSince {
			sincePosts = append(sincePosts, post)
		}
		if post.ID == since {
			isSince = true
		}

		log.Printf("id: %d, isSince: %t, since: %d", post.ID, isSince, since)
	}

	if len(sincePosts) == 0 {
		return []models.Post{}, err
	}

	if limit != 0 && len(sincePosts) > int(limit) {
		return sincePosts[:limit], err
	}

	return sincePosts, err
}

func (r *repository) GetParentTreeSortPosts(threadID, limit, since uint64, desc bool) (posts []models.Post, err error) {
	getParentPosts := `SELECT id FROM post WHERE thread = $1 AND parent = 0`

	if desc {
		getParentPosts = getParentPosts + " ORDER BY created DESC, id DESC"
	} else {
		getParentPosts = getParentPosts + " ORDER BY created, id"
	}

	parentRows, err := r.db.Query(getParentPosts, threadID)
	if err != nil {
		return posts, err
	}
	defer parentRows.Close()

	getChildPosts := `
		WITH RECURSIVE children (id, author, created, forum, is_edited, message, parent, path) AS (
			SELECT p.id, u.nickname, p.created, f.slug, p.is_edited, p.message, p.parent, lpad(p.id::text, 5, '0')
			FROM post p
			LEFT JOIN "user" u ON p.author = u.id
			LEFT JOIN forum f ON p.forum = f.id
			WHERE p.id = $1
		UNION ALL
			SELECT p.id, u.nickname, p.created, f.slug, p.is_edited, p.message, p.parent, c.path || '.' || lpad(p.id::text, 5, '0')
			FROM post p
			INNER JOIN children c ON p.parent = c.id
			LEFT JOIN "user" u ON p.author = u.id
			LEFT JOIN forum f ON p.forum = f.id
		)
		SELECT id, author, created, forum, is_edited, message, parent FROM children
		ORDER BY path`

	var parentCount uint64
	var isSince bool
	var added bool

	for parentRows.Next() && (limit == 0 || parentCount < limit) {
		var parentID uint64

		err = parentRows.Scan(&parentID)
		if err != nil {
			return posts, err
		}

		childRows, err := r.db.Query(getChildPosts, parentID)
		if err != nil {
			return posts, err
		}

		added = false

		for childRows.Next() {
			post := models.Post{Thread: threadID}

			err = childRows.Scan(&post.ID, &post.Author, &post.Created, &post.Forum, &post.IsEdited, &post.Message, &post.Parent)
			if err != nil {
				return posts, err
			}

			if since == 0 || isSince {
				posts = append(posts, post)
				added = true
			}
			if post.ID == since {
				isSince = true
			}

			log.Printf("id: %d, parent: %d, isSince: %t, since: %d, author id: %d, forum id: %d", post.ID, parentID, isSince, since, post.Author, post.Forum)
		}

		if added {
			parentCount++
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

package postPostgres

import (
	"database/sql"
	forumPostgres "forum/forum/repository/postgres"
	"forum/models"
	_post "forum/post"
	threadPostgres "forum/thread/repository/postgres"
	userPostgres "forum/user/repository/postgres"
	"strconv"
	"time"
)

type Repository struct {
	DB *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}

type Post struct {
	ID uint64
	Author uint64
	Created time.Time
	Forum uint64
	IsEdited bool
	Message string
	Parent uint64
	Thread uint64
}

func (r *Repository) toPostgres(post *models.Post) *Post {
	var authorID uint64
	getAuthorID := `SELECT id FROM "user" WHERE nickname = $1`
	if err := r.DB.QueryRow(getAuthorID, post.Author).Scan(&authorID); err != nil {
		authorID = 0
	}

	var forumID uint64
	getForumID := `SELECT forum FROM thread WHERE id = $1`
	if err := r.DB.QueryRow(getForumID, post.Thread).Scan(&forumID); err != nil {
		forumID = 0
	}

	return &Post{
		ID:       post.ID,
		Author:   authorID,
		Created:  post.Created,
		Forum:    forumID,
		IsEdited: post.IsEdited,
		Message:  post.Message,
		Parent:   post.Parent,
		Thread:   post.Thread,
	}
}

func (r *Repository) toModel(post *Post) *models.Post {
	var authorNickname string
	getAuthorNickname := `SELECT nickname FROM "user" WHERE id = $1`
	if err := r.DB.QueryRow(getAuthorNickname, post.Author).Scan(&authorNickname); err != nil {
		authorNickname = ""
	}

	var forumSlug string
	getForumSlug := `SELECT f.slug FROM thread t
					 JOIN forum f ON t.forum = f.id
					 WHERE t.id = $1`
	if err := r.DB.QueryRow(getForumSlug, post.Thread).Scan(&forumSlug); err != nil {
		forumSlug = ""
	}

	return &models.Post{
		Author:   authorNickname,
		Created:  post.Created,
		Forum:    forumSlug,
		ID:       post.ID,
		IsEdited: post.IsEdited,
		Message:  post.Message,
		Parent:   post.Parent,
		Thread:   post.Thread,
	}
}

func (r *Repository) toModelFull(author *userPostgres.User,
								 forum *forumPostgres.Forum,
								 post *Post,
								 thread *threadPostgres.Thread) *models.PostFull {
	getUserID := `SELECT nickname FROM "user" WHERE id = $1`

	var forumUserNickname string
	if err := r.DB.QueryRow(getUserID, forum.User).Scan(&forumUserNickname); err != nil {
		forumUserNickname = ""
	}

	var threadAuthorNickname string
	if err := r.DB.QueryRow(getUserID, thread.Author).Scan(&threadAuthorNickname); err != nil {
		threadAuthorNickname = ""
	}

	postFull := models.PostFull{
		Post: models.Post{
			Author:   author.Nickname,
			Created:  post.Created,
			Forum:    forum.Slug,
			ID:       post.ID,
			IsEdited: post.IsEdited,
			Message:  post.Message,
			Parent:   post.Parent,
			Thread:   thread.ID,
		},
	}

	if *author != (userPostgres.User{}) {
		postFull.Author = &models.User{
			About:    author.About,
			Email:    author.Email,
			FullName: author.FullName,
			Nickname: author.Nickname,
		}
	}

	if *forum != (forumPostgres.Forum{}) {
		postFull.Forum = &models.Forum{
			Slug:  forum.Slug,
			Title: forum.Title,
			User:  forumUserNickname,
		}
	}

	if *thread != (threadPostgres.Thread{}) {
		postFull.Thread = &models.Thread{
			Author:  threadAuthorNickname,
			Created: thread.Created,
			Forum:   forum.Slug,
			ID:      thread.ID,
			Message: thread.Message,
			Slug:    thread.Slug,
			Title:   thread.Title,
		}
	}

	return &postFull
}

func contains(slice []string, searchable string) bool {
	for _, el := range slice {
		if el == searchable {
			return true
		}
	}
	return false
}

func (r *Repository) GetPostFull(id uint64, related []string) (postFull models.PostFull, err error) {
	post := Post{ID: id}

	getPost := `SELECT author, created, forum, is_edited, message, parent, thread
				FROM post WHERE id = $1`
	if err = r.DB.QueryRow(getPost, post.ID).
				  Scan(&post.Author, &post.Created, &post.Forum, &post.IsEdited, &post.Message, &post.Parent, &post.Thread); err != nil {
		return postFull, _post.NewNotFound(id)
	}

	var author userPostgres.User
	authorContains := contains(related, "author")
	if authorContains {
		author.ID = post.Author

		getAuthor := `SELECT about, email, fullname, nickname
				  	  FROM "user" WHERE id = $1`
		if err = r.DB.QueryRow(getAuthor, author.ID).
			Scan(&author.About, &author.Email, &author.FullName, &author.Nickname); err != nil {
			return postFull, err
		}
	}

	var forum forumPostgres.Forum
	forumContains := contains(related, "forum")
	if forumContains {
		forum.ID = post.Forum

		getForum := `SELECT slug, title, "user"
				 	 FROM forum WHERE id = $1`
		if err = r.DB.QueryRow(getForum, forum.ID).Scan(&forum.Slug, &forum.Title, &forum.User); err != nil {
			return postFull, err
		}
	}

	var thread threadPostgres.Thread
	threadContains := contains(related, "thread")
	if threadContains {
		thread.ID = post.Thread

		getThread := `SELECT author, created, forum, message, slug, title
				  	  FROM thread WHERE id = $1`
		if err = r.DB.QueryRow(getThread, thread.ID).
			Scan(&thread.Author, &thread.Created, &thread.Forum, &thread.Message, &thread.Slug, &thread.Title); err != nil {
			return postFull, err
		}
	}

	return *r.toModelFull(&author, &forum, &post, &thread), err
}

func (r *Repository) ChangePost(newPost *models.Post) (post models.Post, err error) {
	changePost := `UPDATE post SET message = $1 WHERE id = $2`
	if _, err = r.DB.Exec(changePost, newPost.Message, newPost.ID); err != nil {
		return post, _post.NewNotFound(newPost.ID)
	}

	getPost := `SELECT id, author, created, forum, is_edited, message, parent, thread
				FROM post WHERE id = $1`
	err = r.DB.QueryRow(getPost, newPost.ID).
			   Scan(&post.ID, &post.Author, &post.Created, &post.Forum, &post.IsEdited, &post.Message, &post.Parent, &post.Thread)

	return post, err
}

func (r *Repository) GetThreadID(threadSlugOrID string) (threadID uint64, err error) {
	var isThreadID bool
	checkThreadID := `SELECT COUNT(*) <> 0 FROM thread WHERE id = $1`
	_ = r.DB.QueryRow(checkThreadID, threadSlugOrID).Scan(&isThreadID)

	if isThreadID {
		threadID, err = strconv.ParseUint(threadSlugOrID, 10, 64)
		if err != nil {
			return threadID, err
		}
	} else {
		getThreadID := `SELECT id FROM thread WHERE slug = $1`
		if err = r.DB.QueryRow(getThreadID, threadSlugOrID).Scan(&threadID); err != nil {
			return threadID, _post.NewThreadNotFound(threadSlugOrID)
		}
	}

	return threadID, err
}

func (r *Repository) CreatePosts(threadSlugOrID string, newPosts []models.Post) (posts []models.Post, err error) {
	threadID, err := r.GetThreadID(threadSlugOrID)
	if err != nil {
		return posts, err
	}

	getThread := `SELECT thread FROM post WHERE id = $1`
	createPost := `
		INSERT INTO post (author, created, forum, is_edited, message, parent, thread)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
	getPost := `
		SELECT id, author, created, forum, is_edited, message, parent, thread
		FROM post WHERE id = $1`

	for _, newPost := range newPosts {
		var postID uint64

		now := time.Now()

		newPost.Thread = threadID
		if err != nil {
			return posts, err
		}

		var parentThread uint64

		err = r.DB.QueryRow(getThread, newPost.Parent).Scan(&parentThread)
		if err != nil || parentThread == newPost.Thread {
			return posts, _post.NewParentNotInThread()
		}

		pgPost := *r.toPostgres(&newPost)
		if err = r.DB.QueryRow(createPost, pgPost.Author, now, pgPost.Forum, false, pgPost.Message, pgPost.Parent, pgPost.Thread).Scan(&postID); err != nil {
			return posts, err
		}

		var post Post

		err = r.DB.QueryRow(getPost, postID).
				   Scan(&post.ID, &post.Author, &post.Created, &post.Forum, &post.IsEdited, &post.Message, &post.Parent, &post.Thread)

		posts = append(posts, *r.toModel(&post))
	}

	return posts, err
}

func (r *Repository) GetFlatSortPosts(threadID, limit, since uint64, desc bool) (posts []models.Post, err error) {
	getPosts := `
		SELECT id, author, created, forum, is_edited, message, parent 
		FROM post WHERE thread = $1 AND id > $2 ORDER BY created`

	if desc {
		getPosts = getPosts + " DESC"
	}

	var parentRows *sql.Rows

	if limit == 0 {
		parentRows, err = r.DB.Query(getPosts, threadID, since)
	} else {
		getPosts = getPosts + " LIMIT $3"

		parentRows, err = r.DB.Query(getPosts, threadID, since, limit)
	}

	if err != nil {
		return posts, err
	}

	for parentRows.Next() {
		post := Post{Thread: threadID}

		err = parentRows.Scan(&post.ID, &post.Author, &post.Created, &post.Forum, &post.IsEdited, &post.Message, &post.Parent)
		if err != nil {
			return posts, err
		}

		posts = append(posts, *r.toModel(&post))
	}

	return posts, err
}

func (r *Repository) GetTreeSortPosts(threadID, limit, since uint64, desc bool) (posts []models.Post, err error) {
	getParentPosts := `SELECT id FROM post WHERE thread = $1 AND parent = 0 AND id > $2 ORDER BY created`

	if desc {
		getParentPosts = getParentPosts + " DESC"
	}

	var parentRows *sql.Rows

	if limit == 0 {
		parentRows, err = r.DB.Query(getParentPosts, threadID, since)
	} else {
		getParentPosts = getParentPosts + " LIMIT $3"

		parentRows, err = r.DB.Query(getParentPosts, threadID, since, limit)
	}

	if err != nil {
		return posts, err
	}

	getChildPosts := `
		WITH RECURSIVE children (id, author, created, forum, is_edited, message, parent) AS (
			SELECT p.id, p.author, p.created, p.forum, p.is_edited, p.message, p.parent
			FROM post p WHERE p.id = $1
		UNION ALL
			SELECT p.id, p.author, p.created, p.forum, p.is_edited, p.message, p.parent
			FROM post p, children c
			WHERE p.parent = c.id
		) SELECT * FROM children`

	var postCount uint64

	for parentRows.Next() && (limit == 0 || postCount < limit) {
		var parentID uint64

		err = parentRows.Scan(&parentID)
		if err != nil {
			return posts, err
		}

		childRows, err := r.DB.Query(getChildPosts, parentID)
		if err != nil {
			return posts, err
		}

		for childRows.Next() && (limit == 0 || postCount < limit) {
			post := Post{Thread: threadID}

			err = childRows.Scan(&post.ID, &post.Author, &post.Created, &post.Forum, &post.IsEdited, &post.Message, &post.Parent)
			if err != nil {
				return posts, err
			}

			posts = append(posts, *r.toModel(&post))
			postCount++
		}
	}

	return posts, err
}

func (r *Repository) GetParentTreeSortPosts(threadID, limit, since uint64, desc bool) (posts []models.Post, err error) {
	getParentPosts := `SELECT id FROM post WHERE thread = $1 AND parent = 0 AND id > $2 ORDER BY created`

	if desc {
		getParentPosts = getParentPosts + " DESC"
	}

	var parentRows *sql.Rows

	if limit == 0 {
		parentRows, err = r.DB.Query(getParentPosts, threadID, since)
	} else {
		getParentPosts = getParentPosts + " LIMIT $3"

		parentRows, err = r.DB.Query(getParentPosts, threadID, since, limit)
	}

	if err != nil {
		return posts, err
	}

	getChildPosts := `
		WITH RECURSIVE children (id, author, created, forum, is_edited, message, parent) AS (
			SELECT p.id, p.author, p.created, p.forum, p.is_edited, p.message, p.parent
			FROM post p WHERE p.id = $1
		UNION ALL
			SELECT p.id, p.author, p.created, p.forum, p.is_edited, p.message, p.parent
			FROM post p, children c
			WHERE p.parent = c.id
		) SELECT * FROM children`

	for parentRows.Next() {
		var parentID uint64

		err = parentRows.Scan(&parentID)
		if err != nil {
			return posts, err
		}

		childRows, err := r.DB.Query(getChildPosts, parentID)
		if err != nil {
			return posts, err
		}

		for childRows.Next() {
			post := Post{Thread: threadID}

			err = childRows.Scan(&post.ID, &post.Author, &post.Created, &post.Forum, &post.IsEdited, &post.Message, &post.Parent)
			if err != nil {
				return posts, err
			}

			posts = append(posts, *r.toModel(&post))
		}
	}

	return posts, err
}

func (r *Repository) GetThreadPosts(threadSlugOrID string, limit, since uint64, sort string, desc bool) (posts []models.Post, err error) {
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

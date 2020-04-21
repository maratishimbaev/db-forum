package postPostgres

import (
	"database/sql"
	"forum/forum"
	forumPostgres "forum/forum/repository/postgres"
	"forum/models"
	_post "forum/post"
	threadPostgres "forum/thread/repository/postgres"
	userPostgres "forum/user/repository/postgres"
	"log"
	"time"
)

type Repository struct {
	db              *sql.DB
	forumRepository forum.Repository
}

func NewRepository(db *sql.DB, forumRepository forum.Repository) *Repository {
	return &Repository{
		db: db,
		forumRepository: forumRepository,
	}
}

type Post struct {
	ID       uint64
	Author   uint64
	Created  time.Time
	Forum    uint64
	IsEdited bool
	Message  string
	Parent   uint64
	Thread   uint64
}

func (r *Repository) toPostgres(post *models.Post) *Post {
	var authorID uint64
	getAuthorID := `SELECT id FROM "user" WHERE LOWER(nickname) = LOWER($1)`
	if err := r.db.QueryRow(getAuthorID, post.Author).Scan(&authorID); err != nil {
		authorID = 0
	}

	var forumID uint64
	getForumID := `SELECT forum FROM thread WHERE id = $1`
	if err := r.db.QueryRow(getForumID, post.Thread).Scan(&forumID); err != nil {
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
	if err := r.db.QueryRow(getAuthorNickname, post.Author).Scan(&authorNickname); err != nil {
		log.Printf("id: %d, nickname: %s, err: %s\n", post.Author, authorNickname, err.Error())
		authorNickname = ""
	}

	var forumSlug string
	getForumSlug := `SELECT f.slug FROM thread t
					 JOIN forum f ON t.forum = f.id
					 WHERE t.id = $1`
	if err := r.db.QueryRow(getForumSlug, post.Thread).Scan(&forumSlug); err != nil {
		log.Printf("thread id: %d, slug: %s, err: %s\n", post.Thread, forumSlug, err.Error())
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

func (r *Repository) toModelFull(author *userPostgres.User, forum *forumPostgres.Forum, post *Post, thread *threadPostgres.Thread) *models.PostFull {
	getUserID := `SELECT nickname FROM "user" WHERE id = $1`

	var forumUserNickname string
	if err := r.db.QueryRow(getUserID, forum.User).Scan(&forumUserNickname); err != nil {
		forumUserNickname = ""
	}

	var threadAuthorNickname string
	if err := r.db.QueryRow(getUserID, thread.Author).Scan(&threadAuthorNickname); err != nil {
		threadAuthorNickname = ""
	}

	var postAuthorNickname string
	if err := r.db.QueryRow(getUserID, post.Author).Scan(&postAuthorNickname); err != nil {
		postAuthorNickname = ""
	}

	getForumSlug := "SELECT slug FROM forum WHERE id = $1"

	var postForumSlug string
	if err := r.db.QueryRow(getForumSlug, post.Forum).Scan(&postForumSlug); err != nil {
		postForumSlug = ""
	}

	var threadForumSlug string
	if err := r.db.QueryRow(getForumSlug, thread.Forum).Scan(&threadForumSlug); err != nil {
		threadForumSlug = ""
	}

	postFull := models.PostFull{
		Post: models.Post{
			Author:   postAuthorNickname,
			Created:  post.Created,
			Forum:    postForumSlug,
			ID:       post.ID,
			IsEdited: post.IsEdited,
			Message:  post.Message,
			Parent:   post.Parent,
			Thread:   post.Thread,
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

	postCount, err := r.forumRepository.GetPostCount(forum.Slug)
	if err != nil {
		postCount = 0
	}

	threadCount, err := r.forumRepository.GetThreadCount(forum.Slug)
	if err != nil {
		threadCount = 0
	}

	if *forum != (forumPostgres.Forum{}) {
		postFull.Forum = &models.Forum{
			Slug:  forum.Slug,
			Title: forum.Title,
			User:  forumUserNickname,
			Posts: postCount,
			Threads: threadCount,
		}
	}

	if *thread != (threadPostgres.Thread{}) {
		postFull.Thread = &models.Thread{
			Author:  threadAuthorNickname,
			Created: thread.Created,
			Forum:   threadForumSlug,
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
	if err = r.db.QueryRow(getPost, post.ID).
		Scan(&post.Author, &post.Created, &post.Forum, &post.IsEdited, &post.Message, &post.Parent, &post.Thread); err != nil {
		return postFull, _post.NotFound
	}

	var author userPostgres.User
	authorContains := contains(related, "user")
	if authorContains {
		author.ID = post.Author

		getAuthor := `SELECT about, email, fullname, nickname
				  	  FROM "user" WHERE id = $1`
		if err = r.db.QueryRow(getAuthor, author.ID).
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
		if err = r.db.QueryRow(getForum, forum.ID).Scan(&forum.Slug, &forum.Title, &forum.User); err != nil {
			return postFull, err
		}
	}

	var thread threadPostgres.Thread
	threadContains := contains(related, "thread")
	if threadContains {
		thread.ID = post.Thread

		getThread := `SELECT author, created, forum, message, slug, title
				  	  FROM thread WHERE id = $1`
		if err = r.db.QueryRow(getThread, thread.ID).
			Scan(&thread.Author, &thread.Created, &thread.Forum, &thread.Message, &thread.Slug, &thread.Title); err != nil {
			return postFull, err
		}
	}

	return *r.toModelFull(&author, &forum, &post, &thread), err
}

func (r *Repository) ChangePost(newPost *models.Post) (post models.Post, err error) {
	var hasPost bool

	checkPost := `SELECT COUNT(*) <> 0 FROM post WHERE id = $1`
	if err = r.db.QueryRow(checkPost, newPost.ID).Scan(&hasPost); err != nil || !hasPost {
		return post, _post.NotFound
	}

	var pgPost Post

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
		SELECT id, author, created, forum, is_edited, message, parent, thread
		FROM post WHERE id = $1`
	err = r.db.QueryRow(getPost, newPost.ID).
		Scan(&pgPost.ID, &pgPost.Author, &pgPost.Created, &pgPost.Forum, &pgPost.IsEdited, &pgPost.Message, &pgPost.Parent, &pgPost.Thread)

	return *r.toModel(&pgPost), err
}

func (r *Repository) GetThreadID(threadSlugOrID string) (threadID uint64, err error) {
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

func (r *Repository) CreatePosts(threadSlugOrID string, newPosts []models.Post) (posts []models.Post, err error) {
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
		SELECT id, author, created, forum, is_edited, message, parent, thread
		FROM post WHERE id = $1`

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

		pgPost := *r.toPostgres(&newPost)
		if err = r.db.QueryRow(createPost, pgPost.Author, now, pgPost.Forum, false, pgPost.Message, pgPost.Parent, pgPost.Thread).Scan(&postID); err != nil {
			return posts, err
		}

		var post Post

		err = r.db.QueryRow(getPost, postID).
			Scan(&post.ID, &post.Author, &post.Created, &post.Forum, &post.IsEdited, &post.Message, &post.Parent, &post.Thread)

		posts = append(posts, *r.toModel(&post))
	}

	return posts, err
}

func (r *Repository) GetFlatSortPosts(threadID, limit, since uint64, desc bool) (posts []models.Post, err error) {
	getPosts := `
		SELECT id, author, created, forum, is_edited, message, parent 
		FROM post WHERE thread = $1`

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
		post := Post{Thread: threadID}

		err = parentRows.Scan(&post.ID, &post.Author, &post.Created, &post.Forum, &post.IsEdited, &post.Message, &post.Parent)
		if err != nil {
			return posts, err
		}

		if since == 0 || isSince {
			posts = append(posts, *r.toModel(&post))
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

func (r *Repository) GetTreeSortPosts(threadID, limit, since uint64, desc bool) (posts []models.Post, err error) {
	getParentPosts := "SELECT id FROM post WHERE thread = $1 AND parent = 0 ORDER BY created, id"

	parentRows, err := r.db.Query(getParentPosts, threadID)
	if err != nil {
		return posts, err
	}
	defer parentRows.Close()

	getChildPosts := `
		WITH RECURSIVE children (id, author, created, forum, is_edited, message, parent, path) AS (
			SELECT p.id, p.author, p.created, p.forum, p.is_edited, p.message, p.parent, lpad(p.id::text, 5, '0')
			FROM post p WHERE p.id = $1
		UNION ALL
			SELECT p.id, p.author, p.created, p.forum, p.is_edited, p.message, p.parent, c.path || '.' || lpad(p.id::text, 5, '0')
			FROM post p, children c
			WHERE p.parent = c.id
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
			post := Post{Thread: threadID}

			err = childRows.Scan(&post.ID, &post.Author, &post.Created, &post.Forum, &post.IsEdited, &post.Message, &post.Parent)
			if err != nil {
				return posts, err
			}

			posts = append(posts, *r.toModel(&post))
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

func (r *Repository) GetParentTreeSortPosts(threadID, limit, since uint64, desc bool) (posts []models.Post, err error) {
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
			SELECT p.id, p.author, p.created, p.forum, p.is_edited, p.message, p.parent, lpad(p.id::text, 5, '0')
			FROM post p WHERE p.id = $1
		UNION ALL
			SELECT p.id, p.author, p.created, p.forum, p.is_edited, p.message, p.parent, c.path || '.' || lpad(p.id::text, 5, '0')
			FROM post p, children c
			WHERE p.parent = c.id
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
			post := Post{Thread: threadID}

			err = childRows.Scan(&post.ID, &post.Author, &post.Created, &post.Forum, &post.IsEdited, &post.Message, &post.Parent)
			if err != nil {
				return posts, err
			}

			if since == 0 || isSince {
				posts = append(posts, *r.toModel(&post))
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

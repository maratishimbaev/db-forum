package forumPostgres

import (
	"database/sql"
	"forum/models"
	"forum/user"
)

type Repository struct {
	DB *sql.DB
	userRepository user.Repository
}

func NewRepository(db *sql.DB, userRepository user.Repository) *Repository {
	return &Repository{
		DB: db,
		userRepository: userRepository,
	}
}

type Forum struct {
	ID uint64
	Slug string
	Title string
	User uint64
}

func (r *Repository) toPostgres(forum *models.Forum) *Forum {
	userID, err := r.userRepository.GetUserIDByNickname(forum.User)
	if err != nil {
		userID = 0
	}

	return &Forum{
		Slug:  forum.Slug,
		Title: forum.Title,
		User:  userID,
	}
}

func (r *Repository) toModel(forum *Forum) *models.Forum {
	userNickname, err := r.userRepository.GetUserNicknameByID(forum.User)
	if err != nil {
		userNickname = ""
	}

	return &models.Forum{
		Slug:    forum.Slug,
		Title:   forum.Title,
		User:    userNickname,
	}
}

func (r *Repository) CreateForum(newForum *models.Forum) (forum models.Forum, err error) {
	pgForum := r.toPostgres(newForum)

	createForum := `INSERT INTO forum (slug, title, "user")
					VALUES ($1, $2, $3)`
	_, err = r.DB.Exec(createForum, pgForum.Slug, pgForum.Title, pgForum.User)
	if err != nil {
		return forum, err
	}

	// TODO: posts and threads fields

	return *r.toModel(pgForum), err
}

func (r *Repository) GetForum(slug string) (forum models.Forum, err error) {
	pgForum := Forum{Slug: slug}

	getForum := `SELECT title, "user"
				 FROM forum WHERE slug = $1`
	err = r.DB.QueryRow(getForum, pgForum.Slug).Scan(&pgForum.Title, &pgForum.User)
	if err != nil {
		return forum, err
	}

	// TODO: posts and threads fields

	return *r.toModel(&pgForum), err
}

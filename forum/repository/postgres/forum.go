package postgres

import "forum/models"

type Repository struct {}

func NewRepository() *Repository {
	return &Repository{}
}

type Forum struct {
	Posts uint64
	Slug string
	Threads uint64
	Title string
	User string
}

func toPostgres(forum models.Forum) *Forum {
	return &Forum{
		Posts:   forum.Posts,
		Slug:    forum.Slug,
		Threads: forum.Threads,
		Title:   forum.Title,
		User:    forum.User,
	}
}

func toModel(forum Forum) *models.Forum {
	return &models.Forum{
		Posts:   forum.Posts,
		Slug:    forum.Slug,
		Threads: forum.Threads,
		Title:   forum.Title,
		User:    forum.User,
	}
}

func (r *Repository) CreateForum(newForum *models.Forum) (forum models.Forum, err error) {
	return forum, err
}

func (r *Repository) GetForum(slug string) (forum models.Forum, err error) {
	return forum, err
}

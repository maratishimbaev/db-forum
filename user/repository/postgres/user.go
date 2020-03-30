package userPostgres

import (
	"database/sql"
	"forum/models"
)

type Repository struct {
	DB *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}

type User struct {
	ID uint64
	About string
	Email string
	FullName string
	Nickname string
}

func toPostgres(user *models.User) *User {
	return &User{
		About:    user.About,
		Email:    user.Email,
		FullName: user.FullName,
		Nickname: user.Nickname,
	}
}

func toModel(user *User) *models.User {
	return &models.User{
		About:    user.About,
		Email:    user.Email,
		FullName: user.FullName,
		Nickname: user.Nickname,
	}
}

func (r *Repository) CreateUser(newUser *models.User) (user models.User, err error) {
	pgUser := toPostgres(newUser)

	createUser := `INSERT INTO "user" (about, email, fullname, nickname)
				   VALUES ($1, $2, $3, $4)`
	_, err = r.DB.Exec(createUser, pgUser.About, pgUser.Email, pgUser.FullName, pgUser.Nickname)
	if err != nil {
		return user, err
	}

	return *newUser, err
}

func (r *Repository) GetUser(nickname string) (user models.User, err error) {
	pgUser := User{Nickname: nickname}

	getUser := `SELECT about, email, fullname
			   FROM "user" WHERE nickname = $1`
	err = r.DB.QueryRow(getUser, pgUser.Nickname).Scan(&pgUser.About, &pgUser.Email, &pgUser.FullName)
	if err != nil {
		return user, err
	}

	return *toModel(&pgUser), err
}

func (r *Repository) ChangeUser(newUser *models.User) (user models.User, err error) {
	pgUser := *toPostgres(newUser)

	changeUser := `UPDATE "user"
				   SET about = $1, email = $2, fullname = $3
				   WHERE nickname = $4`
	_, err = r.DB.Exec(changeUser, pgUser.About, pgUser.Email, pgUser.FullName, pgUser.Nickname)
	if err != nil {
		return user, err
	}

	return *toModel(&pgUser), nil
}

func (r *Repository) GetForumUsers(forumSlug string, limit uint64, since string, desc bool) (users []models.User, err error) {
	getUsers := `SELECT about, email, fullname, nickname FROM post p
				 JOIN "user" u ON p.author = u.id
				 JOIN forum f ON p.forum = f.id
				 WHERE f.slug = $1
				 UNION
				 SELECT about, email, fullname, nickname FROM thread t
				 JOIN "user" u2 ON t.author = u2.id
				 JOIN forum f ON t.forum = f.id
				 WHERE f.slug = $1`

	rows, err := r.DB.Query(getUsers, forumSlug)

	for rows.Next() {
		var user User

		err = rows.Scan(&user.About, &user.Email, &user.FullName, &user.Nickname)
		if err != nil {
			return users, err
		}

		users = append(users, *toModel(&user))
	}

	return users, err
}

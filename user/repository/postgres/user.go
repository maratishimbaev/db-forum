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
	About string
	Email string
	FullName string
	Nickname string
}

func toPostgres(user models.User) *User {
	return &User{
		About:    user.About,
		Email:    user.Email,
		FullName: user.FullName,
		Nickname: user.Nickname,
	}
}

func toModel(user User) *models.User {
	return &models.User{
		About:    user.About,
		Email:    user.Email,
		FullName: user.FullName,
		Nickname: user.Nickname,
	}
}

func (r *Repository) CreateUser(newUser *models.User) (user models.User, err error) {
	createUser := `INSERT INTO "user" (about, email, fullname, nickname)
				   VALUES ($1, $2, $3, $4)`
	_, err = r.DB.Exec(createUser, newUser.About, newUser.Email, newUser.FullName, newUser.Nickname)
	if err != nil {
		return user, err
	}

	return *newUser, err
}

func (r *Repository) GetUser(nickname string) (user models.User, err error) {
	user.Nickname = nickname

	getUser := `SELECT about, email, fullname
			   FROM "user" WHERE nickname = $1`
	err = r.DB.QueryRow(getUser, user.Nickname).Scan(&user.About, &user.Email, &user.FullName)
	if err != nil {
		return user, err
	}

	return user, err
}

func (r *Repository) ChangeUser(newUser *models.User) (user models.User, err error) {
	changeUser := `UPDATE "user"
				   SET about = $1, email = $2, fullname = $3
				   WHERE nickname = $4`
	_, err = r.DB.Exec(changeUser, newUser.About, newUser.Email, newUser.FullName, newUser.Nickname)
	if err != nil {
		return user, err
	}

	return *newUser, nil
}

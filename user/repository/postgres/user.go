package userPostgres

import (
	"database/sql"
	"fmt"
	"forum/models"
	_user "forum/user"
)

type repository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *repository {
	return &repository{db: db}
}

func (r *repository) CreateUser(newUser *models.User) (users []models.User, err error) {
	createUser := "INSERT INTO \"user\" (about, email, fullname, nickname) VALUES ($1, $2, $3, $4)"
	_, err = r.db.Exec(createUser, newUser.About, newUser.Email, newUser.FullName, newUser.Nickname)
	if err != nil {
		getUsers := "SELECT about, email ,fullname, nickname FROM \"user\" WHERE nickname = $1 OR email = $2"
		rows, err := r.db.Query(getUsers, newUser.Nickname, newUser.Email)
		if err != nil {
			return users, err
		}

		for rows.Next() {
			err = rows.Scan(&newUser.About, &newUser.Email, &newUser.FullName, &newUser.Nickname)
			if err != nil {
				return users, err
			}

			users = append(users, *newUser)
		}

		return users, _user.ErrAlreadyExists
	}

	users = append(users, *newUser)

	return users, nil
}

func (r *repository) GetUser(nickname string) (*models.User, error) {
	var user models.User

	getUser := "SELECT about, email, fullname, nickname FROM \"user\" WHERE nickname = $1"
	err := r.db.QueryRow(getUser, nickname).Scan(&user.About, &user.Email, &user.FullName, &user.Nickname)
	if err != nil {
		return nil, _user.ErrNotFound
	}

	return &user, nil
}

func (r *repository) ChangeUser(newUser *models.User) (*models.User, error) {
	var oldUser models.User

	getOldUser := "SELECT about, email, fullname FROM \"user\" WHERE LOWER(nickname) = LOWER($1)"
	err := r.db.QueryRow(getOldUser, newUser.Nickname).Scan(&oldUser.About, &oldUser.Email, &oldUser.FullName)
	if err != nil {
		return nil, _user.ErrNotFound
	}

	if !(newUser.About == "" && newUser.Email == "" && newUser.FullName == "") {
		if newUser.About == "" {
			newUser.About = oldUser.About
		}
		if newUser.Email == "" {
			newUser.Email = oldUser.Email
		}
		if newUser.FullName == "" {
			newUser.FullName = oldUser.FullName
		}

		changeUser := "UPDATE \"user\" SET about = $1, email = $2, fullname = $3 WHERE LOWER(nickname) = LOWER($4)"
		_, err = r.db.Exec(changeUser, newUser.About, newUser.Email, newUser.FullName, newUser.Nickname)
		if err != nil {
			var userCount uint64

			getUserCount := "SELECT COUNT(*) FROM \"user\" WHERE LOWER(nickname) = LOWER($1)"
			err = r.db.QueryRow(getUserCount, newUser.Nickname).Scan(&userCount)
			if err != nil {
				return nil, err
			}

			if userCount == 0 {
				return nil, _user.ErrNotFound
			}

			return nil, _user.ErrConflictData
		}
	}

	var user models.User

	getUser := "SELECT about, email, fullname, nickname FROM \"user\" WHERE LOWER(nickname) = LOWER($1)"
	err = r.db.QueryRow(getUser, newUser.Nickname).Scan(&user.About, &user.Email, &user.FullName, &user.Nickname)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *repository) GetForumUsers(forumSlug string, limit uint64, since string, desc bool) (users []models.User, err error) {
	var forumId uint64

	checkForum := "SELECT id FROM forum WHERE slug = $1"
	err = r.db.QueryRow(checkForum, forumSlug).Scan(&forumId)
	if err != nil || forumId == 0 {
		return nil, _user.ErrForumNotFound
	}

	getUsers := "SELECT about, email, fullname, nickname FROM \"user\" " +
		"WHERE id IN (SELECT \"user\" FROM forum_user WHERE forum = $1)"

	if since != "" {
		var strSign string
		if strSign = ">"; desc {
			strSign = "<"
		}
		getUsers += fmt.Sprintf(" AND nickname %s '%s'", strSign, since)
	}

	var strDesc string
	if desc {
		strDesc = "DESC"
	}
	getUsers += fmt.Sprintf(" ORDER BY nickname %s", strDesc)

	if limit != 0 {
		getUsers += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := r.db.Query(getUsers, forumId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user models.User

		err = rows.Scan(&user.About, &user.Email, &user.FullName, &user.Nickname)
		if err != nil {
			return nil, _user.ErrForumNotFound
		}

		users = append(users, user)
	}

	if len(users) == 0 {
		return []models.User{}, nil
	}

	return users, nil
}

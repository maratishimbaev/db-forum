package userPostgres

import (
	"database/sql"
	"fmt"
	"forum/models"
	_user "forum/user"
	"forum/utils"
	"strings"
)

type repository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *repository {
	return &repository{db: db}
}

func (r *repository) CreateUser(newUser *models.User) (users []models.User, err error) {
	createUser := `INSERT INTO "user" (about, email, fullname, nickname)
				   VALUES ($1, $2, $3, $4)`
	_, err = r.db.Exec(createUser, newUser.About, newUser.Email, newUser.FullName, newUser.Nickname)
	if err != nil {
		getUsers := `
			SELECT about, email ,fullname, nickname
			FROM "user" WHERE nickname = $1 OR email = $2`
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

	return users, err
}

func (r *repository) GetUser(nickname string) (user models.User, err error) {
	getUser := `
		SELECT about, email, fullname, nickname
		FROM "user" WHERE LOWER(nickname) = LOWER($1)`
	err = r.db.QueryRow(getUser, nickname).Scan(&user.About, &user.Email, &user.FullName, &user.Nickname)
	if err != nil {
		return user, fmt.Errorf("error: %w, nickname: %s", _user.ErrNotFound, nickname)
	}

	return user, err
}

func (r *repository) ChangeUser(newUser *models.User) (user models.User, err error) {
	var oldUser models.User

	getOldUser := `
		SELECT about, email, fullname
		FROM "user" WHERE LOWER(nickname) = LOWER($1)`
	err = r.db.QueryRow(getOldUser, newUser.Nickname).Scan(&oldUser.About, &oldUser.Email, &oldUser.FullName)
	if err != nil {
		return user, fmt.Errorf("error: %w, nickname: %s", _user.ErrNotFound, newUser.Nickname)
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

		changeUser := `UPDATE "user"
				   SET about = $1, email = $2, fullname = $3
				   WHERE LOWER(nickname) = LOWER($4)`
		_, err = r.db.Exec(changeUser, newUser.About, newUser.Email, newUser.FullName, newUser.Nickname)
		if err != nil {
			var userCount uint64

			getUserCount := `SELECT COUNT(*) FROM "user" WHERE LOWER(nickname) = LOWER($1)`
			err = r.db.QueryRow(getUserCount, newUser.Nickname).Scan(&userCount)
			if err != nil {
				return user, err
			}

			if userCount == 0 {
				return user, fmt.Errorf("error: %w, nickname: %s", _user.ErrNotFound, newUser.Nickname)
			}

			return user, _user.ErrConflictData
		}
	}

	getUser := `
		SELECT about, email, fullname, nickname
		FROM "user" WHERE LOWER(nickname) = LOWER($1)`
	err = r.db.QueryRow(getUser, newUser.Nickname).Scan(&user.About, &user.Email, &user.FullName, &user.Nickname)
	if err != nil {
		return user, err
	}

	return user, err
}

func (r *repository) GetForumUsers(forumSlug string, limit uint64, since string, desc bool) (users []models.User, err error) {
	var hasForum bool

	checkForum := "SELECT EXISTS(SELECT 1 FROM forum WHERE LOWER(slug) = LOWER($1))"
	err = r.db.QueryRow(checkForum, forumSlug).Scan(&hasForum)
	if err != nil || !hasForum {
		return users, _user.ErrForumNotFound
	}

	getUsers := `
		SELECT about, email, fullname, nickname
		FROM "user"
		WHERE nickname IN (
			SELECT "user"
			FROM forum_user
			WHERE LOWER(forum) = LOWER($1) since
		)`

	if !desc {
		if since != "" {
			getUsers = strings.Replace(getUsers, "since", `AND LOWER("user") > LOWER(?) COLLATE "C"`, 1)
		} else {
			getUsers = strings.Replace(getUsers, "since", "", 1)
		}
		getUsers += ` ORDER BY LOWER(nickname) COLLATE "C"`
	} else {
		if since != "" {
			getUsers = strings.Replace(getUsers, "since", `AND LOWER("user") < LOWER(?) COLLATE "C"`, 1)
		} else {
			getUsers = strings.Replace(getUsers, "since", "", 1)
		}
		getUsers += ` ORDER BY LOWER(nickname) COLLATE "C" DESC`
	}

	if limit != 0 {
		getUsers += ` LIMIT ?`
	}

	getUsers = utils.ReplaceSQL(getUsers, "?", 2)

	var rows *sql.Rows
	switch true {
	case since != "" && limit != 0:
		rows, err = r.db.Query(getUsers, forumSlug, since, limit)
		break
	case since != "":
		rows, err = r.db.Query(getUsers, forumSlug, since)
		break
	case limit != 0:
		rows, err = r.db.Query(getUsers, forumSlug, limit)
		break
	default:
		rows, err = r.db.Query(getUsers, forumSlug)
	}
	if err != nil {
		return users, err
	}
	defer rows.Close()

	for rows.Next() {
		var user models.User

		err = rows.Scan(&user.About, &user.Email, &user.FullName, &user.Nickname)
		if err != nil {
			return users, fmt.Errorf("error: %w, forum slug: %s", _user.ErrForumNotFound, forumSlug)
		}

		users = append(users, user)
	}

	if len(users) == 0 {
		return []models.User{}, err
	}

	return users, err
}

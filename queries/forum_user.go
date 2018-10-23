package queries

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/ArtAndreev/ForumTP/models"
)

func CreateUser(u *models.ForumUser) ([]models.ForumUser, error) {
	var res []models.ForumUser
	// check not null constraint
	if u.Nickname == "" || u.Email == "" {
		return res, &NullFieldError{"User", "nickname and/or email"}
	}

	r1, err := GetUserByNickname(u.Nickname)
	if err != nil {
		if _, ok := err.(*RecordNotFoundError); !ok {
			return res, err // db error
		}
	} else { // record exists
		res = append(res, r1)
	}

	r2, err := GetUserByEmail(u.Email)
	if err != nil { // record doesn't exist or db error
		if _, ok := err.(*RecordNotFoundError); !ok {
			return res, err // db error
		}
	} else { // record exists
		if r1.Email != r2.Email {
			res = append(res, r2)
		}
	}
	if len(res) != 0 {
		return res, &UniqueFieldValueAlreadyExistsError{"User", "nickname and/or email"}
	}

	_, err = db.NamedExec(`
		INSERT INTO forum_user (nickname, fullname, email, about)
		VALUES (:nickname, :fullname, :email, :about)`,
		u)
	if err != nil {
		return res, err
	}

	res = append(res, *u)
	return res, nil
}

func GetUserByNickname(n string) (models.ForumUser, error) {
	res := models.ForumUser{}
	err := db.Get(&res, "SELECT * FROM forum_user WHERE nickname = $1", n)
	if err != nil {
		if err == sql.ErrNoRows {
			return res, &RecordNotFoundError{"User", n}
		}
		return res, err
	}
	return res, nil
}

func txGetUserByNickname(n string, tx *sqlx.Tx) (models.ForumUser, error) {
	res := models.ForumUser{}
	err := tx.Get(&res, "SELECT * FROM forum_user WHERE nickname = $1", n)
	if err != nil {
		if err == sql.ErrNoRows {
			return res, &RecordNotFoundError{"User", n}
		}
		return res, err
	}
	return res, nil
}

func GetUserByEmail(e string) (models.ForumUser, error) {
	res := models.ForumUser{}
	err := db.Get(&res, "SELECT * FROM forum_user WHERE email = $1", e)
	if err != nil {
		if err == sql.ErrNoRows {
			return res, &RecordNotFoundError{"User", e}
		}
		return res, err
	}
	return res, nil
}

func txGetUserByEmail(e string, tx *sqlx.Tx) (models.ForumUser, error) {
	res := models.ForumUser{}
	err := tx.Get(&res, "SELECT * FROM forum_user WHERE email = $1", e)
	if err != nil {
		if err == sql.ErrNoRows {
			return res, &RecordNotFoundError{"User", e}
		}
		return res, err
	}
	return res, nil
}

func UpdateUser(n string, u *models.ForumUser) (models.ForumUser, error) {
	res := models.ForumUser{}
	tx, err := db.Beginx()
	if err != nil {
		return res, err
	}
	defer tx.Rollback()

	// check existence of user
	_, err = txGetUserByNickname(n, tx)
	if err != nil {
		return res, err
	}

	// update user profile
	if u.Nickname != "" {
		// check conflict
		_, err = txGetUserByNickname(u.Nickname, tx)
		if err != nil {
			if _, ok := err.(*RecordNotFoundError); !ok {
				return res, err // db error
			}
		} else {
			return res, &UniqueFieldValueAlreadyExistsError{"User", "nickname"}
		}

		_, err := tx.Exec(`
		UPDATE forum_user
		SET nickname = $1
		WHERE nickname = $2`,
			u.Nickname, n)
		if err != nil {
			return res, err
		}
	}
	if u.Fullname != "" {
		_, err := tx.Exec(`
		UPDATE forum_user
		SET fullname = $1
		WHERE nickname = $2`,
			u.Fullname, n)
		if err != nil {
			return res, err
		}
	}
	if u.Email != "" {
		// check conflict
		_, err = txGetUserByEmail(u.Email, tx)
		if err != nil {
			if _, ok := err.(*RecordNotFoundError); !ok {
				return res, err // db error
			}
		} else {
			return res, &UniqueFieldValueAlreadyExistsError{"User", "email"}
		}

		_, err = tx.Exec(`
		UPDATE forum_user
		SET email = $1
		WHERE nickname = $2`,
			u.Email, n)
		if err != nil {
			return res, err
		}
	}
	if u.About != "" {
		_, err = tx.Exec(`
		UPDATE forum_user
		SET about = $1
		WHERE nickname = $2`,
			u.About, n)
		if err != nil {
			return res, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return res, err
	}

	// get updated user profile
	res, err = GetUserByNickname(n)
	if err != nil {
		return res, err
	}

	return res, nil
}

func GetAllUsersInForum(s string, params *models.UserQueryParams) ([]models.ForumUser, error) {
	res := []models.ForumUser{}
	// check existence of forum
	_, err := GetForumBySlug(s)
	if err != nil {
		return res, err
	}

	q := `
		WITH forum_threads AS (
			SELECT thread_id, thread_author FROM thread t
			JOIN forum f on t.forum = f.forum_id
			WHERE forum_slug = $1
		)
		SELECT DISTINCT nickname, fullname, email, about FROM forum_threads ft
		JOIN post p ON p.thread = ft.thread_id
		JOIN forum_user u ON u.forum_user_id = post_author
	`
	if params.Since != "" {
		if params.Desc {
			q += "WHERE nickname < $2\n"
		} else {
			q += "WHERE nickname > $2\n"
		}
	}
	q += `
		UNION
		SELECT DISTINCT nickname, fullname, email, about FROM forum_threads ft
		JOIN forum_user u ON u.forum_user_id = ft.thread_author
	`
	if params.Since != "" {
		if params.Desc {
			q += "WHERE nickname < $2\n"
		} else {
			q += "WHERE nickname > $2\n"
		}
	}
	q += "ORDER BY nickname "
	if params.Desc {
		q += "DESC"
	}
	if params.Limit != 0 {
		q += fmt.Sprintf("\nLIMIT %v", params.Limit)
	}
	if params.Since == "" {
		err = db.Select(&res, q, s)
	} else {
		err = db.Select(&res, q, s, params.Since)
	}
	if err != nil {
		return res, err
	}
	return res, nil
}

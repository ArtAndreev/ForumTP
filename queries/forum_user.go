package queries

import (
	"database/sql"

	"github.com/ArtAndreev/ForumTP/models"
)

func CreateUser(u *models.BaseForumUser) ([]models.ForumUser, error) {
	var res []models.ForumUser
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
		return res, &UniqueFieldValueAlreadyExistsError{"User", "email and/or nickname"}
	}

	_, err = db.NamedExec(`
		INSERT INTO forum_user (nickname, fullname, email, about)
		VALUES (:nickname, :fullname, :email, :about)`,
		u)
	res = append(res, models.ForumUser{BaseForumUser: *u})
	if err != nil {
		return res, err
	}
	return res, nil
}

func GetUserByNickname(n string) (models.ForumUser, error) {
	res := models.ForumUser{}
	err := db.Get(&res, "SELECT * FROM forum_user WHERE lower(nickname) = lower($1)", n)
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
	err := db.Get(&res, "SELECT * FROM forum_user WHERE lower(email) = lower($1)", e)
	if err != nil {
		if err == sql.ErrNoRows {
			return res, &RecordNotFoundError{"User", e}
		}
		return res, err
	}
	return res, nil
}

func UpdateUser(n string, u *models.BaseForumUser) (models.ForumUser, error) {
	res := models.ForumUser{}
	tx, err := db.Beginx()
	if err != nil {
		return res, err
	}
	defer tx.Rollback()

	// check existence of user
	_, err = GetUserByNickname(n)
	if err != nil {
		return res, err
	}

	// update user profile
	if u.Nickname != "" {
		// check conflict
		_, err = GetUserByNickname(u.Nickname)
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
		_, err = GetUserByEmail(u.Email)
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

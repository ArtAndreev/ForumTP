package queries

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/lib/pq"

	"github.com/ArtAndreev/ForumTP/models"
)

func CreateUser(u *models.ForumUser) ([]*models.ForumUser, error) {
	var res []*models.ForumUser
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
	inserted := &models.ForumUser{}
	*inserted = *u
	res = append(res, inserted)

	return res, nil
}

func GetUserByNickname(n string) (*models.ForumUser, error) {
	res := &models.ForumUser{}
	err := db.Get(res, "SELECT * FROM forum_user WHERE nickname = $1", n)
	if err != nil {
		if err == sql.ErrNoRows {
			return res, &RecordNotFoundError{"User", n}
		}
		return res, err
	}
	return res, nil
}

func GetUserByEmail(e string) (*models.ForumUser, error) {
	res := &models.ForumUser{}
	err := db.Get(res, "SELECT * FROM forum_user WHERE email = $1", e)
	if err != nil {
		if err == sql.ErrNoRows {
			return res, &RecordNotFoundError{"User", e}
		}
		return res, err
	}
	return res, nil
}

func UpdateUser(n string, u *models.ForumUser) (*models.ForumUser, error) {
	if u.Nickname == "" && u.Fullname == "" && u.Email == "" && u.About == "" {
		return GetUserByNickname(n)
	}

	q := strings.Builder{}
	q.WriteString("UPDATE forum_user SET ")
	args := make([]interface{}, 0, 5)
	continues := false
	fieldCount := 0
	if u.Nickname != "" {
		fieldCount++
		q.WriteString("nickname = $" + strconv.Itoa(fieldCount))
		continues = true
		args = append(args, u.Nickname)
	}
	if u.Fullname != "" {
		fieldCount++
		if continues {
			q.WriteString(", fullname = $" + strconv.Itoa(fieldCount))
		} else {
			q.WriteString("fullname = $" + strconv.Itoa(fieldCount))
			continues = true
		}
		args = append(args, u.Fullname)
	}
	if u.Email != "" {
		fieldCount++
		if continues {
			q.WriteString(", email = $" + strconv.Itoa(fieldCount))
		} else {
			q.WriteString("email = $" + strconv.Itoa(fieldCount))
			continues = true
		}
		args = append(args, u.Email)
	}
	if u.About != "" {
		fieldCount++
		if continues {
			q.WriteString(", about = $" + strconv.Itoa(fieldCount))
		} else {
			q.WriteString("about = $" + strconv.Itoa(fieldCount))
		}
		args = append(args, u.About)
	}
	q.WriteString(" WHERE nickname = $" + strconv.Itoa(fieldCount+1) + " RETURNING *")
	args = append(args, n)
	res := &models.ForumUser{}
	err := db.Get(res, q.String(), args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return res, &RecordNotFoundError{"User", n}
		}
		pqErr := err.(*pq.Error)
		if pqErr.Code == UniqueViolationCode {
			switch {
			case strings.HasPrefix(pqErr.Detail, "Key (nickname)"):
				return res, &UniqueFieldValueAlreadyExistsError{"User", "nickname"}
			case strings.HasPrefix(pqErr.Detail, "Key (email)"):
				return res, &UniqueFieldValueAlreadyExistsError{"User", "email"}
			}
		}
		return res, err
	}

	return res, nil
}

func GetAllUsersInForum(s string, params *models.UserQueryParams) ([]models.ForumUser, error) {
	err := CheckExistenceOfForum(s)
	if err != nil {
		return nil, err
	}

	q := strings.Builder{}
	q.WriteString(`
		WITH forum_threads AS (
			SELECT thread_id, thread_author FROM thread t
			WHERE forum = $1
		)
		SELECT DISTINCT nickname, fullname, email, about FROM forum_threads ft
		JOIN post p ON p.thread = ft.thread_id
		JOIN forum_user u ON u.nickname = post_author`)
	if params.Since != "" {
		if params.Desc {
			q.WriteString(" WHERE nickname < $2")
		} else {
			q.WriteString(" WHERE nickname > $2")
		}
	}
	q.WriteString(`
		UNION
		SELECT DISTINCT nickname, fullname, email, about FROM forum_threads ft
		JOIN forum_user u ON u.nickname = ft.thread_author`)
	if params.Since != "" {
		if params.Desc {
			q.WriteString(" WHERE nickname < $2")
		} else {
			q.WriteString(" WHERE nickname > $2")
		}
	}
	q.WriteString(" ORDER BY nickname")
	if params.Desc {
		q.WriteString(" DESC")
	}
	if params.Limit != 0 {
		q.WriteString(fmt.Sprintf(" LIMIT %v", params.Limit))
	}
	res := []models.ForumUser{}
	if params.Since == "" {
		err = db.Select(&res, q.String(), s)
	} else {
		err = db.Select(&res, q.String(), s, params.Since)
	}
	if err != nil {
		return res, err
	}
	return res, nil
}

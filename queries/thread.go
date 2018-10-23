package queries

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/ArtAndreev/ForumTP/models"
)

func CreateThread(t *models.Thread) (models.Thread, error) {
	res := models.Thread{}
	if t.Forum == "" || t.ThreadTitle == "" || t.ThreadAuthor == "" {
		return res, &NullFieldError{"Thread", "some value(-s) is/are null"}
	}

	// check existence of forum
	f, err := GetForumBySlug(t.Forum)
	if err != nil {
		return res, err
	}

	// check existence of user
	u, err := GetUserByNickname(t.ThreadAuthor)
	if err != nil {
		return res, err
	}

	// check existence of thread
	if t.ThreadSlug != nil {
		res, err = GetThreadBySlug(*t.ThreadSlug)
		if err != nil {
			if _, ok := err.(*RecordNotFoundError); !ok {
				return res, err // db error
			}
		} else { // record exists
			return res, &UniqueFieldValueAlreadyExistsError{"Thread", "title"}
		}
	}

	// insert
	qres, err := db.Query(`
		INSERT INTO thread (forum, thread_slug, thread_title, thread_author, thread_created, thread_message)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING thread_id`,
		f.ForumID, t.ThreadSlug, t.ThreadTitle, u.ForumUserID, t.ThreadCreated, t.ThreadMessage)
	if err != nil {
		return res, err
	}

	lastInsertedID, err := getLastInsertedID(qres)
	if err != nil {
		return res, err
	}

	// get new res
	res, err = GetThreadByID(lastInsertedID)
	if err != nil {
		return res, err
	}
	return res, nil
}

func GetThreadByID(id int) (models.Thread, error) {
	res := models.Thread{}
	err := db.Get(&res, `
		SELECT thread_id, forum_slug forum, thread_slug, thread_title, u.nickname thread_author, thread_created, thread_message, votes FROM thread t
		JOIN forum f ON t.forum = f.forum_id
		JOIN forum_user u ON t.thread_author = u.forum_user_id
		WHERE t.thread_id = $1
		`, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return res, &RecordNotFoundError{"Thread", fmt.Sprintf("%v", id)}
		}
		return res, err
	}
	return res, nil
}

func txGetThreadByID(id int, tx *sqlx.Tx) (models.Thread, error) {
	res := models.Thread{}
	err := tx.Get(&res, `
		SELECT thread_id, forum_slug forum, thread_slug, thread_title, u.nickname thread_author, thread_created, thread_message, votes FROM thread t
		JOIN forum f ON t.forum = f.forum_id
		JOIN forum_user u ON t.thread_author = u.forum_user_id
		WHERE t.thread_id = $1
		`, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return res, &RecordNotFoundError{"Thread", fmt.Sprintf("%v", id)}
		}
		return res, err
	}
	return res, nil
}

func GetThreadBySlug(s string) (models.Thread, error) {
	res := models.Thread{}
	err := db.Get(&res, `
		SELECT thread_id, forum_slug forum, thread_slug, thread_title, u.nickname thread_author, thread_created, thread_message, votes FROM thread t
		JOIN forum f ON t.forum = f.forum_id
		JOIN forum_user u ON t.thread_author = u.forum_user_id
		WHERE t.thread_slug = $1
	`, s)
	if err != nil {
		if err == sql.ErrNoRows {
			return res, &RecordNotFoundError{"Thread", s}
		}
		return res, err
	}
	return res, nil
}

func GetAllThreadsInForum(s string, params *models.ThreadQueryParams) ([]models.Thread, error) {
	res := []models.Thread{}
	// check existence of forum
	_, err := GetForumBySlug(s)
	if err != nil {
		return res, err
	}

	q := `
		SELECT thread_id, forum_slug forum, thread_slug, thread_title, u.nickname thread_author, thread_created, thread_message, votes FROM thread t
		JOIN forum f ON t.forum = f.forum_id
		JOIN forum_user u ON t.thread_author = u.forum_user_id
		WHERE forum_slug = $1 `
	var nt time.Time
	if params.Since != nt {
		if params.Desc {
			q += "AND thread_created <= $2\n"
		} else {
			q += "AND thread_created >= $2\n"
		}
	}
	q += "ORDER BY thread_created "
	if params.Desc {
		q += "DESC"
	}
	if params.Limit != 0 {
		q += fmt.Sprintf("\nLIMIT %v", params.Limit)
	}
	if params.Since == nt {
		err = db.Select(&res, q, s)
	} else {
		err = db.Select(&res, q, s, params.Since)
	}
	if err != nil {
		return res, err
	}
	return res, nil
}

func GetThreadBySlugOrID(slugOrID string) (models.Thread, error) {
	res, err := GetThreadBySlug(slugOrID)
	if err != nil {
		if _, ok := err.(*RecordNotFoundError); ok {
			id, convErr := strconv.Atoi(slugOrID)
			if convErr != nil {
				return res, err
			}
			res, err = GetThreadByID(id)
			if err != nil {
				return res, err
			}
		} else {
			return res, err
		}
	}
	return res, nil
}

func UpdateThread(t *models.Thread, path string) (models.Thread, error) {
	res := models.Thread{}
	res, err := GetThreadBySlugOrID(path)
	if err != nil {
		return res, err
	}
	if t.ThreadTitle != "" {
		_, err := db.Exec(`
			UPDATE thread SET thread_title = $1 WHERE thread_id = $2
		`, t.ThreadTitle, res.ThreadID)
		if err != nil {
			return res, err
		}
		res.ThreadTitle = t.ThreadTitle
	}
	if t.ThreadMessage != "" {
		_, err := db.Exec(`
			UPDATE thread SET thread_message = $1 WHERE thread_id = $2
		`, t.ThreadMessage, res.ThreadID)
		if err != nil {
			return res, err
		}
		res.ThreadMessage = t.ThreadMessage
	}
	return res, nil
}

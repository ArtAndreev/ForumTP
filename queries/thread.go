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
	if t.Forum == "" || t.Title == "" || t.Author == "" {
		return res, &NullFieldError{"Thread", "some value(-s) is/are null"}
	}

	// check existence of forum
	f, err := GetForumBySlug(t.Forum)
	if err != nil {
		return res, err
	}

	// check existence of user
	u, err := GetUserByNickname(t.Author)
	if err != nil {
		return res, err
	}

	// check existence of thread
	if t.Slug != nil {
		res, err = GetThreadBySlug(*t.Slug)
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
		INSERT INTO thread (forum, slug, title, author, created, message)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING thread_id`,
		f.ForumID, t.Slug, t.Title, u.ForumUserID, t.Created, t.Message)
	if err != nil {
		return res, err
	}

	_, err = db.Exec(`
		UPDATE forum SET threads = threads + 1
		WHERE forum_id = $1
		`, f.ForumID)

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
		SELECT thread_id, f.slug forum, t.slug, t.title, u.nickname author, created, message, votes FROM thread t
		JOIN forum f ON t.forum = f.forum_id
		JOIN forum_user u ON t.author = u.forum_user_id
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
		SELECT thread_id, f.slug forum, t.slug, t.title, u.nickname author, created, message, votes FROM thread t
		JOIN forum f ON t.forum = f.forum_id
		JOIN forum_user u ON t.author = u.forum_user_id
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
		SELECT thread_id, f.slug forum, t.slug, t.title, u.nickname author, created, message, votes FROM thread t
		JOIN forum f ON t.forum = f.forum_id
		JOIN forum_user u ON t.author = u.forum_user_id
		WHERE t.slug = $1
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
		SELECT thread_id, f.slug forum, t.slug, t.title, u.nickname author, created, message, votes FROM thread t
		JOIN forum f ON t.forum = f.forum_id
		JOIN forum_user u ON t.author = u.forum_user_id
		WHERE f.slug = $1 `
	var nt time.Time
	if params.Since != nt {
		if params.Desc {
			q += "AND created <= $2\n"
		} else {
			q += "AND created >= $2\n"
		}
	}
	q += "ORDER BY created "
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
	if t.Title != "" {
		_, err := db.Exec(`
			UPDATE thread SET title = $1 WHERE thread_id = $2
		`, t.Title, res.ThreadID)
		if err != nil {
			return res, err
		}
		res.Title = t.Title
	}
	if t.Message != "" {
		_, err := db.Exec(`
			UPDATE thread SET message = $1 WHERE thread_id = $2
		`, t.Message, res.ThreadID)
		if err != nil {
			return res, err
		}
		res.Message = t.Message
	}
	return res, nil
}

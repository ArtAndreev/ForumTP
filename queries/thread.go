package queries

import (
	"database/sql"
	"fmt"
	"time"

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

func GetThreadBySlug(s string) (models.Thread, error) {
	res := models.Thread{}
	err := db.Get(&res, `
		SELECT thread_id, f.slug forum, t.slug, t.title, u.nickname author, created, message, votes FROM thread t
		JOIN forum f ON t.forum = f.forum_id
		JOIN forum_user u ON t.author = u.forum_user_id
		WHERE lower(t.slug) = lower($1)
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
		WHERE lower(f.slug) = lower($1) `
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
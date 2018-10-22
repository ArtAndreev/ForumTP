package queries

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/ArtAndreev/ForumTP/models"
)

func CreatePosts(p *[]models.Post, path string) ([]models.Post, error) {
	res := []models.Post{}
	// get thread by slug or id
	t, err := GetThreadBySlugOrID(path)
	if err != nil {
		return res, err
	}
	// get forum
	f, err := GetForumBySlug(t.Forum)
	if err != nil {
		return res, err
	}

	tx, err := db.Beginx()
	if err != nil {
		return res, err
	}
	defer tx.Rollback()

	// get current time, we'll use it for all inserted messages
	now := time.Time{}
	nr := tx.QueryRow("SELECT * FROM now()")
	err = nr.Scan(&now)
	if err != nil {
		return res, err
	}

	for _, v := range *p {
		// get user
		u, err := txGetUserByNickname(v.Author, tx)
		if err != nil {
			return res, err
		}
		// check parent message belongs to the same thread
		if v.Parent != 0 {
			parent, err := txGetPostByID(v.Parent, tx)
			switch err.(type) {
			case *RecordNotFoundError:
				return res, ErrParentPostIsNotInThisThread
			}
			if parent.Thread != t.ThreadID {
				return res, ErrParentPostIsNotInThisThread
			}
		}
		// insert
		qres := tx.QueryRow(`
			INSERT INTO post (forum, thread, parent, author, created, message)
			VALUES ($1, $2, $3, $4, $5, $6) RETURNING post_id`,
			f.ForumID, t.ThreadID, v.Parent, u.ForumUserID, now, v.Message)

		lastInsertedID := 0
		err = qres.Scan(&lastInsertedID)
		if err != nil {
			return res, err
		}

		// get new res
		last, err := txGetPostByID(lastInsertedID, tx)
		if err != nil {
			return res, err
		}
		res = append(res, last)
	}

	return res, tx.Commit()
}

func GetPostByID(id int) (models.Post, error) {
	res := models.Post{}
	err := db.Get(&res, `
		SELECT post_id, f.slug forum, thread, parent, u.nickname author, created, is_edited, message FROM post p
		JOIN forum f ON p.forum = f.forum_id
		JOIN forum_user u ON p.author = u.forum_user_id
		WHERE post_id = $1
		`, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return res, &RecordNotFoundError{"Post", fmt.Sprintf("%v", id)}
		}
		return res, err
	}

	return res, nil
}

func txGetPostByID(id int, tx *sqlx.Tx) (models.Post, error) {
	res := models.Post{}
	err := tx.Get(&res, `
		SELECT post_id, f.slug forum, thread, parent, u.nickname author, created, is_edited, message FROM post p
		JOIN forum f ON p.forum = f.forum_id
		JOIN forum_user u ON p.author = u.forum_user_id
		WHERE post_id = $1
		`, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return res, &RecordNotFoundError{"Post", fmt.Sprintf("%v", id)}
		}
		return res, err
	}

	return res, nil
}

package queries

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/ArtAndreev/ForumTP/models"
)

func CreatePosts(p *[]models.Post, path string) ([]models.Post, error) {
	res := []models.Post{}
	// get thread by slug
	t, err := GetThreadBySlug(path)
	if err != nil {
		if _, ok := err.(*RecordNotFoundError); ok {
			id, convErr := strconv.Atoi(path)
			if convErr != nil {
				return res, err
			}
			t, err = GetThreadByID(id)
			if err != nil {
				return res, err
			}
		} else {
			return res, err
		}
	}
	// get forum
	f, err := GetForumBySlug(t.Forum)
	if err != nil {
		return res, err
	}

	tx, err := db.Begin()
	if err != nil {
		return res, err
	}
	defer tx.Rollback()

	// get current time, we'll use it for all inserted messages
	now := time.Time{}
	err = db.Get(&now, "SELECT * FROM now()")
	if err != nil {
		return res, err
	}

	for _, v := range *p {
		// get user
		u, err := GetUserByNickname(v.Author)
		if err != nil {
			return res, err
		}
		// check parent message belongs to the same thread
		if v.Parent != 0 {
			parent, err := GetPostByID(v.Parent)
			switch err.(type) {
			case *RecordNotFoundError:
				return res, ErrParentPostIsNotInThisThread
			}
			if parent.Thread != t.ThreadID {
				return res, ErrParentPostIsNotInThisThread
			}
		}
		// insert
		qres, err := db.Query(`
			INSERT INTO post (forum, thread, parent, author, created, message)
			VALUES ($1, $2, $3, $4, $5, $6) RETURNING post_id`,
			f.ForumID, t.ThreadID, v.Parent, u.ForumUserID, now, v.Message)
		if err != nil {
			return res, err
		}

		_, err = db.Exec(`
			UPDATE forum SET posts = posts + 1
			WHERE forum_id = $1
		`, f.ForumID)

		lastInsertedID, err := getLastInsertedID(qres)
		if err != nil {
			return res, err
		}
		// get new res
		last, err := GetPostByID(lastInsertedID)
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
			return res, &RecordNotFoundError{"Thread", fmt.Sprintf("%v", id)}
		}
		return res, err
	}

	return res, nil
}

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
			INSERT INTO post (forum, thread, parent, post_author, post_created, post_message)
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
		SELECT post_id, forum_slug forum, thread, parent, u.nickname post_author, post_created, is_edited, post_message FROM post p
		JOIN forum f ON p.forum = f.forum_id
		JOIN forum_user u ON post_author = u.forum_user_id
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
		SELECT post_id, forum_slug forum, thread, parent, u.nickname post_author, post_created, is_edited, post_message FROM post p
		JOIN forum f ON p.forum = f.forum_id
		JOIN forum_user u ON post_author = u.forum_user_id
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

func GetPostInfoByID(id int, params *models.PostQueryArgs) (models.PostInfo, error) {
	res := models.PostInfo{Post: &models.Post{}}
	q := `
		SELECT post_id, forum_slug forum, thread, parent, u.nickname post_author, post_created, is_edited, post_message `
	// var queryArgs map[string]bool
	// for _, v := range params.Related {
	// 	queryArgs[v] = true
	// }
	// if _, ok := queryArgs["author"]; ok {
	// 	q += ", u.about, u.email, u.fullname"
	// }
	// if _, ok := queryArgs["thread"]; ok {
	// 	q += ", t.thread_slug, t.thread_title, t.thread_author, t.thread_created, t.thread_message, votes "
	// }
	// if _, ok := queryArgs["forum"]; ok {
	// 	q += ", f.forum_title, f.forum_slug, f.forum_user "
	// }
	q += "FROM post p"
	// if _, ok := queryArgs["thread"]; ok {
	// 	q += `
	// 	JOIN thread t ON p.thread = t.thread_id`
	// }
	q += `
		JOIN forum f ON p.forum = f.forum_id
		JOIN forum_user u ON post_author = u.forum_user_id
		WHERE post_id = $1`
	row := db.QueryRow(q, id)
	err := row.Scan(&res.Post.PostID, &res.Post.Forum, &res.Post.Thread, &res.Post.Parent,
		&res.Post.Author, &res.Post.Created, &res.Post.IsEdited, &res.Post.Message)
	if err != nil {
		if err == sql.ErrNoRows {
			return res, &RecordNotFoundError{"Post", fmt.Sprintf("%v", id)}
		}
		return res, err
	}

	return res, nil
}

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
		u, err := txGetUserByNickname(v.PostAuthor, tx)
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
			f.ForumID, t.ThreadID, v.Parent, u.ForumUserID, now, v.PostMessage)

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

func GetPostInfoByID(id int, params *[]string) (models.PostInfo, error) {
	res := models.PostInfo{}
	q := `
		SELECT post_id, forum_slug, thread, parent, u.nickname post_author, post_created, is_edited, post_message `
	queryArgs := make(map[string]bool, 3)
	for _, v := range *params {
		queryArgs[v] = true
	}
	if _, ok := queryArgs["user"]; ok {
		q += ", u.about, u.email, u.fullname "
	}
	if _, ok := queryArgs["thread"]; ok {
		q += ", thread_id, thread_slug, thread_title, tu.nickname thread_author, thread_created, thread_message, votes "
	}
	if _, ok := queryArgs["forum"]; ok {
		q += ", forum_title, fu.nickname forum_user, threads, posts "
	}
	q += `
		FROM post p
		JOIN forum f ON p.forum = f.forum_id
		JOIN forum_user u ON post_author = u.forum_user_id`
	if _, ok := queryArgs["thread"]; ok {
		// join for getting thread and thread_author (related with forum_user)
		q += `
		JOIN thread t ON p.thread = t.thread_id
		JOIN forum_user tu ON t.thread_author = tu.forum_user_id`
	}
	if _, ok := queryArgs["forum"]; ok {
		// join for getting thread and thread_author (related with forum_user)
		q += `
		JOIN forum_user fu ON f.forum_user = fu.forum_user_id`
	}

	q += `
		WHERE post_id = $1`
	row := db.QueryRowx(q, id)

	all := &models.PostInfoAllFields{}
	err := row.StructScan(all)
	if err != nil {
		if err == sql.ErrNoRows {
			return res, &RecordNotFoundError{"Post", fmt.Sprintf("%v", id)}
		}
		return res, err
	}

	res.Post = &models.Post{
		PostID:      all.PostID,
		Forum:       all.ForumSlug,
		Thread:      all.Post.Thread,
		Parent:      all.Parent,
		PostAuthor:  all.PostAuthor,
		PostCreated: all.PostCreated,
		IsEdited:    all.IsEdited,
		PostMessage: all.PostMessage,
	}
	if _, ok := queryArgs["user"]; ok {
		res.Author = &models.ForumUser{
			Nickname: all.PostAuthor,
			Fullname: all.Fullname,
			Email:    all.Email,
			About:    all.About,
		}
	}
	if _, ok := queryArgs["thread"]; ok {
		res.Thread = &models.Thread{
			ThreadID:      all.ThreadID,
			Forum:         all.ForumSlug,
			ThreadSlug:    all.ThreadSlug,
			ThreadTitle:   all.ThreadTitle,
			ThreadAuthor:  all.ThreadAuthor,
			ThreadCreated: all.ThreadCreated,
			ThreadMessage: all.ThreadMessage,
		}
	}
	if _, ok := queryArgs["forum"]; ok {
		res.Forum = &models.Forum{
			ForumTitle: all.ForumTitle,
			ForumSlug:  all.ForumSlug,
			ForumUser:  all.Forum.ForumUser,
			Threads:    all.Threads,
			Posts:      all.Posts,
		}
	}

	return res, nil
}

func UpdatePostByID(id int, p *models.Post) (models.Post, error) {
	res := models.Post{}
	res, err := GetPostByID(id)
	if err != nil {
		return res, err
	}
	if res.PostMessage == p.PostMessage || p.PostMessage == "" {
		return res, nil
	}
	_, err = db.Exec(`
		UPDATE post SET post_message = $1, is_edited = TRUE WHERE post_id = $2;`,
		p.PostMessage, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return res, &RecordNotFoundError{"Post", fmt.Sprintf("%v", id)}
		}
		return res, err
	}
	res.IsEdited = true
	res.PostMessage = p.PostMessage

	// fid, err := strconv.Atoi(res.Forum)
	// if err != nil {
	// 	return res, err
	// }
	// res.Forum, err = GetForumSlugByID(fid)
	// if err != nil {
	// 	return res, err
	// }

	// uid, err := strconv.Atoi(res.PostAuthor)
	// if err != nil {
	// 	return res, err
	// }
	// res.PostAuthor, err = GetUserNicknameByID(uid)
	// if err != nil {
	// 	return res, err
	// }

	return res, nil
}

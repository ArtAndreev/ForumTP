package queries

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

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

	// get current time, we'll use it for all inserted messages
	now := time.Time{}
	nr := db.QueryRow("SELECT * FROM now()")
	err = nr.Scan(&now)
	if err != nil {
		return res, err
	}

	tx, err := db.Beginx()
	if err != nil {
		return res, err
	}
	defer tx.Rollback()

	uIDs := make([]int, len(*p))
	for k, v := range *p {
		// get user
		u, err := txGetUserByNickname(v.PostAuthor, tx)
		if err != nil {
			return res, err
		}
		uIDs[k] = u.ForumUserID

		// check parent message belongs to the same thread
		parent := models.Post{}
		if v.Parent != 0 {
			parent, err = txGetPostByID(v.Parent, tx)
			switch err.(type) {
			case *RecordNotFoundError:
				return res, ErrParentPostIsNotInThisThread
			}
			if parent.Thread != t.ThreadID {
				return res, ErrParentPostIsNotInThisThread
			}
		}
		// get path
		(*p)[k].Path = parent.Path

		// get new primary key id
		val := tx.QueryRow("SELECT nextval(pg_get_serial_sequence('post', 'post_id'))")
		err = val.Scan(&(*p)[k].PostID)
		if err != nil {
			return res, err
		}
		(*p)[k].Path = append((*p)[k].Path, int64((*p)[k].PostID))

		// update result forum id
		(*p)[k].Forum = f.ForumSlug
		(*p)[k].Thread = t.ThreadID
		(*p)[k].PostCreated = now
	}

	stmt, err := tx.Prepare(pq.CopyIn("post", "post_id", "forum", "thread", "parent", "path", "post_author", "post_created", "post_message"))
	if err != nil {
		return res, err
	}
	defer stmt.Close()

	for k, v := range *p {
		_, err = stmt.Exec(v.PostID, f.ForumID, t.ThreadID, v.Parent, pq.Array(v.Path), uIDs[k], now, v.PostMessage)
		if err != nil {
			return res, err
		}
	}

	_, err = stmt.Exec()
	if err != nil {
		return res, err
	}

	res = *p

	return res, tx.Commit()
}

func GetPostByID(id int) (models.Post, error) {
	res := models.Post{}
	qres := db.QueryRow(`
		SELECT post_id, forum_slug forum, thread, parent, path, u.nickname post_author, post_created, is_edited, post_message FROM post p
		JOIN forum f ON p.forum = f.forum_id
		JOIN forum_user u ON post_author = u.forum_user_id
		WHERE post_id = $1
		`, id)
	err := qres.Scan(&res.PostID, &res.Forum, &res.Thread, &res.Parent, pq.Array(&res.Path), &res.PostAuthor,
		&res.PostCreated, &res.IsEdited, &res.PostMessage)
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
	qres := tx.QueryRow(`
		SELECT post_id, forum_slug forum, thread, parent, path, u.nickname post_author, post_created, is_edited, post_message FROM post p
		JOIN forum f ON p.forum = f.forum_id
		JOIN forum_user u ON post_author = u.forum_user_id
		WHERE post_id = $1
		`, id)
	err := qres.Scan(&res.PostID, &res.Forum, &res.Thread, &res.Parent, pq.Array(&res.Path), &res.PostAuthor,
		&res.PostCreated, &res.IsEdited, &res.PostMessage)
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

	return res, nil
}

func GetThreadPosts(slug_or_id string, args *models.ThreadPostsQueryArgs) ([]models.Post, error) {
	res := []models.Post{}
	t, err := GetThreadBySlugOrID(slug_or_id)
	if err != nil {
		return res, err
	}
	q := `
		SELECT p.post_id, forum_slug forum, p.thread, p.parent, u.nickname post_author, p.post_created, p.is_edited, p.post_message FROM post p
		JOIN forum f ON p.forum = f.forum_id
		JOIN forum_user u ON p.post_author = u.forum_user_id`
	switch args.Sort {
	case "tree":
		if args.Since > 0 {
			q += `
				JOIN post sp ON sp.post_id = $2
				WHERE p.path `
			if args.Desc {
				q += "< sp.path "
			} else {
				q += "> sp.path "
			}
			q += "AND p.thread = $1"
		} else {
			q += `
				WHERE p.thread = $1`
		}
		q += `
			ORDER BY p.path `
		if args.Desc {
			q += "DESC"
		}
		if args.Limit > 0 {
			if args.Since > 0 {
				q += `
					LIMIT $3`
			} else {
				q += `
					LIMIT $2`
			}
		}
	case "parent_tree":
		q = `
			WITH root_posts AS (
			SELECT p.post_id FROM post p`
		if args.Since > 0 {
			q += `
				JOIN post sp ON sp.post_id = $2`
			if args.Desc {
				q += `WHERE p.path[1] < sp.path[1] AND p.thread = $1 AND p.parent = 0 `
			} else {
				q += `WHERE p.path > sp.path AND p.thread = $1 AND p.parent = 0 `
			}
		} else {
			q += `
				WHERE p.thread = $1 AND p.parent = 0 `
		}
		q += `
			ORDER BY p.path[1] `
		if args.Desc {
			q += "DESC"
		}
		if args.Limit > 0 {
			if args.Since > 0 {
				q += `
					LIMIT $3`
			} else {
				q += `
					LIMIT $2`
			}
		}
		q += `
			)
			SELECT p.post_id, forum_slug forum, p.thread, p.parent, u.nickname post_author, p.post_created, p.is_edited, p.post_message FROM post p
			JOIN forum f ON p.forum = f.forum_id
			JOIN forum_user u ON p.post_author = u.forum_user_id`
		q += ` 
			WHERE p.path[1] in (select post_id from root_posts)`
		q += `
			ORDER BY p.path[1] `
		if args.Desc {
			q += "DESC"
		}
		q += `, p.path`
	default: // flat
		q += `
			WHERE p.thread = $1 `
		if args.Since > 0 {
			if args.Desc {
				q += "AND p.post_id < $2"
			} else {
				q += "AND p.post_id > $2"
			}
		}
		q += `
			ORDER BY p.post_created `
		if args.Desc {
			q += "DESC"
		}
		q += `, p.post_id `
		if args.Desc {
			q += "DESC"
		}
		if args.Limit > 0 {
			if args.Since > 0 {
				q += `
					LIMIT $3`
			} else {
				q += `
					LIMIT $2`
			}
		}
	}
	if args.Since > 0 {
		if args.Limit > 0 {
			err = db.Select(&res, q, t.ThreadID, args.Since, args.Limit)
		} else {
			err = db.Select(&res, q, t.ThreadID, args.Since)
		}

	} else {
		if args.Limit > 0 {
			err = db.Select(&res, q, t.ThreadID, args.Limit)
		} else {
			err = db.Select(&res, q, t.ThreadID)
		}
	}
	if err != nil {
		return res, err
	}

	return res, nil
}

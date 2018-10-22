package queries

import (
	"database/sql"

	"github.com/ArtAndreev/ForumTP/models"
)

func CreateForum(f *models.Forum) (models.Forum, error) {
	var res models.Forum
	// check not null constraint
	if f.Title == "" || f.Slug == "" || f.ForumUser == "" {
		return res, &NullFieldError{"Forum", "title and/or slug and/or user"}
	}

	// check existence of forum
	res, err := GetForumBySlug(f.Slug)
	if err != nil {
		if _, ok := err.(*RecordNotFoundError); !ok {
			return res, err // db error
		}
	} else { // record exists
		return res, &UniqueFieldValueAlreadyExistsError{"Forum", "title and/or slug"}
	}

	// check existence of user
	u, err := GetUserByNickname(f.ForumUser)
	if err != nil {
		return res, err
	}

	// insert
	_, err = db.Exec(
		"INSERT INTO forum (title, slug, forum_user) VALUES ($1, $2, $3)",
		f.Title, f.Slug, u.ForumUserID)
	if err != nil {
		return res, err
	}

	// return res
	res = models.Forum{
		Title:     f.Title,
		Slug:      f.Slug,
		ForumUser: u.Nickname,
	}

	return res, nil
}

func GetForumBySlug(s string) (models.Forum, error) {
	res := models.Forum{}
	err := db.Get(&res, `
		SELECT forum_id, title, slug, u.nickname forum_user, threads, posts FROM forum f 
		JOIN forum_user u ON f.forum_user = u.forum_user_id 
		WHERE slug = $1`,
		s)
	if err != nil {
		if err == sql.ErrNoRows {
			return res, &RecordNotFoundError{"Forum", s}
		}
		return res, err
	}
	return res, nil
}

package queries

import (
	"github.com/ArtAndreev/ForumTP/models"
)

func VoteForPost(v *models.Vote, path string) (models.Thread, error) {
	res := models.Thread{}
	if v.Nickname == "" {
		return res, &NullFieldError{"Vote", "nickname"}
	}
	if v.Voice != -1 && v.Voice != 1 {
		return res, &ValidationError{"Vote", "voice"}
	}
	// get thread by slug or id
	t, err := GetThreadBySlugOrID(path)
	if err != nil {
		return res, err
	}

	// get user
	u, err := GetUserByNickname(v.Nickname)
	if err != nil {
		return res, err
	}

	tx, err := db.Beginx()
	if err != nil {
		return res, err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT INTO vote VALUES ($1, $2, $3)
		ON CONFLICT (nickname, thread) DO UPDATE SET voice = $3`,
		u.ForumUserID, t.ThreadID, v.Voice)

	res, err = txGetThreadByID(t.ThreadID, tx)
	if err != nil {
		return res, err
	}
	return res, tx.Commit()
}

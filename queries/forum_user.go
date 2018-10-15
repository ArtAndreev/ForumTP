package queries

import (
	"github.com/ArtAndreev/ForumTP/models"
)

func CreateUser(u *models.BaseForumUser) error {
	_, err := db.NamedExec(
		"INSERT INTO forum_user (nickname, fullname, email, about)"+
			"VALUES (:nickname, :fullname, :email, :about)",
		u)
	if err != nil {
		return err
	}
	return nil
}

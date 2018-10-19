package queries

import (
	"github.com/ArtAndreev/ForumTP/models"
)

func ClearDatabase() error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec("DELETE FROM vote")
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM post")
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM thread")
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM forum")
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM forum_user")
	if err != nil {
		return err
	}
	return tx.Commit()
}

func GetDatabaseStatus() (models.Status, error) {
	res := models.Status{}
	err := db.Get(&res.User, "SELECT COUNT(*) FROM forum_user")
	if err != nil {
		return res, err
	}
	err = db.Get(&res.Forum, "SELECT COUNT(*) FROM forum")
	if err != nil {
		return res, err
	}
	err = db.Get(&res.Thread, "SELECT COUNT(*) FROM thread")
	if err != nil {
		return res, err
	}
	err = db.Get(&res.Post, "SELECT COUNT(*) FROM post")
	if err != nil {
		return res, err
	}
	return res, nil
}

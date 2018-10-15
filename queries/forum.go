package queries

import (
	"github.com/ArtAndreev/ForumTP/models"
)

func CreateForum(f *models.BaseForum) error {
	id := 0
	_, err := db.Exec(
		"INSERT INTO forum (title, slug, user) VALUES ($1, $2, $3)",
		f.Title, f.Slug, id)
	if err != nil {
		panic(err)
	}

	return nil
}

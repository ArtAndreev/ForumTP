package models

type Forum struct {
	ForumID   int    `json:"-" db:"forum_id"`
	Title     string `json:"title"`
	Slug      string `json:"slug"`
	ForumUser string `json:"user" db:"forum_user"`
	Threads   int    `json:"threads"`
	Posts     int    `json:"posts"`
}

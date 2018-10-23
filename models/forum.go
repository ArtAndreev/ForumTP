package models

type Forum struct {
	ForumID   int    `json:"-" db:"forum_id"`
	Title     string `json:"title" db:"forum_title"`
	Slug      string `json:"slug" db:"forum_slug"`
	ForumUser string `json:"user" db:"forum_user"`
	Threads   int    `json:"threads"`
	Posts     int    `json:"posts"`
}

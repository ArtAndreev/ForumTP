package models

import (
	"time"
)

type Post struct {
	PostID   int       `json:"id" db:"post_id"`
	Forum    string    `json:"forum"`
	Thread   int       `json:"thread"`
	Parent   int       `json:"parent"`
	Author   string    `json:"author" db:"post_author"`
	Created  time.Time `json:"created" db:"post_created"`
	IsEdited bool      `json:"isEdited" db:"is_edited"`
	Message  string    `json:"message" db:"post_message"`
}

type PostInfo struct {
	Post   *Post      `json:"post,omitempty"`
	Forum  *Forum     `json:"forum,omitempty"`
	Thread *Thread    `json:"thread,omitempty"`
	Author *ForumUser `json:"author,omitempty"`
}

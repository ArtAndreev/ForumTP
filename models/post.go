package models

import (
	"time"
)

type Post struct {
	PostID   int       `json:"id" db:"post_id"`
	Forum    string    `json:"forum"`
	Thread   int       `json:"thread"`
	Parent   int       `json:"parent"`
	Author   string    `json:"author"`
	Created  time.Time `json:"created"`
	IsEdited bool      `json:"isEdited" db:"is_edited"`
	Message  string    `json:"message"`
}

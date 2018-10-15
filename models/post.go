package models

import (
	"time"
)

type Post struct {
	PostID   int       `json:"id"`
	Forum    int       `json:"forum"`
	Thread   int       `json:"thread"`
	Parent   int       `json:"parent"`
	Author   int       `json:"author"`
	Created  time.Time `json:"created"`
	IsEdited bool      `json:"isEdited"`
	Message  string    `json:"message"`
}

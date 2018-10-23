package models

import (
	"time"
)

type Thread struct {
	ThreadID int        `json:"id" db:"thread_id"`
	Forum    string     `json:"forum"`
	Slug     *string    `json:"slug,omitempty" db:"thread_slug"`
	Title    string     `json:"title" db:"thread_title"`
	Author   string     `json:"author" db:"thread_author"`
	Created  *time.Time `json:"created,omitempty" db:"thread_created"`
	Message  string     `json:"message" db:"thread_message"`
	Votes    int        `json:"votes"`
}

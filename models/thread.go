package models

import (
	"time"
)

type Thread struct {
	ThreadID int        `json:"id" db:"thread_id"`
	Forum    string     `json:"forum"`
	Slug     *string    `json:"slug,omitempty"`
	Title    string     `json:"title"`
	Author   string     `json:"author"`
	Created  *time.Time `json:"created,omitempty"`
	Message  string     `json:"message"`
	Votes    int        `json:"votes"`
}

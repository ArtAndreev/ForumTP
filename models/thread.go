package models

import "time"

type Thread struct {
	ThreadID int       `json:"id"`
	Forum    int       `json:"forum"`
	Slug     string    `json:"slug"`
	Title    string    `json:"title"`
	Author   int       `json:"author"`
	Created  time.Time `json:"created"`
	Message  string    `json:"message"`
	Votes    int       `json:"votes"`
}

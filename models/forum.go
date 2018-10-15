package models

type Forum struct {
	BaseForum
	Threads int `json:"threads"`
	Posts   int `json:"posts"`
}

type BaseForum struct {
	Title     string `json:"title"`
	Slug      string `json:"slug"`
	ForumUser int    `json:"user"`
}

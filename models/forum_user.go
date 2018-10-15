package models

type ForumUser struct {
	ForumUserID int `json:"-"`
	BaseForumUser
}

type BaseForumUser struct {
	Nickname string `json:"nickname"`
	Fullname string `json:"fullname"`
	Email    string `json:"email"`
	About    string `json:"about"`
}

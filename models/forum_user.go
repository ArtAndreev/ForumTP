package models

type ForumUser struct {
	Nickname string `json:"nickname"`
	Fullname string `json:"fullname"`
	Email    string `json:"email"`
	About    string `json:"about"`
}

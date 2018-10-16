package models

type ForumUser struct {
	ForumUserID int    `json:"-" db:"forum_user_id"`
	Nickname    string `json:"nickname"`
	Fullname    string `json:"fullname"`
	Email       string `json:"email"`
	About       string `json:"about"`
}
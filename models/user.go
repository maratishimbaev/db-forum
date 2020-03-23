package models

type User struct {
	About string `json:"about"`
	Email string `json:"email"`
	FullName string `json:"fullname"`
	Nickname string `json:"nickname"`
}

type UserUpdate struct {
	About string `json:"about"`
	Email string `json:"email"`
	FullName string `json:"fullname"`
}

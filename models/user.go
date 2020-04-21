package models

type User struct {
	About    string `json:"about,omitempty"`
	Email    string `json:"email,omitempty"`
	FullName string `json:"fullname,omitempty"`
	Nickname string `json:"nickname,omitempty"`
}

package models

type Thread struct {
	Author string `json:"author"`
	Created string `json:"created"`
	Forum string `json:"forum"`
	ID uint64 `json:"id"`
	Message string `json:"message"`
	Slug string `json:"slug"`
	Title string `json:"title"`
	Votes uint64 `json:"votes"`
}

package models

type Forum struct {
	Posts uint64 `json:"posts"`
	Slug string `json:"slug"`
	Threads uint64 `json:"threads"`
	Title string `json:"title"`
	User string `json:"user"`
}

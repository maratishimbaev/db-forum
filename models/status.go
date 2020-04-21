package models

type Status struct {
	ForumCount  uint64 `json:"forum"`
	PostCount   uint64 `json:"post"`
	ThreadCount uint64 `json:"thread"`
	UserCount   uint64 `json:"user"`
}

package models

type Forum struct {
	Posts   uint64 `json:"posts,omitempty"`
	Slug    string `json:"slug,omitempty"`
	Threads uint64 `json:"threads,omitempty"`
	Title   string `json:"title,omitempty"`
	User    string `json:"user,omitempty"`
}

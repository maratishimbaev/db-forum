package models

import "time"

type Thread struct {
	Author  string    `json:"author,omitempty"`
	Created time.Time `json:"created,omitempty"`
	Forum   string    `json:"forum,omitempty"`
	ID      uint64    `json:"id,omitempty"`
	Message string    `json:"message,omitempty"`
	Slug    string    `json:"slug,omitempty"`
	Title   string    `json:"title,omitempty"`
	Votes   int64    `json:"votes,omitempty"`
}

package models

import "time"

type Post struct {
	Author string `json:"author"`
	Created time.Time `json:"created"`
	Forum string `json:"forum"`
	ID uint64 `json:"id"`
	IsEdited bool `json:"isEdited"`
	Message string `json:"message"`
	Parent uint64 `json:"parent"`
	Thread uint64 `json:"thread"`
}

type PostFull struct {
	Author User `json:"author"`
	Forum Forum `json:"forum"`
	Post Post `json:"post"`
	Thread Thread `json:"thread"`
}

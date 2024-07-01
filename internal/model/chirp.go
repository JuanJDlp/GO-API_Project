package model

type Chirp struct {
	Id   int    `json:"id" db:"id"`
	Body string `json:"body" db:"body"`
	Author string `json:"author" db:"author"`
}

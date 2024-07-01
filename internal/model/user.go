package model

type User struct {
	Id    int    `json:"id" db:"id"`
	Email string `json:"email" db:"email"`
	Password string `json:"password" db:"password"`
	Expires_in_seconds *int `json:"expires_in_seconds"`
	Is_Chirpy_red bool `json:"is_chirpy_red" db:"membership"`
}

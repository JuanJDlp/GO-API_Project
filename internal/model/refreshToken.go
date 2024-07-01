package model

import "time"

type RefreshToken struct {
	UserEmail      string    `db:"userid"`
	Token          string    `db:"token"`
	ExpirationTime time.Time `db:"expirationtime"`
}

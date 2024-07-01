package database

import (
	"chirpy/internal/model"
	"fmt"
	"sync"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type DataBaseStructure struct {
	connection *sqlx.DB
	mu         *sync.RWMutex
	lastId     int
	userLastId *userCounter
}

func NewDb(user, dbname, password, host, port string) (*DataBaseStructure, error) {
	dbConnection, err := sqlx.Connect("postgres", fmt.Sprintf("user=%s dbname=%s sslmode=disable password=%s host=%s port=%s", user, dbname, password, host, port))
	if err != nil {
		return nil, err
	}
	return &DataBaseStructure{
		dbConnection,
		&sync.RWMutex{},
		1,
		&userCounter{lastId: 1},
	}, nil
}

// CreateChirp creates a new chirp and saves it to disk
func (db *DataBaseStructure) CreateChirp(body, author string) (model.Chirp, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	chirp := model.Chirp{
		Id:     db.lastId,
		Body:   body,
		Author: author,
	}
	db.lastId++

	_, err := db.connection.Exec("INSERT INTO chirps VALUES ($1,$2,$3)", chirp.Id, chirp.Body, chirp.Author)
	if err != nil {
		return model.Chirp{}, err
	}

	return chirp, nil
}

// GetChirps returns all chirps in the database
func (db *DataBaseStructure) GetChirps() ([]model.Chirp, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	chirps := []model.Chirp{}
	err := db.connection.Select(&chirps, "SELECT * FROM chirps")
	if err != nil {
		return []model.Chirp{}, err
	}
	return chirps, nil
}

func (db *DataBaseStructure) ResetDatabase() error {
	_, err := db.connection.Exec("DELETE FROM chirps")
	if err != nil {
		return err
	}
	_, err = db.connection.Exec("DELETE FROM users")
	if err != nil {
		return err
	}
	return nil
}

func (db *DataBaseStructure) GetChirp(id int) (model.Chirp, error) {
	chirp := model.Chirp{}
	err := db.connection.Get(&chirp, "SELECT * FROM chirps WHERE id = $1", id)
	if err != nil {
		return model.Chirp{}, err
	}

	return chirp, nil

}

func (db *DataBaseStructure) UpdateUser(oldEmail, newEmail, newPassword string) error {
	_, err := db.connection.Exec("UPDATE users SET email = $1, password = $2 WHERE email = $3", newEmail, newPassword, oldEmail)

	if err != nil {
		return err
	}

	return nil
}

func (db *DataBaseStructure) GetRefreshToken(token string) (*model.RefreshToken, error) {
	tokens := model.RefreshToken{}
	err := db.connection.Get(&tokens, "SELECT * FROM refresh_tokens WHERE token = $1", token)
	if err != nil {
		return nil, err
	}
	return &tokens, nil
}

func (db *DataBaseStructure) SaveRefreshToken(newRefreshToken *model.RefreshToken) error {
	_, err := db.connection.Exec("INSERT INTO refresh_tokens VALUES ($1,$2,$3)", newRefreshToken.UserEmail, newRefreshToken.Token, newRefreshToken.ExpirationTime)
	return err
}

func (db *DataBaseStructure) RemoveRefreshToken(token string) error {
	_, err := db.connection.Exec("DELETE FROM refresh_tokens WHERE token = $1", token)
	return err
}

func (db *DataBaseStructure) DeleteChirp(chripID int) error {
	_, err := db.connection.Exec("DELETE FROM chirps WHERE id = $1", chripID)
	return err
}

func (db *DataBaseStructure) SetUserMemberShip(userId int, membership bool) error {
	_, err := db.connection.Exec("UPDATE users SET membership = $2 WHERE id = $1", userId, membership)
	return err
}

func (db *DataBaseStructure) GetChirpsById(authorEmail string) ([]model.Chirp, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	chirps := []model.Chirp{}
	err := db.connection.Select(&chirps, "SELECT * FROM chirps WHERE author = $1", authorEmail)
	if err != nil {
		return []model.Chirp{}, err
	}
	return chirps, nil
}

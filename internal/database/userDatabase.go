package database

import (
	"chirpy/internal/model"
)

type userCounter struct {
	lastId int
}

func (db *DataBaseStructure) CreateUser(email, password string) (model.User, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	user := model.User{
		Id:            db.userLastId.lastId,
		Email:         email,
		Password:      password,
		Is_Chirpy_red: false,
	}
	db.userLastId.lastId++
	_, err := db.connection.Exec("INSERT INTO users (id , email , password, membership) VALUES ($1,$2,$3,$4)", user.Id, user.Email, user.Password , user.Is_Chirpy_red)
	if err != nil {
		return model.User{}, err
	}

	user.Password = ""

	return user, nil
}

func (db *DataBaseStructure) GetUsers() ([]model.User, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	users := []model.User{}
	err := db.connection.Select(&users, "SELECT * FROM users")
	if err != nil {
		return []model.User{}, err
	}
	return users, nil
}

func (db *DataBaseStructure) GetUser(email string) (model.User, error) {
	user := model.User{}
	err := db.connection.Get(&user, "SELECT * FROM users WHERE email = $1", email)
	if err != nil {
		return model.User{}, err
	}

	return user, nil

}

package api

import (
	"chirpy/internal/model"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	ApiConfig *ApiConfig
}
type UserWithNoPassword struct {
	Id    int
	Email string
	Token string
}

func (cc *UserService) GetUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	users, err := cc.ApiConfig.Db.GetUsers()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {

			log.Print(err)
			w.WriteHeader(300)
			w.Write([]byte("Nothing in the dabase"))
			return
		} else {
			w.WriteHeader(300)
			w.Write([]byte(err.Error()))
			return
		}
	}
	value, err := json.Marshal(users)
	if err != nil {
		w.WriteHeader(300)
		log.Print("ERROR DECODING")
		return
	}
	w.Write(value)
}

func (cc *UserService) CreateUser(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	user := model.User{}
	err := decoder.Decode(&user)

	if err != nil {
		w.WriteHeader(300)
		log.Print("ERROR DECODING")
		return
	}

	if validateInput(user.Email, w, r) {

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		password, _ := hashPasword(user.Password)
		_, err := cc.ApiConfig.Db.GetUser(user.Email)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(200)
			log.Print(err.Error())
			w.Write([]byte("This user alredy exists"))
			return
		}
		user, err := cc.ApiConfig.Db.CreateUser(user.Email, password)
		if err != nil {
			w.WriteHeader(300)
			log.Print("ERROR DECODING")
			log.Print("there was an error saving the user", err)
			return
		}
		response := struct {
			Id    int
			Email string
		}{
			user.Id,
			user.Email,
		}
		value, err := json.Marshal(response)
		if err != nil {
			w.WriteHeader(300)
			log.Print("ERROR DECODING")
			return
		}
		w.Write(value)
	}

}

func (cc *UserService) LogInUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	user := model.User{}
	if err := decoder.Decode(&user); err != nil {
		w.WriteHeader(300)
		log.Print("ERROR DECODING")
		log.Print("There was an error, please check the body of your request", err)
		return
	}

	userDb, err := cc.ApiConfig.Db.GetUser(user.Email)
	if err != nil {
		log.Print(err.Error())
		return
	}

	if valid := unhashPassword(user.Password, userDb.Password); valid {
		token, err := cc.ApiConfig.Jwt.CreateJwtTokenForUser(user)
		if err != nil {
			w.WriteHeader(501)
			log.Print(err.Error())
			log.Print("ERROR CREATING THE TOKEN")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		refreshToken, err := cc.ApiConfig.Jwt.CreateRefreshRoken(userDb)
		if err != nil {
			log.Print(err)
			refreshToken.Token = "ERR"
		}
		if err = cc.ApiConfig.Db.SaveRefreshToken(refreshToken); err != nil {
			log.Print(err)
			refreshToken.Token = "ERR"
		}

		response := struct {
			Id           int
			Email        string
			Token        string
			RefreshToken string
			MemberShip   bool
		}{
			user.Id,
			user.Email,
			token,
			refreshToken.Token,
			userDb.Is_Chirpy_red,
		}
		data, err := json.Marshal(response)
		if err != nil {
			w.WriteHeader(501)
			log.Print("ERROR MARSHELING")
			return
		}
		w.WriteHeader(200)
		w.Write(data)
	} else {

		w.WriteHeader(401)
		w.Write([]byte("ERROR"))
	}
}

func hashPasword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func unhashPassword(passwordInserted, passwordHashed string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(passwordHashed), []byte(passwordInserted))
	return err == nil
}

func (cc *UserService) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value(ContextUserKey).(string)
	params := model.User{}
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		w.WriteHeader(501)
		log.Print("ERROR DECODING THE REQUEST, PLEASE CHECK THE REQUEST")
		return
	}
	password, err := hashPasword(params.Password)
	if err != nil {
		w.WriteHeader(501)
		log.Print("THAT PASSWORD IS NOT VALID")
		return
	}
	err = cc.ApiConfig.Db.UpdateUser(id, params.Email, password)

	if err != nil {
		w.WriteHeader(501)
		log.Print("THE USER WAS NOT FOUND")
		return
	}

	w.WriteHeader(201)
	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Email string
	}{
		params.Email,
	}
	data, _ := json.Marshal(response)
	w.Write(data)

}

func (cc *UserService) RefreshToken(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := ExtractTokenString("Bearer",r.Header.Get("Authorization"))
	if err != nil {
		w.WriteHeader(401)
		log.Print(err)
		w.Write([]byte("Eror getting the token"))
		return
	}
	tokenRefreshStruct, err := cc.ApiConfig.Db.GetRefreshToken(refreshToken)
	if err != nil {

		w.WriteHeader(401)
		w.Write([]byte("error creating the refresh token"))
		return
	}

	newToken, err := cc.ApiConfig.Jwt.CreateJwtTokenForUser(model.User{
		Email: tokenRefreshStruct.UserEmail,
	})

	if err != nil {
		log.Print(err)
		w.WriteHeader(501)
		w.Write([]byte("Error creating the new jwt token"))
	}
	response := struct {
		Token string
	}{
		Token: newToken,
	}
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(response)
	if err != nil {
		log.Print(err)
		w.WriteHeader(501)
		w.Write([]byte("ERROR MARSHALING THE DATA"))
	}
	w.Write(data)

}

func badRequest(err error, w http.ResponseWriter) {
	w.WriteHeader(401)
	log.Print(err.Error())
	w.Write([]byte("BAD REQUEST"))
}

func (cc *UserService) RevokeRefreshToken(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := ExtractTokenString("Bearer",r.Header.Get("Authorization"))
	if err != nil {
		badRequest(err, w)
		return
	}
	err = cc.ApiConfig.Db.RemoveRefreshToken(refreshToken)
	if err != nil {
		log.Print(err)
		w.WriteHeader(501)
		w.Write([]byte("There was an error removing the token"))
	}
	w.WriteHeader(204)
}

func (cc *UserService) UpgradeUserMembership(w http.ResponseWriter, r *http.Request) {
	api_key, err := ExtractTokenString("ApiKey",r.Header.Get("Authorization"))
	if err != nil || api_key != os.Getenv("API_KEY") {
		w.WriteHeader(401)
		return
	}
	params := model.PolkaWebbHook{}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if params.Event != "user.upgraded" {
		w.WriteHeader(204)
		w.Write([]byte("event not understood"))
		return
	}

	err = cc.ApiConfig.Db.SetUserMemberShip(params.Data.User_id, true)
	if err != nil {
		log.Print(err)
		w.WriteHeader(404)
		w.Write([]byte("The user was not found"))
		return
	}
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("updated"))
}

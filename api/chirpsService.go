package api

import (
	"chirpy/internal/model"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type ChirpsService struct {
	ApiConfig *ApiConfig
}

func (cc *ChirpsService) GetChirp(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(300)
		w.Write([]byte("there was an error handeling the id"))
		return
	}
	chirp, err := cc.ApiConfig.Db.GetChirp(id)

	if err != nil {
		log.Print(err)
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(404)
		w.Write([]byte("the chirps was not found"))
		return
	}

	data, err := json.Marshal(chirp)
	if err != nil {
		log.Print(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(data)
}

func (cc *ChirpsService) GetChirps(w http.ResponseWriter, r *http.Request) {
	authorEmail := r.URL.Query().Get("author_id")
	var chirps []model.Chirp
	var err error = nil
	if authorEmail != "" {
		chirps, err = cc.ApiConfig.Db.GetChirpsById(authorEmail)
	} else {
		chirps, err = cc.ApiConfig.Db.GetChirps()
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
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
	value, err := json.Marshal(chirps)
	if err != nil {
		w.WriteHeader(300)
		log.Print("ERROR DECODING")
		return
	}
	w.Write(value)
}

func (cc *ChirpsService) CreateChirp(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	chrip := model.Chirp{
		Author: r.Context().Value(ContextUserKey).(string),
	}
	err := decoder.Decode(&chrip)

	if err != nil {
		w.WriteHeader(300)
		log.Print("ERROR DECODING")
		return
	}

	if validateInput(chrip.Body, w, r) {

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		chrip.Body = cleanBody(chrip.Body)
		chirp, err := cc.ApiConfig.Db.CreateChirp(chrip.Body, chrip.Author)
		if err != nil {
			w.WriteHeader(300)
			log.Print("ERROR DECODING")
			log.Print("there was an error saving the chirp", err)
			return
		}
		value, err := json.Marshal(chirp)
		if err != nil {
			w.WriteHeader(300)
			log.Print("ERROR DECODING")
			return
		}
		w.Write(value)
	}

}

func validateInput(chrips string, w http.ResponseWriter, _ *http.Request) bool {

	if len(chrips) > 140 {

		type returnValues struct {
			Error string `json:"error"`
		}

		response := returnValues{
			Error: "Chirp is too long",
		}

		data, err := json.Marshal(response)

		if err != nil {
			log.Print("ERROR MARSHALING")
			return false
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		w.Write(data)
		return false
	}

	return true
}

func cleanBody(body string) string {
	body = strings.ToLower(body)
	body = strings.ReplaceAll(body, "kerfuffle", "****")
	body = strings.ReplaceAll(body, "sharbert", "****")
	body = strings.ReplaceAll(body, "fornax", "****")
	return body
}

func (cc *ChirpsService) DeleteChirp(w http.ResponseWriter, r *http.Request) {
	chirpId, err := strconv.Atoi(r.PathValue("chirpID"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("There was a bad request with the chirp id"))
	}
	chirp, err := cc.ApiConfig.Db.GetChirp(chirpId)
	if err != nil {
		log.Print(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("The chirp id was not found"))
	}
	if chirp.Author == r.Context().Value(ContextUserKey).(string) {

		err = cc.ApiConfig.Db.DeleteChirp(chirpId)

		if err != nil {
			log.Print(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("It wasnt possible to delete the chirp"))
		}

		w.WriteHeader(204)
	} else {
		w.WriteHeader(403)
	}

}

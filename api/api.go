package api

import (
	"chirpy/internal/database"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

type ApiConfig struct {
	ServerHits int
	Jwt        *JwtService
	Db         *database.DataBaseStructure
}

func (ac *ApiConfig) MetricsSiteIncrement(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ac.ServerHits++
		next.ServeHTTP(w, r)
	})
}

func (ac *ApiConfig) WriteServerHits(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Hits: %d", ac.ServerHits)))
}

func (ac *ApiConfig) Reset(w http.ResponseWriter, _ *http.Request) {
	ac.ServerHits = 0

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reseted to 0"))
}

func (ac *ApiConfig) TemplateMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(200)
	templ, err := template.ParseFiles("./admins/index.html")

	if err != nil {
		log.Print(err)
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}

	err = templ.Execute(w, ac)
	if err != nil {
		log.Print(err)
	}

}

func (ac *ApiConfig) Healtz(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", " text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

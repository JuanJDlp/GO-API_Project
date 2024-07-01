package main

import (
	"chirpy/api"
	"chirpy/internal/database"
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

const (
	user   = "juanj"
	dbName = "chirps"
	host   = "localhost"
	port   = "5432"
)

func main() {
	dbg := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()

	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading ,env ", err)
	}

	serverMux := http.NewServeMux()
	server := &http.Server{
		Addr:    ":8080",
		Handler: serverMux,
	}

	db, err := database.NewDb(user, dbName, os.Getenv("DB_PASSWORD"), host, port)

	if err != nil {
		log.Fatal(err)
	} else {
		log.Print("DATABASE CONECTION STABLISHED")
	}

	if *dbg {
		err := db.ResetDatabase()
		if err != nil {
			log.Print(err.Error())
		}
	}

	jwt := &api.JwtService{
		JwtSecret: os.Getenv("JWT_SECRET"),
	}
	config := &api.ApiConfig{
		Db:  db,
		Jwt: jwt,
	}
	chirpsController := api.ChirpsService{
		ApiConfig: config,
	}
	UserService := api.UserService{
		ApiConfig: config,
	}

	serverMux.Handle("GET /app/", config.MetricsSiteIncrement(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	serverMux.Handle("GET /app/assets/", http.StripPrefix("/app/assets/", http.FileServer(http.Dir("./assets"))))

	serverMux.HandleFunc("GET /api/healthz", config.Healtz)
	serverMux.HandleFunc("GET /api/metrics", config.WriteServerHits)
	serverMux.HandleFunc("GET /api/reset", config.Reset)
	serverMux.HandleFunc("GET /admin/metrics", config.TemplateMetrics)

	serverMux.HandleFunc("GET /api/chirps/{id}", chirpsController.GetChirp)
	serverMux.HandleFunc("GET /api/chirps", chirpsController.GetChirps)
	serverMux.Handle("POST /api/chirps", jwt.ValidateJWT(http.HandlerFunc(chirpsController.CreateChirp)))
	serverMux.Handle("DELETE /api/chirps/{chirpID}", jwt.ValidateJWT(http.HandlerFunc(chirpsController.DeleteChirp)))

	serverMux.HandleFunc("POST /api/users", UserService.CreateUser)
	serverMux.HandleFunc("POST /api/login", UserService.LogInUser)
	serverMux.Handle("PUT /api/users", jwt.ValidateJWT(http.HandlerFunc(UserService.UpdateUser)))
	serverMux.HandleFunc("POST /api/refresh", UserService.RefreshToken)
	serverMux.HandleFunc("POST /api/revoke", UserService.RevokeRefreshToken)
	serverMux.HandleFunc("POST /api/polka/webhooks", UserService.UpgradeUserMembership)

	err = server.ListenAndServe()
	if err != nil {
		log.Print(err.Error())
	}

}

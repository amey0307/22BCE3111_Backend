package main

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"go_backend_legalForce/auth"
	"go_backend_legalForce/database"
	"go_backend_legalForce/fileupload"
	"go_backend_legalForce/middleware"

	_ "github.com/lib/pq"
)

var db *sql.DB
var redisClient *redis.Client

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	database.InitDB()
	log.Println("Database initialized")
	log.Println("Starting server...")

	router := mux.NewRouter()

	// Public endpoints
	router.HandleFunc("/", test).Methods("GET")
	router.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		auth.RegisterUser(w, r)
	}).Methods("POST")
	router.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		auth.LoginUser(w, r)
	}).Methods("POST")

	router.HandleFunc("/getusers", func(w http.ResponseWriter, r *http.Request) {
		auth.GetUsers(w, r)
	}).Methods("GET")

	router.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		fileupload.SearchFiles(w, r, db, redisClient)
	}).Methods("GET")

	// Serve static files from the uploads directory
	router.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads/"))))
	auth_route := router.PathPrefix("/").Subrouter()
	auth_route.Use(middleware.AuthMiddleware)

	auth_route.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		fileupload.UploadFile(w, r, db, redisClient)
	}).Methods("POST")

	auth_route.HandleFunc("/files", func(w http.ResponseWriter, r *http.Request) {
		fileupload.RetrieveFiles(w, r, db, redisClient)
	}).Methods("GET")

	auth_route.HandleFunc("/share/{file_id:[0-9]+}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		fileID, _ := strconv.Atoi(vars["file_id"])
		fileupload.ShareFile(w, r, db, redisClient, fileID)
	}).Methods("GET")

	log.Println("Server started at :8080")
	http.ListenAndServe(":8080", router)
}

func test(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("API is working!"))
}

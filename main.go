package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"database/sql"
	_ "github.com/lib/pq"
)

type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
}

type File struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	FileName   string    `json:"file_name"`
	UploadDate time.Time `json:"upload_date"`
	Size       int64     `json:"size"`
	S3URL      string    `json:"s3_url"`
}

var db *sql.DB

func initDB() {
	var err error
	connStr := os.Getenv("DB_CONNECTION_STRING")
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	
	log.Println("Connected to the Neon database")
}

func main() {
	err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }
    
	initDB()
	log.Println("Database initialized")
	log.Println("Starting server...")

	router := mux.NewRouter()

	// Public endpoints
	router.HandleFunc("/", test).Methods("GET")

	// Start the server
	log.Println("Server started at :8080")
	http.ListenAndServe(":8080", router)

}

func test(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("API is working!"))
}


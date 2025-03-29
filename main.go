package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"

	"go_backend_legalForce/fileupload"
	"go_backend_legalForce/middleware"
	"go_backend_legalForce/redisconnection"

	"database/sql"

	_ "github.com/lib/pq"
)

type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
}

var db *sql.DB
var redisClient *redis.Client
var wg sync.WaitGroup

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

	redisClient = redisconnection.NewRedisClient()
	log.Println("Connected to the Neon database and Redis")
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

	// Start the background job for file cleanup
	wg.Add(1)
	go func() {
		defer wg.Done()
		startFileCleanup(db)
	}()

	// Public endpoints
	router.HandleFunc("/", test).Methods("GET")
	router.HandleFunc("/register", RegisterUser).Methods("POST")
	router.HandleFunc("/login", LoginUser).Methods("POST")

	// Protected endpoint
	router.Handle("/protected", middleware.AuthMiddleware(http.HandlerFunc(ProtectedEndpoint))).Methods("GET")

	// Serve static files from the uploads directory
	router.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads/"))))

	// File upload endpoint
	router.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		fileupload.UploadFile(w, r, db, redisClient) // Pass redisClient to UploadFile function
	}).Methods("POST")

	// Retrieve files endpoint
	router.HandleFunc("/files", func(w http.ResponseWriter, r *http.Request) {
		fileupload.RetrieveFiles(w, r, db, redisClient) // Pass redisClient to RetrieveFiles function
	}).Methods("GET")

	// Share file endpoint
	router.HandleFunc("/share/{file_id:[0-9]+}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		fileID, _ := strconv.Atoi(vars["file_id"])
		fileupload.ShareFile(w, r, db, redisClient, fileID) // Pass redisClient to ShareFile function
	}).Methods("GET")

	// Search files endpoint
	router.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		fileupload.SearchFiles(w, r, db, redisClient) // Pass redisClient to SearchFiles function
	}).Methods("GET")

	// Start the server
	log.Println("Server started at :8080")
	http.ListenAndServe(":8080", router)

	// Wait for the background job to finish (it won't, but this is for completeness)
	wg.Wait()
}

func test(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("API is working!"))
}

func ProtectedEndpoint(w http.ResponseWriter, r *http.Request) {
	// This endpoint is protected and can only be accessed with a valid JWT

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Welcome to the protected endpoint!"))
}

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}
	user.Password = string(hashedPassword)

	// Save user to the database
	query := `INSERT INTO users (email, password) VALUES ($1, $2) RETURNING id`
	err = db.QueryRow(query, user.Email, user.Password).Scan(&user.ID)
	if err != nil {
		http.Error(w, "Failed to save user", http.StatusInternalServerError)
		return
	}

	// Respond with success message
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})
}

func LoginUser(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Fetch user from the database
	var storedUser User
	query := `SELECT id, email, password FROM users WHERE email = $1`
	err = db.QueryRow(query, user.Email).Scan(&storedUser.ID, &storedUser.Email, &storedUser.Password)
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	// Compare passwords
	err = bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(user.Password))
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": storedUser.Email,
		"exp":   time.Now().Add(time.Hour * 72).Unix(),
	})
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Respond with the token
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

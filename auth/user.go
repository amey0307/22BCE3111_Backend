package auth

import (
	"database/sql"
	"encoding/json"
	"go_backend_legalForce/models"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis/v8"
	"golang.org/x/crypto/bcrypt"

	_ "github.com/lib/pq"
)

var db *sql.DB
var redisClient *redis.Client
var wg sync.WaitGroup

func LoginUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	var storedUser models.User
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

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
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

func GetUsers(w http.ResponseWriter, r *http.Request) {
	print("GetUsers")
	rows, err := db.Query("SELECT id, email FROM users")
	if err != nil {
		http.Error(w, "Failed to retrieve users: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Email); err != nil {
			http.Error(w, "Failed to scan user: "+err.Error(), http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}

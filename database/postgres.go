package database

import (
	"database/sql"
	"go_backend_legalForce/redisconnection"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
)

var db *sql.DB
var redisClient *redis.Client

func InitDB() {
	var err error
	connStr := os.Getenv("DB_CONNECTION_STRING")
	if connStr == "" {
		log.Fatal("DB_CONNECTION_STRING is not set in the environment")
	}

	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to open database connection: %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	redisClient = redisconnection.NewRedisClient()
	if redisClient == nil {
		log.Fatal("Failed to initialize Redis client")
	}

	log.Println("Connected to the Neon database and Redis")
}

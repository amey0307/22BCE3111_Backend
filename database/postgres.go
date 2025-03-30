package database

import (
	"database/sql"
	"go_backend_legalForce/redisconnection"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

var DB *sql.DB
var RedisClient *redis.Client

func InitDB() {
	var err error
	err = godotenv.Load()
	if err != nil {
		log.Println("Warning: Error loading .env file:", err)
	}

	connStr := os.Getenv("DB_CONNECTION_STRING")

	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to open database connection: %v", err)
	}

	err = DB.Ping()
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	RedisClient = redisconnection.NewRedisClient()
	if RedisClient == nil {
		log.Fatal("Failed to initialize Redis client")
	}

	log.Println("Connected to the postgres database and Redis")
}

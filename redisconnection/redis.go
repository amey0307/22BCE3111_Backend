package redisconnection

import (
	"context"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

// Create a context for Redis operations
var ctx = context.Background()

// NewRedisClient creates and returns a new Redis client
func NewRedisClient() *redis.Client {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	// Replace with your Redis Cloud connection details
	redisClient := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_URL"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	// Test the connection
	pong, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}
	log.Printf("Connected to Redis: %s", pong)

	return redisClient
}

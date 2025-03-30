package fileupload

import (
	"database/sql"
	"encoding/json"
	"go_backend_legalForce/models"
	"log"
	"net/http"
	"strconv"

	"github.com/go-redis/redis/v8"
)

func RetrieveFiles(w http.ResponseWriter, r *http.Request, db *sql.DB, redisClient *redis.Client) {
	userID := 3                                      // for my own account -> ameytrips0307@gmail It will be dynamic
	cacheKey := "user_files:" + strconv.Itoa(userID) // Define it here, outside the if block

	// Check Redis cache first
	if redisClient != nil {
		cachedFiles, err := redisClient.Get(ctx, cacheKey).Result()
		if err == nil && redisClient != nil {
			// log.Print("Cache hit", cachedFiles)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(cachedFiles))
			return
		}
	}

	// If cache miss, retrieve from database
	rows, err := db.Query("SELECT id, file_name, upload_date, size, local_path, file_type, s3_url, description, is_shared, expiration_date FROM files WHERE user_id = $1", userID)
	if err != nil {
		http.Error(w, "Failed to retrieve files", http.StatusInternalServerError)
		log.Print(err)
		return
	}
	defer rows.Close()

	var files []models.File
	for rows.Next() {
		var file models.File
		if err := rows.Scan(&file.ID, &file.FileName, &file.UploadDate, &file.Size, &file.LocalPath,
			&file.FileType, &file.S3URL, &file.Description, &file.IsShared, &file.Expiration); err != nil {
			log.Printf("Error scanning row: %v", err)
			http.Error(w, "Failed to scan file", http.StatusInternalServerError)
			return
		}
		files = append(files, file)
	}

	// Cache the result in Redis if redisClient is not nil
	if redisClient != nil {
		cachedData, _ := json.Marshal(files)
		redisClient.Set(ctx, cacheKey, cachedData, 0)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(files)
}

func ShareFile(w http.ResponseWriter, r *http.Request, db *sql.DB, redisClient *redis.Client, fileID int) {
	var err error

	cacheKey := "shared_file:" + strconv.Itoa(fileID)

	if redisClient != nil {
		cachedURL, err := redisClient.Get(ctx, cacheKey).Result()
		if err == nil {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(cachedURL))
			return
		}
	}

	var file models.File
	err = db.QueryRow("SELECT file_name, local_path, is_shared FROM files WHERE id = $1", fileID).Scan(&file.FileName, &file.LocalPath, &file.IsShared)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Construct the public URL -> this is the demo
	publicURL := "http://localhost:8080/" + file.LocalPath

	// Add a nil check before using redisClient
	if redisClient != nil {
		redisClient.Set(ctx, cacheKey, publicURL, 3600) // Cache for 1 hour
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(publicURL))
}

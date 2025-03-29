package fileupload

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

var ctx = context.Background()

// File model
type File struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	FileName   string    `json:"file_name"`
	UploadDate time.Time `json:"upload_date"`
	Size       int64     `json:"size"`
	LocalPath  string    `json:"local_path"` // Path for local storage
}

// UploadFile handles the file upload
func UploadFile(w http.ResponseWriter, r *http.Request, db *sql.DB, redisClient *redis.Client) {

	tx, err := db.Begin() // Start a new transaction
	if err != nil {
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}

	// Get the file from the request
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Unable to get file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Generate a unique file name
	fileName := uuid.New().String() + filepath.Ext(fileHeader.Filename) // Keep the original file extension

	// Save the file locally and get the public URL
	publicURL, err := saveFileLocally(file, fileName)
	if err != nil {
		http.Error(w, "Failed to save file locally", http.StatusInternalServerError)
		return
	}

	// Save file metadata to the database
	fileMetadata := File{
		UserID:     1, // Replace with actual user ID from context
		FileName:   fileName,
		UploadDate: time.Now(),
		Size:       fileHeader.Size,
		LocalPath:  publicURL, // Store the public URL instead of local path
	}

	// Insert file metadata into the database
	query := `INSERT INTO files (user_id, file_name, upload_date, size, local_path) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	err = db.QueryRow(query, fileMetadata.UserID, fileMetadata.FileName, fileMetadata.UploadDate, fileMetadata.Size, fileMetadata.LocalPath).Scan(&fileMetadata.ID)

	if err != nil {
		tx.Rollback() // Rollback the transaction on error
		http.Error(w, "Failed to save file metadata", http.StatusInternalServerError)
		return
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	// Cache the file metadata in Redis
	cachedData, _ := json.Marshal(fileMetadata)
	redisClient.Set(ctx, "file_metadata:"+strconv.Itoa(fileMetadata.ID), cachedData, 0) // Cache with no expiration

	// Respond with the public URL
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"file_url": publicURL})
}

// saveFileLocally saves the uploaded file to the local filesystem
func saveFileLocally(file multipart.File, fileName string) (string, error) {
	// Create a directory for uploads if it doesn't exist
	uploadDir := "uploads"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return "", err
	}

	// Create the file on the local filesystem
	localFilePath := filepath.Join(uploadDir, fileName)
	outFile, err := os.Create(localFilePath)
	if err != nil {
		return "", err
	}
	defer outFile.Close()

	// Copy the uploaded file to the local file
	if _, err := io.Copy(outFile, file); err != nil {
		return "", err
	}

	// Construct the public URL
	publicURL := "http://localhost:8080/" + uploadDir + "/" + fileName

	return publicURL, nil
}

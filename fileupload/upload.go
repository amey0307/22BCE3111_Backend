package fileupload

import (
	"context"
	"database/sql"
	"encoding/json"
	"go_backend_legalForce/models"
	"io"
	"log"
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

// UploadFile handles the file upload
func UploadFile(w http.ResponseWriter, r *http.Request, db *sql.DB, redisClient *redis.Client) {
	err := r.ParseMultipartForm(10 << 20) // 10 MB maximum file size
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
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

	// Prepare file metadata
	// Retrieve the user ID from the context or session
	userID := 14 // Replace with actual user ID from context

	// Check if the user exists in the users table
	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", userID).Scan(&exists)
	if err != nil {
		http.Error(w, "Failed to check user existence", http.StatusInternalServerError)
		return
	}

	if !exists {
		http.Error(w, "User  not found", http.StatusNotFound)
		return
	}

	// Prepare file metadata
	fileMetadata := models.File{
		UserID:      userID,
		FileName:    fileName,
		UploadDate:  time.Now(),
		Size:        fileHeader.Size,
		LocalPath:   publicURL,                         // Store the public URL instead of local path
		FileType:    filepath.Ext(fileHeader.Filename), // Get the file type from the file extension
		S3URL:       "",                                // Leave empty since we're not using S3
		Description: "",                                // Leave empty for now
		IsShared:    false,                             // Default to false
		Expiration:  time.Time{},                       // No expiration set
	}

	// Insert file metadata into the database
	query := `INSERT INTO files (user_id, file_name, upload_date, size, local_path, file_type, s3_url, description, is_shared, expiration_date) 
          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id`
	err = db.QueryRow(query, fileMetadata.UserID, fileMetadata.FileName, fileMetadata.UploadDate, fileMetadata.Size,
		fileMetadata.LocalPath, fileMetadata.FileType, fileMetadata.S3URL, fileMetadata.Description,
		fileMetadata.IsShared, fileMetadata.Expiration).Scan(&fileMetadata.ID)

	if err != nil {
		log.Printf("Failed to save file metadata: %v", err, fileMetadata) // Log the actual error
		http.Error(w, "Failed to save file metadata", http.StatusInternalServerError)
		return
	}

	//clear cache
	// redisClient.Del(ctx, "user_files:14")

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

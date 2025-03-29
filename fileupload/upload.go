package fileupload

import (
	"database/sql"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

type File struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	FileName   string    `json:"file_name"`
	UploadDate time.Time `json:"upload_date"`
	Size       int64     `json:"size"`
	LocalPath  string    `json:"local_path"`
}

func UploadFile(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Unable to get file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileName := uuid.New().String() + filepath.Ext(fileHeader.Filename)

	resultChan := make(chan string)
	var wg sync.WaitGroup

	// Start a goroutine to handle the file upload
	wg.Add(1)
	go func() {
		defer wg.Done()
		publicURL, err := saveFileLocally(file, fileName)
		if err != nil {
			resultChan <- "Failed to save file locally"
			return
		}
		resultChan <- publicURL
	}()

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	publicURL := <-resultChan

	if publicURL == "Failed to save file locally" {
		http.Error(w, publicURL, http.StatusInternalServerError)
		return
	}

	fileMetadata := File{
		UserID:     1,
		FileName:   fileName,
		UploadDate: time.Now(),
		Size:       fileHeader.Size,
		LocalPath:  publicURL,
	}

	query := `INSERT INTO files (user_id, file_name, upload_date, size, local_path) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	err = db.QueryRow(query, fileMetadata.UserID, fileMetadata.FileName, fileMetadata.UploadDate, fileMetadata.Size, fileMetadata.LocalPath).Scan(&fileMetadata.ID)
	if err != nil {
		http.Error(w, "Failed to save file metadata", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"file_url": publicURL})
}

func saveFileLocally(file multipart.File, fileName string) (string, error) {
	uploadDir := "uploads"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return "", err
	}

	localFilePath := filepath.Join(uploadDir, fileName)
	outFile, err := os.Create(localFilePath)
	if err != nil {
		return "", err
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, file); err != nil {
		return "", err
	}

	publicURL := "http://localhost:8080/" + uploadDir + "/" + fileName

	return publicURL, nil
}

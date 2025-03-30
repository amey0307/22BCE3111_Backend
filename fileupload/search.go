package fileupload

import (
	"database/sql"
	"encoding/json"
	"go_backend_legalForce/models"
	"net/http"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

// SearchFiles handles the search functionality for files based on metadata
func SearchFiles(w http.ResponseWriter, r *http.Request, db *sql.DB, redisClient *redis.Client) {
	userID := 3 // Replace with actual user ID from context this is for ameytrips0307@gmail.com

	name := r.URL.Query().Get("name")
	uploadDate := r.URL.Query().Get("upload_date")
	fileType := r.URL.Query().Get("file_type")

	query := "SELECT id, file_name, upload_date, size, local_path FROM files WHERE user_id = $1"
	args := []interface{}{userID}

	if name != "" {
		query += " AND file_name ILIKE $" + strconv.Itoa(len(args)+1) // ILIKE for case-insensitive search
		args = append(args, "%"+name+"%")
	}
	if uploadDate != "" {
		date, err := time.Parse("2006-01-02", uploadDate)
		if err == nil {
			query += " AND upload_date::date = $" + strconv.Itoa(len(args)+1)
			args = append(args, date)
		}
	}
	if fileType != "" {
		query += " AND file_name ILIKE $" + strconv.Itoa(len(args)+1)
		args = append(args, "%."+fileType)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		http.Error(w, "Failed to retrieve files", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var files []models.File
	for rows.Next() {
		var file models.File
		if err := rows.Scan(&file.ID, &file.FileName, &file.UploadDate, &file.Size, &file.LocalPath); err != nil {
			http.Error(w, "Failed to scan file", http.StatusInternalServerError)
			return
		}
		files = append(files, file)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(files)
}

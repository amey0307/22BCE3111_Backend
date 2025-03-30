package job

import (
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
)

func StartFileCleanup(db *sql.DB) {
	for {
		// Check for expired files
		deleteExpiredFiles(db)

		// Wait for a specified interval before checking again (e.g., every hour)
		time.Sleep(1 * time.Hour)
	}
}

func deleteExpiredFiles(db *sql.DB) {
	// Query to find expired files
	query := `SELECT id, local_path FROM files WHERE expiration_date < NOW()`
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("Error querying expired files: %v", err)
		return
	}
	defer rows.Close()

	// if no rows are returned, there are no expired files
	if !rows.Next() {
		log.Println("No expired files found")
		return
	}

	var fileID int
	var localPath string
	for rows.Next() {
		if err := rows.Scan(&fileID, &localPath); err != nil {
			log.Printf("Error scanning expired file: %v", err)
			continue
		}

		// Delete the file from local storage
		if err := os.Remove(localPath); err != nil {
			log.Printf("Error deleting file %s: %v", localPath, err)
			continue
		}

		// Remove the corresponding metadata from the database
		_, err := db.Exec("DELETE FROM files WHERE id = $1", fileID)
		if err != nil {
			log.Printf("Error deleting metadata for file ID %d: %v", fileID, err)
		} else {
			log.Printf("Deleted expired file %s and its metadata from the database", localPath)
		}
	}
}

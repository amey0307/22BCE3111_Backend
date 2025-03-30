package job

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// StartFileCleanup initiates a background routine that periodically checks for and deletes expired files
func StartFileCleanup(db *sql.DB) {
	// Check if db is nil before proceeding
	if db == nil {
		log.Println("ERROR: Database connection is nil, file cleanup service cannot start")
		return
	}

	log.Println("Starting file cleanup service")

	for {
		// Check for expired files
		filesDeleted, err := deleteExpiredFiles(db)
		if err != nil {
			log.Printf("Error during file cleanup: %v", err)
		} else {
			log.Printf("File cleanup completed: %d expired files removed", filesDeleted)
		}

		// Wait for a specified interval before checking again (e.g., every hour)
		log.Println("Next file cleanup scheduled in 1 hour")
		time.Sleep(1 * time.Hour)
	}
}

// deleteExpiredFiles removes files whose expiration date has passed
// Returns the number of files deleted and any error encountered
func deleteExpiredFiles(db *sql.DB) (int, error) {
	// Query to find expired files
	query := `SELECT id, local_path FROM files WHERE expiration_date < NOW() AND expiration_date != '0001-01-01 00:00:00'`
	rows, err := db.Query(query)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	// Count of successfully deleted files
	deletedFiles := 0

	// Loop through all expired files
	for rows.Next() {
		var fileID int
		var localPath string

		if err := rows.Scan(&fileID, &localPath); err != nil {
			log.Printf("Error scanning expired file row: %v", err)
			continue
		}

		log.Printf("Processing expired file: %s (ID: %d)", localPath, fileID)

		// Make sure the path is properly formatted
		// If the localPath is a URL, extract just the file path portion
		if filepath.IsAbs(localPath) == false && localPath != "" {
			// Handle case where localPath is stored as a URL
			if filepath.HasPrefix(localPath, "http://") || filepath.HasPrefix(localPath, "https://") {
				localPath = filepath.Join("uploads", filepath.Base(localPath))
				log.Printf("Converted URL to local path: %s", localPath)
			} else {
				// Ensure it's a full path
				localPath = filepath.Join("uploads", localPath)
			}
		}

		// Check if file exists before attempting deletion
		if _, err := os.Stat(localPath); os.IsNotExist(err) {
			log.Printf("File %s does not exist on disk, but will remove database entry", localPath)
		} else {
			// Delete the file from local storage
			if err := os.Remove(localPath); err != nil {
				log.Printf("Error deleting file %s: %v", localPath, err)
				// Continue with metadata deletion even if file deletion fails
			} else {
				log.Printf("Successfully deleted file from disk: %s", localPath)
			}
		}

		// Remove the corresponding metadata from the database
		result, err := db.Exec("DELETE FROM files WHERE id = $1", fileID)
		if err != nil {
			log.Printf("Error deleting metadata for file ID %d: %v", fileID, err)
			continue
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			deletedFiles++
			log.Printf("Successfully deleted metadata for file ID %d", fileID)
		} else {
			log.Printf("No rows affected when deleting metadata for file ID %d", fileID)
		}
	}

	// Check for any errors that occurred during iteration
	if err = rows.Err(); err != nil {
		log.Printf("Error iterating over rows: %v", err)
		return deletedFiles, err
	}

	return deletedFiles, nil
}

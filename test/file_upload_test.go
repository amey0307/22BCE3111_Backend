package test

import (
	"bytes"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go_backend_legalForce/fileupload"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-redis/redis/v8"
)

func TestMain(m *testing.M) {
	log.Println("=== Starting test suite ===")
	cleanUploadFolder()

	log.Println("=== Running tests ===")
	result := m.Run()

	log.Println("=== Cleaning up after tests ===")
	cleanUploadFolder()

	log.Println("=== Test suite completed ===")
	os.Exit(result)
}

func cleanUploadFolder() {
	uploadDir := filepath.Join("..", "uploads")
	log.Printf("Cleaning upload directory: %s", uploadDir)

	if _, err := os.Stat(uploadDir); !os.IsNotExist(err) {
		err := os.RemoveAll(uploadDir)
		if err != nil {
			log.Printf("WARNING: Failed to clean up upload folder: %v", err)
		} else {
			log.Printf("Successfully cleaned up upload folder: %s", uploadDir)
		}
	} else {
		log.Printf("Upload directory does not exist, nothing to clean")
	}
}

func TestFileUpload(t *testing.T) {
	log.Println("--- Starting TestFileUpload ---")
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()
	log.Println("Created mock database")

	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	log.Println("Created mock Redis client")

	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test-upload.txt")
	content := []byte("This is a test file for upload testing")
	log.Printf("Created temporary file: %s", tempFile)

	if err := os.WriteFile(tempFile, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)
	log.Println("Preparing multipart form request")

	file, err := os.Open(tempFile)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer file.Close()

	part, err := writer.CreateFormFile("file", filepath.Base(tempFile))
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		t.Fatalf("Failed to copy file content: %v", err)
	}

	err = writer.Close()
	if err != nil {
		t.Fatalf("Failed to close multipart writer: %v", err)
	}
	log.Println("Multipart form request prepared successfully")

	req := httptest.NewRequest("POST", "/upload", &requestBody)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	log.Printf("Created test request with Content-Type: %s", writer.FormDataContentType())

	rr := httptest.NewRecorder()

	log.Println("Setting up SQL mock expectations")

	userID := 3
	mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM users WHERE id = \\$1\\)").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	log.Printf("Added mock for user existence check (userID: %d)", userID)

	mock.ExpectQuery("INSERT INTO files .* RETURNING id").
		WithArgs(
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	log.Println("Added mock for file insertion query")

	log.Println("Calling UploadFile handler")
	fileupload.UploadFile(rr, req, db, redisClient)
	log.Printf("Handler finished with status code: %d", rr.Code)

	if status := rr.Code; status != http.StatusOK {
		log.Printf("ERROR: Handler returned wrong status code: got %v, want %v", status, http.StatusOK)
		t.Errorf("Handler returned wrong status code: got %v, want %v", status, http.StatusOK)
	} else {
		log.Printf("Handler returned correct status code: %v", status)
	}

	responseBody := rr.Body.String()
	if !strings.Contains(responseBody, "file_url") {
		log.Printf("ERROR: Response doesn't contain file_url: %s", responseBody)
		t.Errorf("Response doesn't contain file_url: %s", responseBody)
	} else {
		log.Printf("Response contains file_url: %s", responseBody)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		log.Printf("ERROR: Unfulfilled expectations: %s", err)
		t.Errorf("Unfulfilled expectations: %s", err)
	} else {
		log.Println("All database expectations were met")
	}

	log.Println("--- TestFileUpload completed ---")
}

func TestFileUploadNoFile(t *testing.T) {
	log.Println("--- Starting TestFileUploadNoFile ---")
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()
	log.Println("Created mock database")

	req := httptest.NewRequest("POST", "/upload", nil)
	req.Header.Set("Content-Type", "multipart/form-data")
	log.Println("Created test request with no file")

	rr := httptest.NewRecorder()

	log.Println("Calling UploadFile handler with no file")
	fileupload.UploadFile(rr, req, db, nil)
	log.Printf("Handler finished with status code: %d", rr.Code)

	if status := rr.Code; status != http.StatusBadRequest {
		log.Printf("ERROR: Handler returned wrong status code: got %v, want %v", status, http.StatusBadRequest)
		t.Errorf("Handler returned wrong status code: got %v, want %v", status, http.StatusBadRequest)
	} else {
		log.Printf("Handler returned correct status code: %v", status)
	}

	log.Println("--- TestFileUploadNoFile completed ---")
}

func TestFileUploadUserNotFound(t *testing.T) {
	log.Println("--- Starting TestFileUploadUserNotFound ---")
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()
	log.Println("Created mock database")

	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test-upload.txt")
	content := []byte("This is a test file for upload testing")
	log.Printf("Created temporary file: %s", tempFile)

	if err := os.WriteFile(tempFile, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)
	log.Println("Preparing multipart form request")

	file, err := os.Open(tempFile)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer file.Close()

	part, err := writer.CreateFormFile("file", filepath.Base(tempFile))
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		t.Fatalf("Failed to copy file content: %v", err)
	}

	err = writer.Close()
	if err != nil {
		t.Fatalf("Failed to close multipart writer: %v", err)
	}
	log.Println("Multipart form request prepared successfully")

	req := httptest.NewRequest("POST", "/upload", &requestBody)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	log.Printf("Created test request with Content-Type: %s", writer.FormDataContentType())

	rr := httptest.NewRecorder()

	userID := 3
	log.Println("Setting up SQL mock expectations - user does not exist")
	mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM users WHERE id = \\$1\\)").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	log.Printf("Added mock for user existence check (userID: %d, exists: false)", userID)

	log.Println("Calling UploadFile handler")
	fileupload.UploadFile(rr, req, db, nil)
	log.Printf("Handler finished with status code: %d", rr.Code)

	if status := rr.Code; status != http.StatusNotFound {
		log.Printf("ERROR: Handler returned wrong status code: got %v, want %v", status, http.StatusNotFound)
		t.Errorf("Handler returned wrong status code: got %v, want %v", status, http.StatusNotFound)
	} else {
		log.Printf("Handler returned correct status code: %v", status)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		log.Printf("ERROR: Unfulfilled expectations: %s", err)
		t.Errorf("Unfulfilled expectations: %s", err)
	} else {
		log.Println("All database expectations were met")
	}

	log.Println("--- TestFileUploadUserNotFound completed ---")
}

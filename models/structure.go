package models

import "time"

type File struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	FileName    string    `json:"file_name"`
	UploadDate  time.Time `json:"upload_date"`
	Size        int64     `json:"size"`
	LocalPath   string    `json:"local_path"`      // Path for local storage
	FileType    string    `json:"file_type"`       // Type of the file (e.g., jpg, pdf)
	S3URL       string    `json:"s3_url"`          // URL if stored in S3 (leave empty if not used)
	Description string    `json:"description"`     // Description of the file
	IsShared    bool      `json:"is_shared"`       // Indicates if the file is shared
	Expiration  time.Time `json:"expiration_date"` // Expiration date for shared files
}

type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
}

package models

import "time"

type File struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	FileName    string    `json:"file_name"`
	UploadDate  time.Time `json:"upload_date"`
	Size        int64     `json:"size"`
	LocalPath   string    `json:"local_path"`      
	FileType    string    `json:"file_type"`      
	S3URL       string    `json:"s3_url"`          
	Description string    `json:"description"`     
	IsShared    bool      `json:"is_shared"`       
	Expiration  time.Time `json:"expiration_date"` 
}

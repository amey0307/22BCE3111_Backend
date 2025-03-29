# File Management API

This is a backend API for managing file uploads, retrieval, and sharing functionalities. The API is built using Go and utilizes PostgreSQL for data storage and Redis for caching.

## Table of Contents

- [Base URL](#base-url)
- [Authentication](#authentication)
- [Endpoints](#endpoints)
  - [Upload File](#upload-file)
  - [Retrieve Files](#retrieve-files)
  - [Search Files](#search-files)
  - [Share File](#share-file)
- [Error Handling](#error-handling)
- [Technologies Used](#technologies-used)
- [Installation](#installation)
- [Usage](#usage)

## Base URL
http://localhost:8080


## Authentication

- The API does not currently implement authentication. User IDs are hardcoded for demonstration purposes. In a production environment, consider implementing JWT or OAuth for secure access.

## Endpoints

### 1. Upload File

- **Endpoint**: `/upload`
- **Method**: `POST`
- **Description**: Uploads a file and saves its metadata in the database.
- **Request**:
  - **Form Data**:
    - `file`: The file to be uploaded.
- **Response**:
  - **Status**: `200 OK`
  - **Body**: 
    ```json
    {
      "file_url": "http://localhost:8080/uploads/test-file.txt"
    }
    ```

### 2. Retrieve Files

- **Endpoint**: `/files`
- **Method**: `GET`
- **Description**: Retrieves a list of files uploaded by a specific user.
- **Request**: 
  - **Query Parameters**:
    - `user_id`: The ID of the user whose files are to be retrieved (currently hardcoded).
- **Response**:
  - **Status**: `200 OK`
  - **Body**: 
    ```json
    [
      {
        "id": 1,
        "user_id": 1,
        "file_name": "test-file.txt",
        "upload_date": "2023-01-01T00:00:00Z",
        "size": 1024,
        "local_path": "http://localhost:8080/uploads/test-file.txt",
        "file_type": "txt",
        "s3_url": "",
        "description": "",
        "is_shared": false,
        "expiration_date": null
      }
    ]
    ```

### 3. Search Files

- **Endpoint**: `/search`
- **Method**: `GET`
- **Description**: Searches for files based on metadata.
- **Request**: 
  - **Query Parameters**:
    - `name`: (optional) The name of the file to search for.
    - `upload_date`: (optional) The date the file was uploaded (format: `YYYY-MM-DD`).
    - `file_type`: (optional) The type of the file (e.g., `txt`, `pdf`).
- **Response**:
  - **Status**: `200 OK`
  - **Body**: 
    ```json
    [
      {
        "id": 1,
        "user_id": 1,
        "file_name": "test-file.txt",
        "upload_date": "2023-01-01T00:00:00Z",
        "size": 1024,
        "local_path": "http://localhost:8080/uploads/test-file.txt",
        "file_type": "txt",
        "s3_url": "",
        "description": "",
        "is_shared": false,
        "expiration_date": null
      }
    ]
    ```

### 4. Share File

- **Endpoint**: `/share/{file_id}`
- **Method**: `GET`
- **Description**: Shares a file via a public link.
- **Request**: 
  - **Path Parameter**:
    - `file_id`: The ID of the file to be shared.
- **Response**:
  - **Status**: `200 OK`
  - **Body**: 
    ```json
    "http://localhost:8080/uploads/test-file.txt"
    ```

## Error Handling

- The API returns appropriate HTTP status codes for different error scenarios:
  - `400 Bad Request`: Invalid request parameters.
  - `404 Not Found`: Resource not found.
  - `500 Internal Server Error`: An error occurred on the server.

## Technologies Used

- Go
- PostgreSQL
- Redis
- Gorilla Mux (for routing)
- Go-Redis (for Redis client)

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/your-repo.git
   cd your-repo


### Customization

- **Base URL**: Update the base URL if your application is hosted elsewhere.
- **Authentication**: If you implement authentication, update the documentation accordingly.
- **
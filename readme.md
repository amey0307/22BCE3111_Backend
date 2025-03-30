# File Management API Documentation

# File Management API

A RESTful API for managing file uploads, retrieval, and sharing functionalities. Built with Go, PostgreSQL, and Redis.

## Table of Contents

- [Base URL](https://www.notion.so/File-Management-API-Documentation-1b867f328534801c8e84d5dc59994d51?pvs=21)
- [Authentication](https://www.notion.so/File-Management-API-Documentation-1b867f328534801c8e84d5dc59994d51?pvs=21)
- [Endpoints](https://www.notion.so/File-Management-API-Documentation-1b867f328534801c8e84d5dc59994d51?pvs=21)
    - [User Management](https://www.notion.so/File-Management-API-Documentation-1b867f328534801c8e84d5dc59994d51?pvs=21)
        - [Register User](https://www.notion.so/File-Management-API-Documentation-1b867f328534801c8e84d5dc59994d51?pvs=21)
        - [Login User](https://www.notion.so/File-Management-API-Documentation-1b867f328534801c8e84d5dc59994d51?pvs=21)
        - [Get Users](https://www.notion.so/File-Management-API-Documentation-1b867f328534801c8e84d5dc59994d51?pvs=21)
    - [File Management](https://www.notion.so/File-Management-API-Documentation-1b867f328534801c8e84d5dc59994d51?pvs=21)
        - [Upload File](https://www.notion.so/File-Management-API-Documentation-1b867f328534801c8e84d5dc59994d51?pvs=21)
        - [Retrieve Files](https://www.notion.so/File-Management-API-Documentation-1b867f328534801c8e84d5dc59994d51?pvs=21)
        - [Share File](https://www.notion.so/File-Management-API-Documentation-1b867f328534801c8e84d5dc59994d51?pvs=21)
        - [Search Files](https://www.notion.so/File-Management-API-Documentation-1b867f328534801c8e84d5dc59994d51?pvs=21)
- [Error Handling](https://www.notion.so/File-Management-API-Documentation-1b867f328534801c8e84d5dc59994d51?pvs=21)
- [Environment Variables](https://www.notion.so/File-Management-API-Documentation-1b867f328534801c8e84d5dc59994d51?pvs=21)
- [Development Setup](https://www.notion.so/File-Management-API-Documentation-1b867f328534801c8e84d5dc59994d51?pvs=21)

## Base URL http://localhost:8080

## Authentication

The API uses JWT (JSON Web Token) for authentication.

- To access protected endpoints, include the token in the `Authorization` header of your requests.
- Example: `Authorization: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ...`

## Endpoints

### User Management

### Register User

Creates a new user account.

- **URL**: `/register`
- **Method**: `POST`
- **Authentication**: None
- **Request Body**:
    
    ```json
    {
      "email": "user@example.com",
      "password": "secure_password"
    }
    ```
    

**Response**

:

- **Status**: `200 OK`
- **Body**:
    
    ```jsx
    {
    
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ..."
    
    }
    ```
    

### Get Users

Retrieves a list of all registered users.

- **URL**: `/getusers`
- **Method**: `GET`
- **Authentication**: None
- **Response**:
    - **Status**: `200 OK`
    - **Body:**
    
    ```jsx
    [
      {
        "id": 1,
        "email": "user1@example.com"
      },
      {
        "id": 2,
        "email": "user2@example.com"
      }
    ]
    ```
    

### **File Management**

### Upload File

Uploads a file and saves its metadata in the database.

- **URL**: `/upload`
- **Method**: `POST`
- **Authentication**: Required
- **Request**:
    - **Content-Type**: multipart/form-data
    - **Form Fields**:
        - file: The file to be uploaded (required)
- **Response**:
    - **Status**: `200 OK`
    - **Body**:

```jsx
{
  "file_url": "http://localhost:8080/uploads/b78c0293-c4a3-4587-b9e1-b83df4e6d496.txt"
}
```

### Retrieve Files

Gets a list of all files uploaded by the authenticated user.

- **URL**: `/files`
- **Method**: `GET`
- **Authentication**: Required
- **Response**:
    - **Status**: `200 OK`
    - **Body**:

```jsx
[
  {
    "id": 1,
    "user_id": 3,
    "file_name": "b78c0293-c4a3-4587-b9e1-b83df4e6d496.txt",
    "upload_date": "2025-03-30T15:10:25Z",
    "size": 36,
    "local_path": "http://localhost:8080/uploads/b78c0293-c4a3-4587-b9e1-b83df4e6d496.txt",
    "file_type": ".txt",
    "s3_url": "",
    "description": "",
    "is_shared": false,
    "expiration_date": "0001-01-01T00:00:00Z"
  }
]
```

### Share File

Creates a shareable link for a specific file.

- **URL**: `/share/{file_id}`
- **Method**: `GET`
- **Authentication**: Required
- **URL Parameters**:
    - `file_id`: ID of the file to share
- **Response**:
    - **Status**: `200 OK`
    - **Body**

```jsx
http://localhost:8080/uploads/b78c0293-c4a3-4587-b9e1-b83df4e6d496.txt
```

### Search Files

Searches for files based on metadata filters.

- **URL**: `/search`
- **Method**: `GET`
- **Authentication**: None
- **Query Parameters**:
    - name: (Optional) Filter by file name
    - `upload_date`: (Optional) Filter by upload date (format: `YYYY-MM-DD`)
    - `file_type`: (Optional) Filter by file extension (e.g. `txt`, `pdf`)
- **Response**:
    - **Status**: `200 OK`
    - **Body**

```jsx
[
  {
    "id": 1,
    "file_name": "b78c0293-c4a3-4587-b9e1-b83df4e6d496.txt",
    "upload_date": "2025-03-30T15:10:25Z",
    "size": 36,
    "local_path": "http://localhost:8080/uploads/b78c0293-c4a3-4587-b9e1-b83df4e6d496.txt"
  }
]
```

## **Error Handling**

The API returns appropriate HTTP status codes along with error messages:

- `400 Bad Request`: Invalid input or request format
- `401 Unauthorized`: Authentication required or invalid credentials
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Server-side error

## **Environment Variables**

The application requires the following environment variables:

- `DB_CONNECTION_STRING`: PostgreSQL connection string
- `JWT_SECRET_KEY`: Secret key for JWT token generation
- `REDIS_URL`: Redis server URL
- `REDIS_PASSWORD`: Redis server password

## **Development Setup**

1. Clone the repository
2. Create a .env file with the required environment variables
3. Run the application:
    
    go run cmd/main.go
    

```jsx
go run cmd/main.go
```

1. go run cmd/main.go

```jsx
go run cmd/main.go
```
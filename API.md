# API Documentation

## Base URL
```
http://localhost:3000
```

## Authentication
Most endpoints require a Bearer token in the Authorization header:
```
Authorization: Bearer <jwt-token>
```

## Response Format
- Success responses: JSON with relevant data
- Error responses: JSON Problem Details (RFC 7807)

## Endpoints

### Health Check
```http
GET /health
```
**Response:**
```json
{
  "status": "OK",
  "timestamp": "2024-01-01T00:00:00.000Z"
}
```

### Authentication

#### Register User
```http
POST /api/auth/register
Content-Type: application/json
```
**Body:**
```json
{
  "username": "string (3-50 chars, alphanumeric + underscore)",
  "email": "string (valid email)",
  "password": "string (min 6 chars, must contain upper, lower, number)"
}
```

#### Login User
```http
POST /api/auth/login
Content-Type: application/json
```
**Body:**
```json
{
  "username": "string",
  "password": "string"
}
```

### Books

#### List All Books
```http
GET /api/books?page=1&limit=10
```

#### Search Books
```http
GET /api/books/search?title=<title>&author=<author>&genre=<genre>&availability=<available|unavailable>&sort=<field:order>&page=<page>&limit=<limit>
```

**Query Parameters:**
- `title`: Partial title search
- `author`: Partial author search  
- `genre`: Partial genre search
- `availability`: Filter by availability
- `sort`: Sort format `field:order` where field is `title|author|genre|created_at` and order is `asc|desc`
- `page`: Page number (default: 1)
- `limit`: Items per page (default: 10, max: 100)

#### Get Book by ID
```http
GET /api/books/:id
```

#### Get Available Genres
```http
GET /api/books/meta/genres
```

#### Get Available Authors
```http
GET /api/books/meta/authors
```

### Loans (Requires Authentication)

#### Borrow Book
```http
POST /api/loans/borrow
Authorization: Bearer <token>
Content-Type: application/json
```
**Body:**
```json
{
  "bookId": "string",
  "daysToReturn": "integer (1-30, default: 14)"
}
```
**Note:** This operation is idempotent - borrowing the same book twice returns the existing loan.

#### Return Book
```http
PATCH /api/loans/return/:loanId
Authorization: Bearer <token>
```

#### Get Borrowing History
```http
GET /api/loans/history?status=<borrowed|returned|overdue>&page=<page>&limit=<limit>
Authorization: Bearer <token>
```

#### Get Active Loans
```http
GET /api/loans/active
Authorization: Bearer <token>
```

#### Get Loan Statistics
```http
GET /api/loans/stats
Authorization: Bearer <token>
```

## HTTP Status Codes

### Success Codes
- `200 OK`: Request successful
- `201 Created`: Resource created successfully

### Client Error Codes
- `400 Bad Request`: Invalid request data
- `401 Unauthorized`: Authentication required
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource not found
- `409 Conflict`: Resource conflict (e.g., user already exists)
- `429 Too Many Requests`: Rate limit exceeded

### Server Error Codes
- `500 Internal Server Error`: Unexpected server error

## Error Response Format

All errors follow JSON Problem Details (RFC 7807):

```json
{
  "type": "https://httpstatuses.com/400",
  "title": "Validation Error",
  "status": 400,
  "detail": "Detailed error description",
  "instance": "/api/path",
  "errors": [
    {
      "field": "fieldName",
      "message": "Field-specific error message",
      "value": "invalid-value"
    }
  ]
}
```

## Rate Limiting

- **Limit**: 60 requests per minute per IP address
- **Headers**: Rate limit info in `RateLimit-*` headers
- **Reset**: Window resets every minute

## Data Models

### User
```json
{
  "id": "uuid",
  "username": "string",
  "email": "string",
  "created_at": "datetime",
  "updated_at": "datetime"
}
```

### Book
```json
{
  "id": "string",
  "title": "string",
  "author": "string", 
  "genre": "string",
  "isbn": "string",
  "total_copies": "integer",
  "available_copies": "integer",
  "created_at": "datetime",
  "updated_at": "datetime"
}
```

### Loan
```json
{
  "id": "uuid",
  "user_id": "uuid",
  "book_id": "string",
  "borrowed_at": "datetime",
  "due_date": "datetime",
  "returned_at": "datetime|null",
  "status": "borrowed|returned",
  "created_at": "datetime",
  "updated_at": "datetime",
  "title": "string",
  "author": "string",
  "genre": "string",
  "current_status": "borrowed|returned|overdue"
}
```
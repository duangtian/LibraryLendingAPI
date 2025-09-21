# LibraryLendingAPI

A RESTful API for a library book lending system built with Node.js, Express, and SQLite.

## Features

- **User Authentication**: Register and login with JWT tokens
- **Book Search**: Search books with filtering (title, author, genre, availability), sorting, and pagination
- **Book Borrowing**: Borrow and return books with idempotent operations
- **Borrowing History**: View personal borrowing history and statistics
- **Rate Limiting**: 60 requests per minute per user
- **Error Handling**: JSON Problem Details (RFC 7807) format
- **Docker Support**: Full containerization support

## Architecture

- **HTTP Semantics**: Proper use of GET/POST/PATCH/DELETE with appropriate status codes (2xx/4xx/5xx)
- **Validation**: Comprehensive input validation with detailed error messages
- **Security**: JWT authentication, password hashing, rate limiting, security headers
- **Database**: SQLite with proper indexing and transaction support

## Quick Start

### Using Docker (Recommended)

```bash
# Clone the repository
git clone <repository-url>
cd LibraryLendingAPI

# Start with docker-compose
docker-compose up -d

# The API will be available at http://localhost:3000
```

### Manual Setup

```bash
# Install dependencies
npm install

# Start the server
npm run dev
```

## API Endpoints

### Authentication

#### Register User
```http
POST /api/auth/register
Content-Type: application/json

{
  "username": "johndoe",
  "email": "john@example.com",
  "password": "SecurePass123"
}
```

#### Login User
```http
POST /api/auth/login
Content-Type: application/json

{
  "username": "johndoe",
  "password": "SecurePass123"
}
```

### Books

#### Search Books
```http
GET /api/books/search?title=gatsby&author=fitzgerald&genre=classic&availability=available&sort=title:asc&page=1&limit=10
```

#### Get All Books
```http
GET /api/books?page=1&limit=10
```

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
Authorization: Bearer <jwt-token>
Content-Type: application/json

{
  "bookId": "1",
  "daysToReturn": 14
}
```

#### Return Book
```http
PATCH /api/loans/return/:loanId
Authorization: Bearer <jwt-token>
```

#### Get Borrowing History
```http
GET /api/loans/history?status=borrowed&page=1&limit=10
Authorization: Bearer <jwt-token>
```

#### Get Active Loans
```http
GET /api/loans/active
Authorization: Bearer <jwt-token>
```

#### Get Loan Statistics
```http
GET /api/loans/stats
Authorization: Bearer <jwt-token>
```

## Error Handling

The API uses JSON Problem Details (RFC 7807) format for error responses:

```json
{
  "type": "https://httpstatuses.com/400",
  "title": "Validation Error",
  "status": 400,
  "detail": "Validation failed",
  "instance": "/api/auth/register",
  "errors": [
    {
      "field": "email",
      "message": "Please provide a valid email address",
      "value": "invalid-email"
    }
  ]
}
```

## Rate Limiting

- **Limit**: 60 requests per minute per IP address
- **Headers**: Rate limit information is provided in response headers
- **Error**: Returns 429 status code when limit is exceeded

## Environment Variables

```env
PORT=3000
JWT_SECRET=your-super-secret-jwt-key-change-in-production
NODE_ENV=development
DB_PATH=./database.sqlite
RATE_LIMIT_WINDOW_MS=60000
RATE_LIMIT_MAX=60
```

## Database Schema

### Users
- `id` (PRIMARY KEY)
- `username` (UNIQUE)
- `email` (UNIQUE)
- `password_hash`
- `created_at`
- `updated_at`

### Books
- `id` (PRIMARY KEY)
- `title`
- `author`
- `genre`
- `isbn` (UNIQUE)
- `total_copies`
- `available_copies`
- `created_at`
- `updated_at`

### Loans
- `id` (PRIMARY KEY)
- `user_id` (FOREIGN KEY)
- `book_id` (FOREIGN KEY)
- `borrowed_at`
- `due_date`
- `returned_at`
- `status` (borrowed/returned)
- `created_at`
- `updated_at`

## Sample Data

The API comes pre-seeded with 10 sample books including classics like "The Great Gatsby", "1984", "Harry Potter", etc.

## Security Features

- **JWT Authentication**: Secure token-based authentication
- **Password Hashing**: bcrypt with salt rounds
- **Rate Limiting**: Prevents abuse
- **Security Headers**: Helmet.js for security headers
- **Input Validation**: Comprehensive validation with express-validator
- **SQL Injection Protection**: Parameterized queries

## Development

```bash
# Start development server with auto-reload
npm run dev

# Start production server
npm start
```

## Health Check

The API provides a health check endpoint:

```http
GET /health
```

Response:
```json
{
  "status": "OK",
  "timestamp": "2024-01-01T00:00:00.000Z"
}
```

## License

MIT
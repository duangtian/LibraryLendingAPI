package models

import "time"

type User struct {
	ID        int64     `db:"id" json:"id"`
	Email     string    `db:"email" json:"email"`
	PasswordHash string `db:"password_hash" json:"-"`
	Role      string    `db:"role" json:"role"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type Book struct {
	ID             int64  `db:"id" json:"id"`
	ISBN           string `db:"isbn" json:"isbn"`
	Title          string `db:"title" json:"title"`
	Author         string `db:"author" json:"author"`
	Genre          string `db:"genre" json:"genre"`
	TotalCopies    int    `db:"total_copies" json:"total_copies"`
	AvailableCopies int   `db:"available_copies" json:"available_copies"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
}

type Loan struct {
	ID         int64     `db:"id" json:"id"`
	UserID     int64     `db:"user_id" json:"user_id"`
	BookID     int64     `db:"book_id" json:"book_id"`
	BorrowedAt time.Time `db:"borrowed_at" json:"borrowed_at"`
	DueAt      time.Time `db:"due_at" json:"due_at"`
	ReturnedAt *time.Time `db:"returned_at" json:"returned_at,omitempty"`
}

type IdempotencyKey struct {
	Key       string    `db:"key" json:"key"`
	ResponseBody []byte `db:"response_body" json:"-"`
	StatusCode int      `db:"status_code" json:"-"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

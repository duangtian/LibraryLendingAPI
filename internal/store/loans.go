package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/example/librarylendingapi/internal/models"
)

var (
	// ErrBookUnavailable indicates there are no copies left to borrow.
	ErrBookUnavailable = errors.New("book unavailable")
	// ErrLoanNotFound indicates the loan (or referenced book) was not found.
	ErrLoanNotFound = errors.New("loan not found")
)

// LoanStore provides operations for creating, returning and listing loans.
type LoanStore struct{ db *pgxpool.Pool }

func NewLoanStore(db *pgxpool.Pool) *LoanStore { return &LoanStore{db: db} }

// Create creates a new loan for the given user & book if a copy is available.
func (s *LoanStore) Create(ctx context.Context, userID, bookID int64, duration time.Duration) (*models.Loan, error) {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() {
		if tx != nil {
			tx.Rollback(ctx)
		}
	}()

	// Lock the book row to ensure availability check & decrement are atomic.
	var available int
	if err := tx.QueryRow(ctx, `SELECT available_copies FROM books WHERE id=$1 FOR UPDATE`, bookID).Scan(&available); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrLoanNotFound
		}
		return nil, err
	}
	if available <= 0 {
		return nil, ErrBookUnavailable
	}

	if _, err := tx.Exec(ctx, `UPDATE books SET available_copies = available_copies - 1 WHERE id=$1`, bookID); err != nil {
		return nil, err
	}
	dueAt := time.Now().Add(duration)
	row := tx.QueryRow(ctx, `INSERT INTO loans (user_id, book_id, due_at) VALUES ($1,$2,$3) RETURNING id,user_id,book_id,borrowed_at,due_at,returned_at`, userID, bookID, dueAt)
	var l models.Loan
	if err := row.Scan(&l.ID, &l.UserID, &l.BookID, &l.BorrowedAt, &l.DueAt, &l.ReturnedAt); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	tx = nil
	return &l, nil
}

// Return marks the loan as returned (idempotent) and increments book availability.
func (s *LoanStore) Return(ctx context.Context, loanID int64) (*models.Loan, error) {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() {
		if tx != nil {
			tx.Rollback(ctx)
		}
	}()

	var l models.Loan
	if err := tx.QueryRow(ctx, `SELECT id,user_id,book_id,borrowed_at,due_at,returned_at FROM loans WHERE id=$1 FOR UPDATE`, loanID).Scan(&l.ID, &l.UserID, &l.BookID, &l.BorrowedAt, &l.DueAt, &l.ReturnedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrLoanNotFound
		}
		return nil, err
	}
	// If already returned, treat as idempotent success.
	if l.ReturnedAt != nil {
		return &l, nil
	}

	now := time.Now()
	if _, err := tx.Exec(ctx, `UPDATE loans SET returned_at=$2 WHERE id=$1`, loanID, now); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `UPDATE books SET available_copies = available_copies + 1 WHERE id=$1`, l.BookID); err != nil {
		return nil, err
	}
	l.ReturnedAt = &now
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	tx = nil
	return &l, nil
}

// ListByUser lists a user's loans with pagination.
func (s *LoanStore) ListByUser(ctx context.Context, userID int64, limit, offset int) ([]models.Loan, int, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := s.db.Query(ctx, `SELECT id,user_id,book_id,borrowed_at,due_at,returned_at FROM loans WHERE user_id=$1 ORDER BY borrowed_at DESC LIMIT $2 OFFSET $3`, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []models.Loan
	for rows.Next() {
		var l models.Loan
		if err := rows.Scan(&l.ID, &l.UserID, &l.BookID, &l.BorrowedAt, &l.DueAt, &l.ReturnedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, l)
	}
	var total int
	if err := s.db.QueryRow(ctx, `SELECT count(*) FROM loans WHERE user_id=$1`, userID).Scan(&total); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

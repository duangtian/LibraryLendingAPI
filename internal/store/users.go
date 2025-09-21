package store

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/example/librarylendingapi/internal/models"
)

var ErrUserNotFound = errors.New("user not found")
var ErrEmailExists = errors.New("email already exists")

type UserStore struct{ db *pgxpool.Pool }

func NewUserStore(db *pgxpool.Pool) *UserStore { return &UserStore{db: db} }

func (s *UserStore) Create(ctx context.Context, email, password, role string) (*models.User, error) {
	pwHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	row := s.db.QueryRow(ctx, `INSERT INTO users (email, password_hash, role) VALUES ($1,$2,$3) RETURNING id, email, password_hash, role, created_at`, email, string(pwHash), role)
	var u models.User
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			// Unique violation: 23505. Either check constraint name or code.
			if pgErr.Code == "23505" && pgErr.ConstraintName == "users_email_key" {
				return nil, ErrEmailExists
			}
		}
		return nil, err
	}
	return &u, nil
}

func (s *UserStore) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	row := s.db.QueryRow(ctx, `SELECT id, email, password_hash, role, created_at FROM users WHERE email=$1`, email)
	var u models.User
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (s *UserStore) VerifyPassword(u *models.User, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
}

func (s *UserStore) CreateIfNotExistsOAuth(ctx context.Context, email string) (*models.User, error) { // placeholder for future
	id := uuid.New().String()
	_ = id
	return nil, errors.New("not implemented")
}

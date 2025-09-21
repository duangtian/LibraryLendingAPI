package store

import (
    "context"
    "errors"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

var ErrIdempotencyKeyNotFound = errors.New("idempotency key not found")

type IdempotencyStore struct { db *pgxpool.Pool }

func NewIdempotencyStore(db *pgxpool.Pool) *IdempotencyStore { return &IdempotencyStore{db: db} }

type IdempotencyRecord struct {
    Key string
    StatusCode int
    Body []byte
}

func (s *IdempotencyStore) Get(ctx context.Context, key string) (*IdempotencyRecord, error) {
    row := s.db.QueryRow(ctx, `SELECT key, status_code, response_body FROM idempotency_keys WHERE key=$1`, key)
    var rec IdempotencyRecord
    if err := row.Scan(&rec.Key, &rec.StatusCode, &rec.Body); err != nil {
        if errors.Is(err, pgx.ErrNoRows) { return nil, ErrIdempotencyKeyNotFound }
        return nil, err
    }
    return &rec, nil
}

func (s *IdempotencyStore) Save(ctx context.Context, key string, status int, body []byte) error {
    _, err := s.db.Exec(ctx, `INSERT INTO idempotency_keys (key, status_code, response_body) VALUES ($1,$2,$3) ON CONFLICT (key) DO NOTHING`, key, status, body)
    return err
}

package store

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/example/librarylendingapi/internal/models"
)

type BookFilter struct {
	Title string
	Author string
	Genre string
	Available *bool
	Sort string // title,-title,author,-author
	Limit int
	Offset int
}

type BookStore struct { db *pgxpool.Pool }

func NewBookStore(db *pgxpool.Pool) *BookStore { return &BookStore{db: db} }

func (s *BookStore) Search(ctx context.Context, f BookFilter) ([]models.Book, int, error) {
	var where []string
	var args []any
	idx := 1
	if f.Title != "" { where = append(where, fmt.Sprintf("title ILIKE '%%' || $%d || '%%'", idx)); args = append(args, f.Title); idx++ }
	if f.Author != "" { where = append(where, fmt.Sprintf("author ILIKE '%%' || $%d || '%%'", idx)); args = append(args, f.Author); idx++ }
	if f.Genre != "" { where = append(where, fmt.Sprintf("genre = $%d", idx)); args = append(args, f.Genre); idx++ }
	if f.Available != nil { if *f.Available { where = append(where, "available_copies > 0") } else { where = append(where, "available_copies = 0") } }
	clause := ""
	if len(where) > 0 { clause = "WHERE " + strings.Join(where, " AND ") }

	order := "title ASC"
	switch f.Sort {
	case "title": order = "title ASC"
	case "-title": order = "title DESC"
	case "author": order = "author ASC"
	case "-author": order = "author DESC"
	case "genre": order = "genre ASC"
	case "-genre": order = "genre DESC"
	}
	limit := 20
	if f.Limit > 0 && f.Limit <= 100 { limit = f.Limit }
	offset := f.Offset

	query := fmt.Sprintf("SELECT id,isbn,title,author,genre,total_copies,available_copies,created_at FROM books %s ORDER BY %s LIMIT %d OFFSET %d", clause, order, limit, offset)
	rows, err := s.db.Query(ctx, query, args...)
	if err != nil { return nil, 0, err }
	defer rows.Close()
	var list []models.Book
	for rows.Next() {
		var b models.Book
		if err := rows.Scan(&b.ID,&b.ISBN,&b.Title,&b.Author,&b.Genre,&b.TotalCopies,&b.AvailableCopies,&b.CreatedAt); err != nil { return nil,0,err }
		list = append(list, b)
	}
	countQuery := fmt.Sprintf("SELECT count(*) FROM books %s", clause)
	var total int
	if err := s.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil { return nil,0,err }
	return list, total, nil
}

package http

import (
	"bytes"
	"encoding/json"
	stdhttp "net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/example/librarylendingapi/internal/auth"
	"github.com/example/librarylendingapi/internal/store"
	"github.com/example/librarylendingapi/internal/validation"
)

type Server struct {
	Router *chi.Mux
	Books  *store.BookStore
	Users  *store.UserStore
	JWT    *auth.JWTManager
	Loans  *store.LoanStore
	Idem   *store.IdempotencyStore
}

func NewServer(books *store.BookStore, users *store.UserStore, jwt *auth.JWTManager, loans *store.LoanStore, idem *store.IdempotencyStore) *Server {
	r := chi.NewRouter()
	s := &Server{Router: r, Books: books, Users: users, JWT: jwt, Loans: loans, Idem: idem}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.Router.Get("/v1/healthz", func(w stdhttp.ResponseWriter, r *stdhttp.Request) { w.Write([]byte("ok")) })
	s.Router.Get("/v1/books", s.handleSearchBooks)
	s.Router.Post("/v1/auth/register", s.handleRegister)
	s.Router.Post("/v1/auth/login", s.handleLogin)
	s.Router.Post("/v1/loans", s.handleCreateLoan)
	s.Router.Patch("/v1/loans/{id}/return", s.handleReturnLoan)
	s.Router.Get("/v1/me/loans", s.handleListMyLoans)
}

func (s *Server) handleSearchBooks(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	q := r.URL.Query()
	var f store.BookFilter
	f.Title = q.Get("title")
	f.Author = q.Get("author")
	f.Genre = q.Get("genre")
	if av := q.Get("available"); av != "" {
		if av == "true" {
			b := true
			f.Available = &b
		} else if av == "false" {
			b := false
			f.Available = &b
		}
	}
	f.Sort = q.Get("sort")
	if l := q.Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil {
			f.Limit = v
		}
	}
	if o := q.Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil {
			f.Offset = v
		}
	}
	books, total, err := s.Books.Search(r.Context(), f)
	if err != nil {
		stdhttp.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{"))
	w.Write([]byte("\"total\":" + strconv.Itoa(total) + ","))
	w.Write([]byte("\"items\":"))
	enc := json.NewEncoder(w)
	enc.Encode(books)
	// json.Encoder adds newline; wrap JSON manually minimal for now
	w.Write([]byte("}"))
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type authResponse struct {
	Token string `json:"token"`
}

func (s *Server) handleRegister(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		validation.Write(w, 400, "bad request", "invalid json", nil)
		return
	}
	var verr validation.Errors
	if fe := validation.NotBlank("email", req.Email); fe != nil {
		verr = append(verr, *fe)
	}
	if fe := validation.MinLen("password", req.Password, 6); fe != nil {
		verr = append(verr, *fe)
	}
	if len(verr) > 0 {
		validation.Write(w, 400, "validation error", "invalid input", verr.ToInvalidParams())
		return
	}
	user, err := s.Users.Create(r.Context(), req.Email, req.Password, "member")
	if err != nil {
		if err == store.ErrEmailExists {
			validation.Write(w, 409, "conflict", "email exists", nil)
			return
		}
		validation.Write(w, 500, "internal error", err.Error(), nil)
		return
	}
	token, err := s.JWT.Generate(user.ID, user.Role)
	if err != nil {
		validation.Write(w, 500, "internal error", err.Error(), nil)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(authResponse{Token: token})
}

func (s *Server) handleLogin(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		validation.Write(w, 400, "bad request", "invalid json", nil)
		return
	}
	user, err := s.Users.GetByEmail(r.Context(), req.Email)
	if err != nil {
		validation.Write(w, 401, "unauthorized", "invalid credentials", nil)
		return
	}
	if err := s.Users.VerifyPassword(user, req.Password); err != nil {
		validation.Write(w, 401, "unauthorized", "invalid credentials", nil)
		return
	}
	token, err := s.JWT.Generate(user.ID, user.Role)
	if err != nil {
		validation.Write(w, 500, "internal error", err.Error(), nil)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(authResponse{Token: token})
}

type createLoanRequest struct {
	BookID int64 `json:"book_id"`
}

func (s *Server) handleCreateLoan(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	claims := r.Context().Value("userClaims")
	if claims == nil {
		validation.Write(w, 401, "unauthorized", "login required", nil)
		return
	}
	var req createLoanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		validation.Write(w, 400, "bad request", "invalid json", nil)
		return
	}
	if req.BookID <= 0 {
		validation.Write(w, 400, "validation error", "book_id invalid", nil)
		return
	}
	idemKey := r.Header.Get("Idempotency-Key")
	if idemKey == "" {
		validation.Write(w, 400, "validation error", "missing Idempotency-Key header", nil)
		return
	}
	if rec, err := s.Idem.Get(r.Context(), idemKey); err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(rec.StatusCode)
		w.Write(rec.Body)
		return
	}
	loan, err := s.Loans.Create(r.Context(), claims.(*auth.Claims).UserID, req.BookID, 14*24*time.Hour)
	if err != nil {
		validation.Write(w, 400, "cannot create loan", err.Error(), nil)
		return
	}
	b := bytes.Buffer{}
	json.NewEncoder(&b).Encode(loan)
	s.Idem.Save(r.Context(), idemKey, 201, b.Bytes())
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(b.Bytes())
}

func (s *Server) handleReturnLoan(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	loan, err := s.Loans.Return(r.Context(), id)
	if err != nil {
		validation.Write(w, 400, "cannot return loan", err.Error(), nil)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(loan)
}

func (s *Server) handleListMyLoans(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	claims := r.Context().Value("userClaims")
	if claims == nil {
		validation.Write(w, 401, "unauthorized", "login required", nil)
		return
	}
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))
	list, total, err := s.Loans.ListByUser(r.Context(), claims.(*auth.Claims).UserID, limit, offset)
	if err != nil {
		validation.Write(w, 500, "internal error", err.Error(), nil)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"total": total, "items": list})
}

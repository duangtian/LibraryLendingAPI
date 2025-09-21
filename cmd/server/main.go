package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"

	"github.com/example/librarylendingapi/internal/auth"
	api "github.com/example/librarylendingapi/internal/http"
	"github.com/example/librarylendingapi/internal/store"
)

func main() {
	port := getEnv("PORT", "8080")

	dsn := getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/library?sslmode=disable")
	ctx := context.Background()
	db, err := store.Connect(ctx, dsn)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer db.Close()
	if err := db.RunMigrations(ctx, "migrations"); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	bookStore := store.NewBookStore(db.Pool)
	userStore := store.NewUserStore(db.Pool)
	jwtSecret := getEnv("JWT_SECRET", "devsecret")
	jwtMgr := auth.NewJWTManager(jwtSecret, "library-api", 24*time.Hour)
	loanStore := store.NewLoanStore(db.Pool)
	idemStore := store.NewIdempotencyStore(db.Pool)
	srvAPI := api.NewServer(bookStore, userStore, jwtMgr, loanStore, idemStore)

	r := chi.NewRouter()
	// Global coarse rate limit (remaining per-user limiter present in middleware logic if extended)
	r.Use(httprate.LimitByIP(60, 1*time.Minute))
	r.Mount("/", srvAPI.Router)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
		BaseContext:  func(l net.Listener) context.Context { return context.Background() },
	}

	log.Printf("server listening on :%s", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %v", err)
	}
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

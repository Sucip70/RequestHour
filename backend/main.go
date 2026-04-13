package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"requesthour/backend/internal/handler"
	"requesthour/backend/internal/repository"
	"requesthour/backend/internal/service"
)

func main() {
	_ = godotenv.Load()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set (use your Supabase connection string from Project Settings → Database)")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("failed to create pool: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("connected but ping failed: %v", err)
	}

	sessionRepo := repository.NewSessionRepository(pool)
	sessionSvc := service.NewSessionService(sessionRepo)
	sessionHandler := handler.NewSessionHandler(sessionSvc)

	addr := ":" + getenv("PORT", "8080")
	mux := http.NewServeMux()
	mux.HandleFunc("POST /session", sessionHandler.CreateSession)

	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

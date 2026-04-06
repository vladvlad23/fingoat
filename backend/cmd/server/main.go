package main

import (
	"context"
	"log"
	"net/http"

	"github.com/fingoat/api/internal/config"
	"github.com/fingoat/api/internal/db"
	"github.com/fingoat/api/internal/handler"
	mw "github.com/fingoat/api/internal/middleware"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("db ping: %v", err)
	}

	m, err := migrate.New("file://internal/db/migration", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("migrate init: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("migrate up: %v", err)
	}

	store := &handler.QueriesStore{Queries: db.New(pool)}

	authH := handler.NewAuthHandler(store, cfg)
	userH := handler.NewUserHandler(store, cfg)
	goalH := handler.NewGoalHandler(store, cfg)
	txH := handler.NewTransactionHandler(store, pool, cfg)

	r := chi.NewRouter()
	r.Use(chimw.Recoverer)
	r.Use(mw.Logger)

	r.Post("/api/auth/register", authH.Register)
	r.Post("/api/auth/login", authH.Login)

	r.Group(func(r chi.Router) {
		r.Use(mw.Auth(cfg.JWTSecret))

		r.Get("/api/me", userH.GetMe)
		r.Patch("/api/me", userH.UpdateMe)

		r.Get("/api/goals", goalH.ListGoals)
		r.Post("/api/goals", goalH.CreateGoal)
		r.Get("/api/goals/{id}", goalH.GetGoal)
		r.Put("/api/goals/{id}", goalH.UpdateGoal)
		r.Delete("/api/goals/{id}", goalH.DeleteGoal)

		r.Get("/api/transactions", txH.ListTransactions)
		r.Post("/api/transactions", txH.CreateTransaction)
		r.Get("/api/transactions/{id}", txH.GetTransaction)
		r.Put("/api/transactions/{id}", txH.UpdateTransaction)
		r.Delete("/api/transactions/{id}", txH.DeleteTransaction)
	})

	log.Printf("server starting on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("server: %v", err)
	}
}

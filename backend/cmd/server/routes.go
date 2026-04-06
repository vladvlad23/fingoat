package main

import (
	"net/http"

	"github.com/fingoat/api/internal/config"
	"github.com/fingoat/api/internal/db"
	"github.com/fingoat/api/internal/handler"
	mw "github.com/fingoat/api/internal/middleware"
	"github.com/fingoat/api/internal/stores"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

func registerRoutes(pool *pgxpool.Pool, cfg *config.Config) http.Handler {
	store := &stores.QueriesStore{Queries: db.New(pool)}

	authH := handler.NewAuthHandler(store, cfg)
	userH := handler.NewUserHandler(store, cfg)
	goalH := handler.NewGoalHandler(store, cfg)
	txH := handler.NewTransactionHandler(store, pool, cfg)
	summaryH := handler.NewSummaryHandler(store, cfg)

	r := chi.NewRouter()
	r.Use(chimw.Recoverer)
	r.Use(mw.Logger)

	r.Post("/api/auth/register", authH.Register)
	r.Post("/api/auth/login", authH.Login)
	r.Post("/api/auth/refresh", authH.Refresh)
	r.Post("/api/auth/logout", authH.Logout)

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

		r.Get("/api/summary", summaryH.GetSummary)
	})

	return r
}

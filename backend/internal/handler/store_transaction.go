package handler

import (
	"context"

	"github.com/fingoat/api/internal/db"
)

// TransactionStore defines transaction-related database operations.
type TransactionStore interface {
	CreateTransaction(ctx context.Context, arg db.CreateTransactionParams) (db.Transaction, error)
	ListTransactions(ctx context.Context, arg db.ListTransactionsParams) ([]db.Transaction, error)
	GetTransactionByID(ctx context.Context, id int64) (db.Transaction, error)
	UpdateTransaction(ctx context.Context, arg db.UpdateTransactionParams) (db.Transaction, error)
	DeleteTransaction(ctx context.Context, id int64) error
	GetTransactionSummary(ctx context.Context, arg db.GetTransactionSummaryParams) ([]db.TransactionSummaryRow, error)
}

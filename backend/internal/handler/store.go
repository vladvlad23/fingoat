package handler

import (
	"context"

	"github.com/fingoat/api/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// Store defines the database operations used by handlers.
// *db.Queries satisfies this interface.
type Store interface {
	// Users
	CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error)
	GetUserByEmail(ctx context.Context, email string) (db.User, error)
	GetUserByID(ctx context.Context, id int64) (db.User, error)
	UpdateUserIncome(ctx context.Context, id int64, monthlyIncome pgtype.Numeric) (db.User, error)
	UpdateUserSettings(ctx context.Context, arg db.UpdateUserSettingsParams) (db.User, error)

	// Goals
	CreateGoal(ctx context.Context, arg db.CreateGoalParams) (db.Goal, error)
	ListGoalsByUser(ctx context.Context, userID int64) ([]db.Goal, error)
	GetGoalByID(ctx context.Context, id int64) (db.Goal, error)
	UpdateGoal(ctx context.Context, arg db.UpdateGoalParams) (db.Goal, error)
	DeleteGoal(ctx context.Context, id int64) error
	UpdateGoalCurrentAmount(ctx context.Context, id int64, delta pgtype.Numeric) (db.Goal, error)

	// Transactions
	CreateTransaction(ctx context.Context, arg db.CreateTransactionParams) (db.Transaction, error)
	ListTransactions(ctx context.Context, arg db.ListTransactionsParams) ([]db.Transaction, error)
	GetTransactionByID(ctx context.Context, id int64) (db.Transaction, error)
	UpdateTransaction(ctx context.Context, arg db.UpdateTransactionParams) (db.Transaction, error)
	DeleteTransaction(ctx context.Context, id int64) error
	GetTransactionSummary(ctx context.Context, arg db.GetTransactionSummaryParams) ([]db.TransactionSummaryRow, error)

	// WithTx returns a Store that runs queries within the given transaction.
	WithTx(tx pgx.Tx) Store
}

// TxBeginner abstracts the ability to begin a database transaction.
// *pgxpool.Pool satisfies this interface.
type TxBeginner interface {
	BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error)
}

// QueriesStore wraps *db.Queries to satisfy the Store interface.
// The WithTx method on *db.Queries returns *db.Queries, so this wrapper
// adapts it to return Store.
type QueriesStore struct {
	*db.Queries
}

// WithTx returns a new QueriesStore scoped to the given transaction.
func (qs *QueriesStore) WithTx(tx pgx.Tx) Store {
	return &QueriesStore{Queries: qs.Queries.WithTx(tx)}
}

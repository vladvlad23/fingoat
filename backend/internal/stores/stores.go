package stores

import (
	"context"

	"github.com/fingoat/api/internal/db"
	"github.com/jackc/pgx/v5"
)

// AuthStore combines the stores needed by AuthHandler.
type AuthStore interface {
	UserStore
	RefreshTokenStore
}

// SummaryStore combines the stores needed by SummaryHandler.
type SummaryStore interface {
	GoalStore
	TransactionStore
}

// TxableStore combines GoalStore and TransactionStore with transaction support,
// used by TransactionHandler for atomic goal-amount updates.
type TxableStore interface {
	GoalStore
	TransactionStore
	WithTx(tx pgx.Tx) TxableStore
}

// TxBeginner abstracts the ability to begin a database transaction.
// *pgxpool.Pool satisfies this interface.
type TxBeginner interface {
	BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error)
}

// QueriesStore wraps *db.Queries to satisfy all store interfaces.
// The WithTx method on *db.Queries returns *db.Queries, so this wrapper
// adapts it to return TxableStore.
type QueriesStore struct {
	*db.Queries
}

// WithTx returns a new QueriesStore scoped to the given transaction.
func (qs *QueriesStore) WithTx(tx pgx.Tx) TxableStore {
	return &QueriesStore{Queries: qs.Queries.WithTx(tx)}
}

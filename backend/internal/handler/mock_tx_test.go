package handler

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// mockTxBeginner implements TxBeginner for unit tests.
type mockTxBeginner struct {
	beginTxFn func(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error)
}

func (m *mockTxBeginner) BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error) {
	if m.beginTxFn != nil {
		return m.beginTxFn(ctx, opts)
	}
	return &mockTx{}, nil
}

// mockTx implements pgx.Tx for unit tests.
type mockTx struct {
	commitFn   func(ctx context.Context) error
	rollbackFn func(ctx context.Context) error
}

func (m *mockTx) Begin(ctx context.Context) (pgx.Tx, error) {
	return m, nil
}

func (m *mockTx) Commit(ctx context.Context) error {
	if m.commitFn != nil {
		return m.commitFn(ctx)
	}
	return nil
}

func (m *mockTx) Rollback(ctx context.Context) error {
	if m.rollbackFn != nil {
		return m.rollbackFn(ctx)
	}
	return nil
}

func (m *mockTx) CopyFrom(_ context.Context, _ pgx.Identifier, _ []string, _ pgx.CopyFromSource) (int64, error) {
	return 0, nil
}

func (m *mockTx) SendBatch(_ context.Context, _ *pgx.Batch) pgx.BatchResults {
	return nil
}

func (m *mockTx) LargeObjects() pgx.LargeObjects {
	return pgx.LargeObjects{}
}

func (m *mockTx) Prepare(_ context.Context, _ string, _ string) (*pgconn.StatementDescription, error) {
	return nil, nil
}

func (m *mockTx) Exec(_ context.Context, _ string, _ ...interface{}) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}

func (m *mockTx) Query(_ context.Context, _ string, _ ...interface{}) (pgx.Rows, error) {
	return nil, nil
}

func (m *mockTx) QueryRow(_ context.Context, _ string, _ ...interface{}) pgx.Row {
	return nil
}

func (m *mockTx) Conn() *pgx.Conn {
	return nil
}

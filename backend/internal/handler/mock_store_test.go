package handler

import (
	"context"
	"errors"

	"github.com/fingoat/api/internal/db"
	"github.com/fingoat/api/internal/stores"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// errNotFound is a sentinel error used in mock implementations.
var errNotFound = errors.New("not found")

// mockStore is a hand-written mock that implements the Store interface.
// Each method can be overridden by assigning a function to the corresponding field.
type mockStore struct {
	createUserFn                  func(ctx context.Context, arg db.CreateUserParams) (db.User, error)
	getUserByEmailFn              func(ctx context.Context, email string) (db.User, error)
	getUserByIDFn                 func(ctx context.Context, id int64) (db.User, error)
	updateUserIncomeFn            func(ctx context.Context, id int64, income pgtype.Numeric) (db.User, error)
	updateUserSettingsFn          func(ctx context.Context, arg db.UpdateUserSettingsParams) (db.User, error)
	createGoalFn                  func(ctx context.Context, arg db.CreateGoalParams) (db.Goal, error)
	listGoalsByUserFn             func(ctx context.Context, userID int64) ([]db.Goal, error)
	getGoalByIDFn                 func(ctx context.Context, id int64) (db.Goal, error)
	updateGoalFn                  func(ctx context.Context, arg db.UpdateGoalParams) (db.Goal, error)
	deleteGoalFn                  func(ctx context.Context, id int64) error
	updateGoalCurrentAmountFn     func(ctx context.Context, id int64, delta pgtype.Numeric) (db.Goal, error)
	createTransactionFn           func(ctx context.Context, arg db.CreateTransactionParams) (db.Transaction, error)
	listTransactionsFn            func(ctx context.Context, arg db.ListTransactionsParams) ([]db.Transaction, error)
	getTransactionByIDFn          func(ctx context.Context, id int64) (db.Transaction, error)
	updateTransactionFn           func(ctx context.Context, arg db.UpdateTransactionParams) (db.Transaction, error)
	deleteTransactionFn           func(ctx context.Context, id int64) error
	getTransactionSummaryFn       func(ctx context.Context, arg db.GetTransactionSummaryParams) ([]db.TransactionSummaryRow, error)
	createRefreshTokenFn          func(ctx context.Context, arg db.CreateRefreshTokenParams) (db.RefreshToken, error)
	getRefreshTokenByHashFn       func(ctx context.Context, tokenHash string) (db.RefreshToken, error)
	deleteRefreshTokenFn          func(ctx context.Context, id int64) error
	deleteRefreshTokensByUserIDFn func(ctx context.Context, userID int64) error
}

func (m *mockStore) CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error) {
	if m.createUserFn != nil {
		return m.createUserFn(ctx, arg)
	}
	return db.User{}, errors.New("CreateUser not implemented")
}

func (m *mockStore) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	if m.getUserByEmailFn != nil {
		return m.getUserByEmailFn(ctx, email)
	}
	return db.User{}, errNotFound
}

func (m *mockStore) GetUserByID(ctx context.Context, id int64) (db.User, error) {
	if m.getUserByIDFn != nil {
		return m.getUserByIDFn(ctx, id)
	}
	return db.User{}, errNotFound
}

func (m *mockStore) UpdateUserIncome(ctx context.Context, id int64, income pgtype.Numeric) (db.User, error) {
	if m.updateUserIncomeFn != nil {
		return m.updateUserIncomeFn(ctx, id, income)
	}
	return db.User{}, errors.New("UpdateUserIncome not implemented")
}

func (m *mockStore) UpdateUserSettings(ctx context.Context, arg db.UpdateUserSettingsParams) (db.User, error) {
	if m.updateUserSettingsFn != nil {
		return m.updateUserSettingsFn(ctx, arg)
	}
	return db.User{}, errors.New("UpdateUserSettings not implemented")
}

func (m *mockStore) CreateGoal(ctx context.Context, arg db.CreateGoalParams) (db.Goal, error) {
	if m.createGoalFn != nil {
		return m.createGoalFn(ctx, arg)
	}
	return db.Goal{}, errors.New("CreateGoal not implemented")
}

func (m *mockStore) ListGoalsByUser(ctx context.Context, userID int64) ([]db.Goal, error) {
	if m.listGoalsByUserFn != nil {
		return m.listGoalsByUserFn(ctx, userID)
	}
	return nil, errors.New("ListGoalsByUser not implemented")
}

func (m *mockStore) GetGoalByID(ctx context.Context, id int64) (db.Goal, error) {
	if m.getGoalByIDFn != nil {
		return m.getGoalByIDFn(ctx, id)
	}
	return db.Goal{}, errNotFound
}

func (m *mockStore) UpdateGoal(ctx context.Context, arg db.UpdateGoalParams) (db.Goal, error) {
	if m.updateGoalFn != nil {
		return m.updateGoalFn(ctx, arg)
	}
	return db.Goal{}, errors.New("UpdateGoal not implemented")
}

func (m *mockStore) DeleteGoal(ctx context.Context, id int64) error {
	if m.deleteGoalFn != nil {
		return m.deleteGoalFn(ctx, id)
	}
	return errors.New("DeleteGoal not implemented")
}

func (m *mockStore) UpdateGoalCurrentAmount(ctx context.Context, id int64, delta pgtype.Numeric) (db.Goal, error) {
	if m.updateGoalCurrentAmountFn != nil {
		return m.updateGoalCurrentAmountFn(ctx, id, delta)
	}
	return db.Goal{}, errors.New("UpdateGoalCurrentAmount not implemented")
}

func (m *mockStore) CreateTransaction(ctx context.Context, arg db.CreateTransactionParams) (db.Transaction, error) {
	if m.createTransactionFn != nil {
		return m.createTransactionFn(ctx, arg)
	}
	return db.Transaction{}, errors.New("CreateTransaction not implemented")
}

func (m *mockStore) ListTransactions(ctx context.Context, arg db.ListTransactionsParams) ([]db.Transaction, error) {
	if m.listTransactionsFn != nil {
		return m.listTransactionsFn(ctx, arg)
	}
	return nil, errors.New("ListTransactions not implemented")
}

func (m *mockStore) GetTransactionByID(ctx context.Context, id int64) (db.Transaction, error) {
	if m.getTransactionByIDFn != nil {
		return m.getTransactionByIDFn(ctx, id)
	}
	return db.Transaction{}, errNotFound
}

func (m *mockStore) UpdateTransaction(ctx context.Context, arg db.UpdateTransactionParams) (db.Transaction, error) {
	if m.updateTransactionFn != nil {
		return m.updateTransactionFn(ctx, arg)
	}
	return db.Transaction{}, errors.New("UpdateTransaction not implemented")
}

func (m *mockStore) DeleteTransaction(ctx context.Context, id int64) error {
	if m.deleteTransactionFn != nil {
		return m.deleteTransactionFn(ctx, id)
	}
	return errors.New("DeleteTransaction not implemented")
}

func (m *mockStore) GetTransactionSummary(ctx context.Context, arg db.GetTransactionSummaryParams) ([]db.TransactionSummaryRow, error) {
	if m.getTransactionSummaryFn != nil {
		return m.getTransactionSummaryFn(ctx, arg)
	}
	return nil, errors.New("GetTransactionSummary not implemented")
}

func (m *mockStore) CreateRefreshToken(ctx context.Context, arg db.CreateRefreshTokenParams) (db.RefreshToken, error) {
	if m.createRefreshTokenFn != nil {
		return m.createRefreshTokenFn(ctx, arg)
	}
	return db.RefreshToken{}, errors.New("CreateRefreshToken not implemented")
}

func (m *mockStore) GetRefreshTokenByHash(ctx context.Context, tokenHash string) (db.RefreshToken, error) {
	if m.getRefreshTokenByHashFn != nil {
		return m.getRefreshTokenByHashFn(ctx, tokenHash)
	}
	return db.RefreshToken{}, errNotFound
}

func (m *mockStore) DeleteRefreshToken(ctx context.Context, id int64) error {
	if m.deleteRefreshTokenFn != nil {
		return m.deleteRefreshTokenFn(ctx, id)
	}
	return nil
}

func (m *mockStore) DeleteRefreshTokensByUserID(ctx context.Context, userID int64) error {
	if m.deleteRefreshTokensByUserIDFn != nil {
		return m.deleteRefreshTokensByUserIDFn(ctx, userID)
	}
	return nil
}

func (m *mockStore) WithTx(_ pgx.Tx) stores.TxableStore {
	// In tests, WithTx returns the same mock store, since we don't
	// actually use pgx transactions in unit tests.
	return m
}

package db

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type User struct {
	ID            int64          `json:"id"`
	Email         string         `json:"email"`
	PasswordHash  string         `json:"password_hash"`
	MonthlyIncome pgtype.Numeric `json:"monthly_income"`
	PaymentDay    *int32         `json:"payment_day"`
	CreatedAt     time.Time      `json:"created_at"`
}

type Goal struct {
	ID            int64          `json:"id"`
	UserID        int64          `json:"user_id"`
	Title         string         `json:"title"`
	TargetAmount  pgtype.Numeric `json:"target_amount"`
	CurrentAmount pgtype.Numeric `json:"current_amount"`
	Deadline      *pgtype.Date   `json:"deadline"`
	Status        string         `json:"status"`
	CreatedAt     time.Time      `json:"created_at"`
}

type Transaction struct {
	ID        int64          `json:"id"`
	UserID    int64          `json:"user_id"`
	GoalID    *int64         `json:"goal_id"`
	Title     string         `json:"title"`
	Amount    pgtype.Numeric `json:"amount"`
	Type      string         `json:"type"`
	Category  *string        `json:"category"`
	Date      pgtype.Date    `json:"date"`
	CreatedAt time.Time      `json:"created_at"`
}

type RefreshToken struct {
	ID        int64
	UserID    int64
	TokenHash string
	ExpiresAt time.Time
	CreatedAt time.Time
}

package model

import (
	"fmt"
	"time"

	"github.com/fingoat/api/internal/db"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserResponse struct {
	ID            int64     `json:"id"`
	Email         string    `json:"email"`
	MonthlyIncome string    `json:"monthlyIncome"`
	PaymentDay    *int32    `json:"paymentDay,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
}

type GoalResponse struct {
	ID            int64     `json:"id"`
	UserID        int64     `json:"userId"`
	Title         string    `json:"title"`
	TargetAmount  string    `json:"targetAmount"`
	CurrentAmount string    `json:"currentAmount"`
	Deadline      *string   `json:"deadline,omitempty"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"createdAt"`
}

type TransactionResponse struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"userId"`
	GoalID    *int64    `json:"goalId,omitempty"`
	Title     string    `json:"title"`
	Amount    string    `json:"amount"`
	Type      string    `json:"type"`
	Category  *string   `json:"category,omitempty"`
	Date      string    `json:"date"`
	CreatedAt time.Time `json:"createdAt"`
}

func numericToString(n pgtype.Numeric) string {
	if !n.Valid {
		return "0.00"
	}
	v, err := n.Value()
	if err != nil || v == nil {
		return "0.00"
	}
	s, ok := v.(string)
	if !ok {
		return fmt.Sprintf("%v", v)
	}
	return s
}

func UserToResponse(u db.User) UserResponse {
	return UserResponse{
		ID:            u.ID,
		Email:         u.Email,
		MonthlyIncome: numericToString(u.MonthlyIncome),
		PaymentDay:    u.PaymentDay,
		CreatedAt:     u.CreatedAt,
	}
}

func GoalToResponse(g db.Goal) GoalResponse {
	resp := GoalResponse{
		ID:            g.ID,
		UserID:        g.UserID,
		Title:         g.Title,
		TargetAmount:  numericToString(g.TargetAmount),
		CurrentAmount: numericToString(g.CurrentAmount),
		Status:        g.Status,
		CreatedAt:     g.CreatedAt,
	}
	if g.Deadline != nil && g.Deadline.Valid {
		s := g.Deadline.Time.Format("2006-01-02")
		resp.Deadline = &s
	}
	return resp
}

func TransactionToResponse(t db.Transaction) TransactionResponse {
	return TransactionResponse{
		ID:        t.ID,
		UserID:    t.UserID,
		GoalID:    t.GoalID,
		Title:     t.Title,
		Amount:    numericToString(t.Amount),
		Type:      t.Type,
		Category:  t.Category,
		Date:      t.Date.Time.Format("2006-01-02"),
		CreatedAt: t.CreatedAt,
	}
}
